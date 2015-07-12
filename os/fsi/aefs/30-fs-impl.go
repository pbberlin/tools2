package aefs

import (
	"fmt"
	"os"
	"sync/atomic"

	"appengine/datastore"

	"github.com/pbberlin/tools/os/fsi"
	"github.com/pbberlin/tools/os/fsi/fsc"
)

func (fs AeFileSys) Name() string { return "aefs" }

func (fs AeFileSys) String() string { return fs.mount }

//---------------------------------------

// Create opens for read-write.
// Open opens for readonly access.
func (fs *AeFileSys) Create(name string) (fsi.File, error) {

	// WriteFile & Create
	dir, bname := fs.pathInternalize(name)

	f := AeFile{}
	f.fSys = fs
	f.BName = bname
	f.Dir = dir

	// let all the properties by set by fs.saveFileByPath
	err := f.Sync()
	if err != nil {
		return nil, err
	}

	// return &f, nil
	ff := fsi.File(&f)
	return ff, err

}

// No distinction between Stat (links are followed)
// and LStat (links go unresolved)
// We don't support links yet, anyway
func (fs *AeFileSys) Lstat(path string) (os.FileInfo, error) {
	fi, err := fs.Stat(path)
	return fi, err
}

// Strangely, neither MkdirAll nor Mkdir seem to have
// any concept of current working directory.
// They seem to operate relative to root.
func (fs *AeFileSys) Mkdir(name string, perm os.FileMode) error {
	_, err := fs.saveDirByPath(name)
	return err
}

func (fs *AeFileSys) MkdirAll(path string, perm os.FileMode) error {
	_, err := fs.saveDirByPath(path)
	return err
}

// Open() open existing file for readonly access.
// Create() should be used   for read-write.

// We could make provisions to ensure exclusive access;

// complies  with   os.Open()
// conflicts with  vfs.Open() signature
// conflicts with file.Open() interface of Afero
func (fs *AeFileSys) Open(name string) (fsi.File, error) {

	dir, bname := fs.pathInternalize(name)

	f, err := fs.fileByPath(dir + bname)
	if err != nil {
		return nil, err
	}

	atomic.StoreInt64(&f.at, 0) // why is this not nested into f.Lock()-f.Unlock()?

	if f.closed == false { // already open
		// return ErrFileInUse // instead of waiting for lock?
	}

	f.Lock()
	f.closed = false
	f.Unlock()

	// return &f, nil
	ff := fsi.File(&f)
	return ff, nil
}

func (fs *AeFileSys) OpenFile(name string, flag int, perm os.FileMode) (fsi.File, error) {
	return fs.Open(name)
}

// See fsi.FileSystem interface.
func (fs *AeFileSys) ReadDir(path string) ([]os.FileInfo, error) {
	dirs, err := fs.dirsByPath(path)
	if err != nil && err != fsi.EmptyQueryResult {
		return nil, err
	}
	files, err := fs.filesByPath(path)
	if err != nil {
		return nil, err
	}
	for _, v := range files {
		dirs = append(dirs, os.FileInfo(v))
	}
	return dirs, nil
}

func (fs *AeFileSys) Remove(name string) error {

	// log.Printf("trying to remove %-20v", name)
	f, err := fs.fileByPath(name)
	if err == nil {
		// log.Printf("   found file %v", f.Dir+f.BName)
		// log.Printf("   fkey %-26v", f.Key)
		err = datastore.Delete(fs.Ctx(), f.Key)
		if err != nil {
			return fmt.Errorf("error removing file %v", err)
		}
	}

	d, err := fs.dirByPath(name)
	if err == nil {
		// log.Printf("   dkey %v", d.Key)
		err = datastore.Delete(fs.Ctx(), d.Key)
		d.MemCacheDelete()
		if err != nil {
			return fmt.Errorf("error removing dir %v", err)
		}
	}

	return nil

}

func (fs *AeFileSys) RemoveAll(path string) error {

	paths := []string{}
	walkRemove := func(path string, f os.FileInfo, err error) error {
		if err != nil {
			// do nothing
		} else {
			if f != nil { // && f.IsDir() to constrain
				paths = append(paths, path)
			}
		}
		return nil
	}

	err := fsc.Walk(fs, path, walkRemove)
	if err != nil {
		fs.Ctx().Errorf("Error removing %v => %v", path, err)
	}

	// Walk crawls directories first, files second.
	// Intuitively removal in reverse order should always work. Or does it not?
	for i := 0; i < len(paths); i++ {
		iRev := len(paths) - 1 - i
		err := fs.Remove(paths[iRev])
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *AeFileSys) Rename(oldname, newname string) error {
	// we could use a walk similar to remove all
	return fsi.NotImplemented
}

func (fs *AeFileSys) Stat(path string) (os.FileInfo, error) {
	f, err := fs.fileByPath(path)
	if err != nil {
		// log.Printf("isno file err %-24q =>  %v", path, err)
		dir, err := fs.dirByPath(path)
		if err != nil {
			return nil, err
		}
		fiDir := os.FileInfo(dir)
		// log.Printf("Stat for dire %-24q => %-24v, %v", path, fiDir.Name(), err)
		return fiDir, nil
	} else {
		fiFi := os.FileInfo(f)
		// log.Printf("Stat for file %-24q => %-24v, %v", path, fiFi.Name(), err)
		return fiFi, nil
	}
}

func (fs *AeFileSys) ReadFile(path string) ([]byte, error) {

	file, err := fs.fileByPath(path)
	if err != nil {
		return []byte{}, err
	}
	return file.Data, err
}

// Only one save operation required
func (fs *AeFileSys) WriteFile(name string, data []byte, perm os.FileMode) error {

	// WriteFile & Create
	dir, bname := fs.pathInternalize(name)
	f := AeFile{}
	f.Dir = dir
	f.BName = bname
	f.fSys = fs

	var err error
	_, err = f.Write(data)
	if err != nil {
		return err
	}

	err = f.Sync()
	if err != nil {
		return err
	}

	return nil
}
