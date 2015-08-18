package fetch_rss

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"path"

	"appengine"

	"github.com/pbberlin/tools/net/http/loghttp"
	"github.com/pbberlin/tools/net/http/tplx"
	"github.com/pbberlin/tools/stringspb"
)

var ConfigDefaults = map[string]FetchCommand{
	"unspecified": FetchCommand{
		CondenseTrailingDirs: 0,
		DepthTolerance:       0,
		DesiredNumber:        5,
	},
	"www.handelsblatt.com": FetchCommand{
		CondenseTrailingDirs: 2,
		DepthTolerance:       1,
		DesiredNumber:        5,
	},
	"www.economist.com": FetchCommand{
		CondenseTrailingDirs: 0,
		DepthTolerance:       2,
		DesiredNumber:        5,
	},
	"www.welt.de": FetchCommand{
		CondenseTrailingDirs: 2,
		DepthTolerance:       0,
		DesiredNumber:        5,
	},
}

/*

[{ 	'Host':           'www.handelsblatt.com',
 	'RssXMLURI':      '/contentexport/feed/schlagzeilen',
 	'SearchPrefixs':  [ '/politik/international', '/politik/deutschland' ]
}]


curl -X POST -d "[{ \"Host\": \"www.handelsblatt.com\" }] "  localhost:8085/fetch/command-receive
curl -X POST -d "[{ \"Host\": \"www.handelsblatt.com\", 	\"RssXMLURI\": \"/contentexport/feed/schlagzeilen\", \"SearchPrefixs\": [ \"/politik/international\", \"/politik/deutschland\" ] }]"  localhost:8085/fetch/command-receive

curl -X POST -d "[{ \"Host\": \"www.welt.de\",  \"RssXMLURI\": \"/wirtschaft/?service=Rss\", \"SearchPrefixs\": [ \"/wirtschaft/deutschland\", \"/wirtschaft/international\" ] }]" localhost:8085/fetch/command-receive

"www.welt.de/wirtschaft/?service=Rss"

*/

func fetchCommandReceiver(w http.ResponseWriter, r *http.Request, m map[string]interface{}) {

	lg, lge := loghttp.Logger(w, r)

	var fcs []FetchCommand

	// The type of resp.body  <io.Reader> lends itself to using decode.
	// http://stackoverflow.com/ - ... using-json-unmarshal-vs-json-newdecoder-decode
	//
	// Nevertheless, we use Unmarshal here, because we want to inspect the bytes of body.
	var Unmarshal_versus_Decode = true

	if Unmarshal_versus_Decode {

		body, err := ioutil.ReadAll(r.Body) // no response write until here !
		lge(err)

		if len(body) == 0 {
			lg("empty body")
			return
		}

		err = json.Unmarshal(body, &fcs)
		if err != nil {
			lge(err)
			lg("body is %s", body)
			return
		}

	} else {

		dec := json.NewDecoder(r.Body)
		for {
			if err := dec.Decode(&fcs); err == io.EOF {
				break
			} else if err != nil {
				lge(err)
				return
			}
			lg("command loop is: %s", *stringspb.IndentedDump(fcs))
		}

	}

	FetchHTML(w, r, fcs)

}

func FetchHTML(w http.ResponseWriter, r *http.Request, fcs []FetchCommand) {

	lg, lge := loghttp.Logger(w, r)
	var err error

	fs := getFs(appengine.NewContext(r))
	// fs = fsi.FileSystem(memMapFileSys)

	wpf(w, tplx.ExecTplHelper(tplx.Head, map[string]string{"HtmlTitle": "Requesting files"}))
	defer wpf(w, tplx.Foot)

	wpf(w, "<pre>")
	defer wpf(w, "</pre>")

	err = fs.WriteFile(path.Join(docRoot, "msg.html"), msg, 0644)
	lge(err)

	// err = fs.WriteFile(path.Join(docRoot, "index.html"), []byte("content of index.html"), 0644)
	// lge(err)

	err = fs.MkdirAll(path.Join(docRoot, "testDirX/testDirY"), 0755)
	lge(err)

	for _, config := range fcs {
		config = addDefaults(config)
		Fetch(w, r, fs, config)
	}

	lg("fetching complete")

}

func addDefaults(in FetchCommand) FetchCommand {

	var preset FetchCommand

	h := in.Host
	if exactPreset, ok := ConfigDefaults[h]; ok {
		preset = exactPreset
	} else {
		preset = ConfigDefaults["unspecified"]
	}

	in.DepthTolerance = preset.DepthTolerance
	in.CondenseTrailingDirs = preset.CondenseTrailingDirs
	if in.DesiredNumber == 0 {
		in.DesiredNumber = preset.DesiredNumber
	}

	return in
}
