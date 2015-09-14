package repo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/pbberlin/tools/net/http/fetch"
	"github.com/pbberlin/tools/net/http/loghttp"
	"github.com/pbberlin/tools/os/fsi"
	"github.com/pbberlin/tools/os/osutilpb"
	"github.com/pbberlin/tools/stringspb"
	"golang.org/x/net/html"
)

func dirTreeStrRec(buf *bytes.Buffer, d *DirTree, lvl int) {
	ind2 := strings.Repeat("    ", lvl+1)
	keys := make([]string, 0, len(d.Dirs))
	for k, _ := range d.Dirs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		buf.WriteString(ind2)
		indir := d.Dirs[key]
		buf.WriteString(stringspb.ToLen(indir.Name, 44-len(ind2)))
		if indir.EndPoint {
			buf.WriteString(fmt.Sprintf(" EP"))
		}
		buf.WriteByte(10)
		dirTreeStrRec(buf, &indir, lvl+1)
	}
}

func (d DirTree) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString(d.Name)
	// buf.WriteString(fmt.Sprintf(" %v ", len(d.Dirs)))
	if d.Dirs == nil {
		buf.WriteString(" (nil)")
	}
	buf.WriteByte(10)
	dirTreeStrRec(buf, &d, 0)
	return buf.String()
}

func switchTData(w http.ResponseWriter, r *http.Request) {

	lg, lge := loghttp.Logger(w, r)
	_ = lge

	b := fetch.TestData["test.economist.com"]
	sub1 := []byte(`<li><a href="/sections/newcontinent">xxx</a></li>`)

	sub2 := []byte(`<li><a href="/sections/asia">Asia</a></li>`)
	sub3 := []byte(`<li><a href="/sections/asia">Asia</a></li>
		<li><a href="/sections/newcontinent">xxx</a></li>`)

	if bytes.Contains(b, sub1) {
		b = bytes.Replace(b, sub1, []byte{}, -1)
	} else {
		b = bytes.Replace(b, sub2, sub3, -1)
	}

	if bytes.Contains(b, sub1) {
		lg("now contains %s", sub1)
	} else {
		lg("NOT contains %s", sub1)
	}

	fetch.TestData["test.economist.com"] = b

}

func path2DirTree(lg loghttp.FuncBufUniv, treeX *DirTree, articles []FullArticle, domain string, IsRSS bool) {

	if treeX == nil {
		return
	}
	var trLp *DirTree
	trLp = treeX

	pfx1 := "http://" + domain
	pfx2 := "https://" + domain

	for _, art := range articles {
		href := art.Url
		if art.Mod.IsZero() {
			art.Mod = time.Now()
		}
		href = strings.TrimPrefix(href, pfx1)
		href = strings.TrimPrefix(href, pfx2)
		if strings.HasPrefix(href, "/") { // ignore other domains
			parsed, err := url.Parse(href)
			lg(err)
			href = parsed.Path
			// lg("%v", href)
			trLp = treeX
			// lg("trLp is %v", trLp.String())
			dir, remainder, remDirs := "", href, []string{}
			lvl := 0
			for {

				dir, remainder, remDirs = osutilpb.PathDirReverse(remainder)

				if dir == "/" && remainder == "" {
					// skip root
					break
				}

				if lvl > 0 {
					trLp.Name = dir // lvl==0 => root
				}
				trLp.LastFound = art.Mod

				// lg("   %v, %v", dir, remainder)

				// New creation
				if _, ok := trLp.Dirs[dir]; !ok {
					if IsRSS {
						trLp.Dirs[dir] = DirTree{Name: dir, Dirs: map[string]DirTree{}, SrcRSS: true}
					} else {
						trLp.Dirs[dir] = DirTree{Name: dir, Dirs: map[string]DirTree{}}
					}
				}

				// We "cannot assign" to map struct directly:
				// trLp.Dirs[dir].LastFound = art.Mod   // fails with "cannot assign"
				addressable := trLp.Dirs[dir]
				addressable.LastFound = art.Mod

				// We can rely that the *last* dir or html is an endpoint.
				// We cannot tell about higher paths, unless explicitly linked somewhere
				// Previous distinction between RSS URLs and crawl URLs dropped
				if len(remDirs) < 1 {
					addressable.EndPoint = true
				}

				if dir == "/2015" || dir == "/08" || dir == "/09" {
					addressable.EndPoint = true
				}

				trLp.Dirs[dir] = addressable
				trLp = &addressable

				if remainder == "" {
					// lg("break\n")
					break
				}

				lvl++
			}

		}
	}

}

// Append of all links of a DOM to an in-memory dirtree
func addAnchors(lg loghttp.FuncBufUniv, host string, bts []byte, dirTree *DirTree) {

	doc, err := html.Parse(bytes.NewReader(bts))
	lg(err)
	if err != nil {
		return
	}
	anchors := []FullArticle{}
	var fr func(*html.Node)
	fr = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			art := FullArticle{}
			art.Url = attrX(n.Attr, "href")
			art.Mod = time.Now()
			anchors = append(anchors, art)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			fr(c)
		}
	}
	fr(doc)
	path2DirTree(lg, dirTree, anchors, host, false)
	dirTree.LastFound = time.Now() // Marker for later accumulated saving

}

