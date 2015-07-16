package backend

import (
	"bytes"
	"net/http"

	"appengine"

	"github.com/pbberlin/tools/net/http/htmlfrag"
	"github.com/pbberlin/tools/net/http/loghttp"
	"github.com/pbberlin/tools/net/http/tplx"

	"github.com/pbberlin/tools/os/fsi"
	"github.com/pbberlin/tools/os/fsi/aefs"
	"github.com/pbberlin/tools/os/fsi/fstest"
	"github.com/pbberlin/tools/os/fsi/memfs"
	"github.com/pbberlin/tools/os/fsi/osfs"
)

var backendFragFsiAefs = new(bytes.Buffer)

var memMapFileSys = memfs.New()
var osFileSys = osfs.New()

func init() {

	//
	// handler registration
	http.HandleFunc("/fs/aefs/create-objects", loghttp.Adapter(createSys))
	http.HandleFunc("/fs/aefs/retrieve-by-query", loghttp.Adapter(retrieveByQuery))
	http.HandleFunc("/fs/aefs/retrieve-by-read-dir", loghttp.Adapter(retrieveByReadDir))
	http.HandleFunc("/fs/aefs/walk", loghttp.Adapter(walkH))
	http.HandleFunc("/fs/aefs/remove", loghttp.Adapter(removeSubtree))

	http.HandleFunc("/fs/aefs/delete-all", loghttp.Adapter(deleteAll))

	http.HandleFunc("/fs/aefs/reset", loghttp.Adapter(resetMountPoint))
	http.HandleFunc("/fs/aefs/decr", loghttp.Adapter(decrMountPoint))

	//
	// admin widgets
	htmlfrag.Wb(backendFragFsiAefs, "create", "/fs/aefs/create-objects")

	htmlfrag.Wb(backendFragFsiAefs, "query", "/fs/aefs/retrieve-by-query")
	htmlfrag.Wb(backendFragFsiAefs, "readdir", "/fs/aefs/retrieve-by-read-dir")
	htmlfrag.Wb(backendFragFsiAefs, "walk", "/fs/aefs/walk")
	htmlfrag.Wb(backendFragFsiAefs, "remove", "/fs/aefs/remove")

	htmlfrag.Wb(backendFragFsiAefs, "delete all fs entities", "/fs/aefs/delete-all")

	htmlfrag.Wb(backendFragFsiAefs, "reset", "/fs/aefs/reset")
	htmlfrag.Wb(backendFragFsiAefs, "decr", "/fs/aefs/decr")

}

func callTestX(w http.ResponseWriter, r *http.Request,
	f1 func() string,
	f2 func(fsi.FileSystem) (*bytes.Buffer, string)) {

	wpf(w, tplx.Head)
	wpf(w, "<pre>\n")
	defer wpf(w, tplx.Foot)
	defer wpf(w, "\n</pre>")

	var fs fsi.FileSystem

	if false {
		fsc := aefs.New(aefs.MountName(f1()), aefs.AeContext(appengine.NewContext(r)))
		fs = fsi.FileSystem(fsc)
	} else if false {
		fs = fsi.FileSystem(osFileSys)
	} else {
		fs = fsi.FileSystem(memMapFileSys)
	}

	bb := new(bytes.Buffer)
	msg := ""
	wpf(bb, "created fs %v\n\n", aefs.MountPointLast())
	bb, msg = f2(fs)
	w.Write([]byte(msg))
	w.Write([]byte("\n\n"))
	w.Write(bb.Bytes())

}

func createSys(w http.ResponseWriter, r *http.Request, m map[string]interface{}) {
	callTestX(w, r, aefs.MountPointNext, fstest.CreateSys)
}

func retrieveByQuery(w http.ResponseWriter, r *http.Request, m map[string]interface{}) {
	callTestX(w, r, nil, fstest.RetrieveByQuery)
}

func retrieveByReadDir(w http.ResponseWriter, r *http.Request, m map[string]interface{}) {
	callTestX(w, r, nil, fstest.RetrieveByReadDir)
}

func walkH(w http.ResponseWriter, r *http.Request, m map[string]interface{}) {
	callTestX(w, r, nil, fstest.WalkDirs)
}

func removeSubtree(w http.ResponseWriter, r *http.Request, m map[string]interface{}) {
	callTestX(w, r, nil, fstest.RemoveSubtree)
}

//
// aefs specific
func deleteAll(w http.ResponseWriter, r *http.Request, m map[string]interface{}) {

	wpf(w, tplx.Head)
	wpf(w, "<pre>\n")
	defer wpf(w, "\n</pre>")
	defer wpf(w, tplx.Foot)

	fs := aefs.New(aefs.AeContext(appengine.NewContext(r)))
	wpf(w, "aefs:\n")
	msg, err := fs.DeleteAll()
	if err != nil {
		wpf(w, "err during delete %v\n", err)
	}
	wpf(w, msg)

	memMapFileSys = memfs.New()
	wpf(w, "\n")
	wpf(w, "memMapFs new")

	osFileSys = osfs.New()

}

func resetMountPoint(w http.ResponseWriter, r *http.Request, m map[string]interface{}) {

	wpf(w, tplx.Head)
	wpf(w, "<pre>\n")
	defer wpf(w, "\n</pre>")
	defer wpf(w, tplx.Foot)

	wpf(w, "reset %v\n", aefs.MountPointReset())

}

func decrMountPoint(w http.ResponseWriter, r *http.Request, m map[string]interface{}) {

	wpf(w, tplx.Head)
	wpf(w, "<pre>\n")
	defer wpf(w, "\n</pre>")
	defer wpf(w, tplx.Foot)

	wpf(w, "counted down %v\n", aefs.MountPointDecr())

}
