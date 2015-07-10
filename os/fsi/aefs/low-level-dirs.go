package aefs

import (
	"os"
	"sort"
	"strings"
	"time"

	"github.com/pbberlin/tools/logif"

	pth "path"

	ds "appengine/datastore"
)

// Retrieves a directory in one go.
// Also used to check existence; returning ds.ErrNoSuchEntity
func (fs *AeFileSys) dirByPath(path string) (AeDir, error) {

	path = cleanseLeadingSlash(path)
	if path != fs.RootName() && !strings.HasPrefix(path, fs.RootName()) {
		path = fs.RootDir() + path
	}
	// logif.Pf("  %v", path)

	fo := AeDir{}
	fo.fSys = fs

	preciseK := ds.NewKey(fs.c, tdir, path, 0, nil)
	fo.Key = preciseK

	err := fo.MemCacheGet(path)
	if err == nil {
		return fo, nil
	}

	err = ds.Get(fs.c, preciseK, &fo)
	if err == ds.ErrNoSuchEntity {
		// uncomment to see where directories do not exist:
		// logif.Pf("no directory: %-20v ", path)
		// runtimepb.StackTrace(4)
		return fo, err
	} else if err != nil {
		logif.E(err)
	}

	fo.MemCacheSet()
	return fo, err
}

// dirsByPath might not find recently added directories.
// Upon finding nothing, it therefore returns the
// "warning" fsi.EmptyQueryResult
func (fs *AeFileSys) dirsByPath(path string) ([]os.FileInfo, error) {

	path = cleanseLeadingSlash(path)
	if !strings.HasPrefix(path, fs.RootName()) {
		path = fs.RootDir() + path
	}
	// logif.Pf("  %v", path)

	var fis []os.FileInfo

	dirs, err := fs.subdirsByPath(path, true)

	for _, v := range dirs {
		// logif.Pf("%15v => %-24v", "", v.Dir+v.BName)
		fi := os.FileInfo(v)
		fis = append(fis, fi)
	}

	sort.Sort(FileInfoByName(fis))

	return fis, err

}

func (fs *AeFileSys) saveDirByPath(path string) (AeDir, error) {

	fo := AeDir{}
	fo.isDir = true
	fo.MModTime = time.Now()
	fo.fSys = fs

	if path == fs.RootDir() || path+sep == fs.RootDir() {
		fo.Dir = fs.RootDir()
		fo.BName = ""
	} else {
		path = cleanseLeadingSlash(path)
		if !strings.HasPrefix(path, fs.RootName()) {
			path = fs.RootDir() + path
		}
		dir, base := pth.Split(path)
		fo.Dir = dir
		fo.BName = base
	}

	preciseK := ds.NewKey(fs.c, tdir, path, 0, nil)

	fo.Key = preciseK

	effKey, err := ds.Put(fs.c, preciseK, &fo)

	if err != nil {
		logif.E(err)
		return fo, err
	}

	if !preciseK.Equal(effKey) {
		fs.Ctx().Errorf("keys unequal %v - %v", preciseK, effKey)
	}

	fo.MemCacheSet()

	// recurse upwards
	_, err = fs.dirByPath(fo.Dir)
	if err == ds.ErrNoSuchEntity {
		_, err = fs.saveDirByPath(fo.Dir)
		if err != nil {
			return fo, err
		}
	}

	return fo, nil
}
