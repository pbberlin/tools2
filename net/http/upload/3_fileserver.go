package upload

import (
	"bytes"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/pbberlin/tools/net/http/tplx"
	"github.com/pbberlin/tools/os/fsi/dsfs"

	"appengine"
)

// We cannot use http.FileServer(http.Dir("./css/") to dispatch our dsfs files.
// We need the appengine context to initialize dsfs.
//
// Thus we re-implement a serveFile method:
func ServeDsFsFile(w http.ResponseWriter, r *http.Request, m map[string]interface{}) {

	c := appengine.NewContext(r)

	b1 := new(bytes.Buffer)

	fclose := func() {
		// Only upon error.
		// If everything is fine, we reset fclose at the end.
		w.Write(b1.Bytes())
	}
	defer fclose()

	wpf(b1, tplx.ExecTplHelper(tplx.Head, map[string]string{"HtmlTitle": "Half-Static-File-Server"}))
	wpf(b1, "<pre>\n")

	mnt := dsfs.MountPointLast()

	p := r.URL.Path
	if strings.HasPrefix(p, "/") {
		p = p[1:]
	}

	wpf(b1, "raw path = %q \n", p)

	dirs := strings.Split(p, "/")
	// wpf(b1, "dirs = %q \n", dirs)
	if len(dirs) > 0 {
		mnt = dirs[0]
		p = strings.Join(dirs[1:], "/")
	}

	wpf(b1, "mnt = %q  path = %q \n", mnt, p)

	if p == "" {
		p += "index.html"
	}

	if len(p) > 0 {

		fs1 := dsfs.New(
			dsfs.MountName(mnt),
			dsfs.AeContext(c),
		)

		fullP := path.Join(docRootDataStore, p)

		f, err := fs1.Open(fullP)
		if err != nil {
			wpf(b1, "err opening file %v - %v\n", fullP, err)
			return
		}

		inf, err := f.Stat()
		if err != nil {
			wpf(b1, "err opening fileinfo %v - %v\n", fullP, err)
			return
		}

		if inf.IsDir() {

			wpf(b1, "%v is a directory - trying index.html...\n", fullP)

			fullP += "/index.html"

			f, err = fs1.Open(fullP)
			if err != nil {
				wpf(b1, "err opening index file %v - %v\n", fullP, err)
				return
			}

			inf, err = f.Stat()
			if err != nil {
				wpf(b1, "err opening index fileinfo %v - %v\n", fullP, err)
				return
			}

		}

		wpf(b1, "opened file %v - %v -  %v\n", f.Name(), inf.Size(), err)

		bts1, err := ioutil.ReadAll(f)
		if err != nil {
			wpf(b1, "err with ReadAll %v - %v\n", fullP, err)
			return
		}

		tp := mime.TypeByExtension(path.Ext(fullP))

		w.Header().Set("Content-Type", tp)
		w.Write(bts1)

		// on success -
		// reset the message log
		// => dumps an empty buffer
		b1 = new(bytes.Buffer)

	}

}