func loadDigest(w http.ResponseWriter, r *http.Request, lg loghttp.FuncBufUniv, fs fsi.FileSystem, fnDigest string, treeX *DirTree) {

	bts, err := fs.ReadFile(fnDigest)

	lg(err)
	if err == nil {
		err = json.Unmarshal(bts, &treeX)
		lg(err)
	}

	lg("DirTree   %5.2vkB loaded for %v", len(bts)/1024, fnDigest)

}

// requesting via http; not from filesystem
// unused
func fetchDigest(hostWithPrefix, domain string) (*DirTree, error) {

	lg, lge := loghttp.Logger(nil, nil)
	_ = lg

	surl := path.Join(hostWithPrefix, domain, "digest2.json")
	bts, _, err := fetch.UrlGetter(nil, fetch.Options{URL: surl})
	lge(err)
	if err != nil {
		return nil, err
	}

	// lg("%s", bts)
	dirTree := &DirTree{Name: "/", Dirs: map[string]DirTree{}, EndPoint: true}

	if err == nil {
		err = json.Unmarshal(bts, dirTree)
		lge(err)
		if err != nil {
			return nil, err
		}
	}

	lg("DirTree   %5.2vkB loaded for %v", len(bts)/1024, surl)

	age := time.Now().Sub(dirTree.LastFound)
	lg("DirTree is %5.2v hours old (%v)", age.Hours(), dirTree.LastFound.Format(time.ANSIC))

	return dirTree, nil

}

func saveDigest(w http.ResponseWriter, r *http.Request, fs fsi.FileSystem, fnDigest string, treeX *DirTree) {

	lg, lge := loghttp.Logger(w, r)
	_ = lg

	treeX.LastFound = time.Now()

	b, err := json.MarshalIndent(treeX, "", "\t")
	lge(err)

	err = fs.MkdirAll(path.Dir(fnDigest), 0755)
	lge(err)

	err = fs.WriteFile(fnDigest, b, 0755)
	lge(err)

}

// Fetches URL if local file is outdated.
// saves fetched file
//
// link extraction, link addition to treeX now accumulated one level higher
func fetchCrawlSave(w http.ResponseWriter, r *http.Request,
	lg loghttp.FuncBufUniv, fs fsi.FileSystem, surl string) ([]byte, time.Time, error) {

	// Determine FileName
	ourl, err := fetch.URLFromString(surl)
	fc := FetchCommand{}
	fc.Host = ourl.Host
	fc = addDefaults(w, r, fc)
	semanticUri := condenseTrailingDir(surl, fc.CondenseTrailingDirs)
	fn := path.Join(docRoot, semanticUri)

	lg("crawlin %q", surl)

	// File already exists?
	// Open file for age check
	var bts []byte
	var mod time.Time
	f := func() error {
		file1, err := fs.Open(fn)
		// lg(err) // file may simply not exist
		if err != nil {
			return err // file may simply not exist
		}
		defer file1.Close() // file close *fast* at the end of *this* anonymous func

		fi, err := file1.Stat()
		lg(err)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			lg("\t\t file is a directory, skipping - %v", fn)
			return fmt.Errorf("is directory: %v", fn)
		}

		mod = fi.ModTime()
		age := time.Now().Sub(mod)
		if age.Hours() > 10 {
			lg("\t\t file %4.2v hours old, refetch ", age.Hours())
			return fmt.Errorf("too old: %v", fn)
		}

		lg("\t\t file only %4.2v hours old, skipping", age.Hours())
		bts, err = ioutil.ReadAll(file1)
		if err != nil {
			return err
		}
		return nil
	}

	err = f()
	if err == nil {
		return bts, mod, err
	}

	//
	// Fetch
	bts, inf, err := fetch.UrlGetter(r, fetch.Options{URL: surl, RedirectHandling: 1})
	lg(err)
	if err != nil {
		lg("tried to fetch %v, %v", surl, inf.URL)
		return []byte{}, inf.Mod, err
	}
	if inf.Mod.IsZero() {
		inf.Mod = time.Now().Add(-75 * time.Minute)
	}
	lg("retrivd %q; %vkB ", inf.URL.Host+inf.URL.Path, len(bts)/1024)

	//
	//
	lg("saved   %q crawled file", fn)
	dir := path.Dir(fn)
	err = fs.MkdirAll(dir, 0755)
	lg(err)
	err = fs.Chtimes(dir, time.Now(), time.Now())
	lg(err)
	err = fs.WriteFile(fn, bts, 0644)
	lg(err)
	err = fs.Chtimes(fn, inf.Mod, inf.Mod)
	lg(err)

	return bts, inf.Mod, nil

}
