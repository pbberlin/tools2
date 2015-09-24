package weedout

import (
	"bytes"
	"fmt"
	"net/http"

	"appengine"

	"github.com/pbberlin/tools/net/http/domclean2"
	"github.com/pbberlin/tools/net/http/fetch"
	"github.com/pbberlin/tools/net/http/htmlfrag"
	"github.com/pbberlin/tools/net/http/loghttp"
	"github.com/pbberlin/tools/net/http/routes"
	"github.com/pbberlin/tools/net/http/tplx"
	"golang.org/x/net/html"
)

// InitHandlers is called from outside,
// and makes the EndPoints available.
func InitHandlers() {
	http.HandleFunc(routes.WeedOutURI, loghttp.Adapter(weedOutHTTP))
}

// BackendUIRendered returns a userinterface rendered to HTML
func BackendUIRendered() *bytes.Buffer {
	var b1 = new(bytes.Buffer)
	htmlfrag.Wb(b1, "Weed Out", "")

	fullURL := fmt.Sprintf("%s?%s=%s&cnt=%v", routes.WeedOutURI, routes.URLParamKey, URLs[0], numTotal-1)
	htmlfrag.Wb(b1, "Weed out", fullURL)

	return b1
}

// weedOutHTTP wraps WeedOut()
func weedOutHTTP(w http.ResponseWriter, r *http.Request, m map[string]interface{}) {

	lg, b := loghttp.BuffLoggerUniversal(w, r)
	closureOverBuf := func(bUnused *bytes.Buffer) {
		loghttp.Pf(w, r, b.String())
	}
	defer closureOverBuf(b) // the argument is ignored,

	r.Header.Set("X-Custom-Header-Counter", "nocounter")

	wpf(b, tplx.ExecTplHelper(tplx.Head, map[string]string{"HtmlTitle": "Weedout redundant stuff"}))
	defer wpf(b, tplx.Foot)

	wpf(b, "<pre>")
	defer wpf(b, "</pre>")

	err := r.ParseForm()
	lg(err)

	// countSimilar := 3
	// sCountSimilar := r.FormValue("cnt")
	// if sCountSimilar != "" {
	// 	i, err := strconv.Atoi(strings.TrimSpace(sCountSimilar))
	// 	if err == nil {
	// 		countSimilar = i
	// 	}
	// }

	surl := r.FormValue(routes.URLParamKey)
	ourl, err := fetch.URLFromString(surl)
	lg(err)
	if err != nil {
		return
	}
	if ourl.Host == "" {
		lg("host is empty (%v)", surl)
		return
	}

	lg("%v, %v", ourl.Host, ourl.Path)

	fs := GetFS(appengine.NewContext(r), 0)

	least3Files := FetchAndDecodeJSON(r, ourl.String(), lg, fs)

	lg("Fetched and decoded; found %v", len(least3Files))
	if len(least3Files) > 0 {
		doc := WeedOut(least3Files, lg, fs)

		fNamer := domclean2.FileNamer(logDir, 0)
		fNamer() // first call yields key
		fsPerm := GetFS(appengine.NewContext(r), 0)
		fileDump(lg, fsPerm, doc, fNamer, "_fin.html")

		lg("MapSimiliarCompares: %v SimpleCompares: %v LevenstheinComp: %v\n", breakMapsTooDistinct, appliedLevenshtein, appliedCompare)
		lg("Finish\n")

		var b2 bytes.Buffer
		err := html.Render(&b2, doc)
		lg(err)
		if err != nil {
			return
		}

		b = new(bytes.Buffer)
		// w.Write([]byte("aa"))
		w.Header().Set("Content-type", "text/html; charset=utf-8")
		w.Write(b2.Bytes())

	}

}