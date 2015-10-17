// Package htmlfrag contains convenience functions to generate html fragments.
package htmlfrag

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var sp func(format string, a ...interface{}) string = fmt.Sprintf

// GetSpanner is a simple helper to output
//  aligned html elements
//  for backend or debug pages
//  where the template package is not imperative
func GetSpanner() func(interface{}, int) string {

	cntr := 0
	style := `
		<style>
			.ib {
				vertical-align:middle;
				display:inline-block;
				width:95px;
			}
		</style>
	`
	return func(i interface{}, w int) string {

		s1 := fmt.Sprint(i)
		sw := fmt.Sprint(w)
		s2 := fmt.Sprintf("<span class='ib' style='width:" + sw + "px;'>" + s1 + "</span>")

		if cntr < 1 {
			s2 = style + "\n\n" + s2
		}
		cntr++

		return s2
	}

}

// Wb is a helper to write an inline block with an link inside.
// If url is omitted, a newline + a chapter-header is rendered.
func Wb(buf1 io.Writer, linktext, url string, descs ...string) {

	wstr := func(w io.Writer, s string) {
		w.Write([]byte(s))
	}

	if url == "" {
		wstr(buf1, "<br>\n")
	}

	if url == "nobr" { // hack, indicating no break
		url = ""
	}

	desc := ""
	if len(descs) > 0 {
		desc = descs[0]
	}

	const styleX = "display:inline-block; width:13%; margin: 4px 0px; margin-right:12px; vertical-align:top"

	wstr(buf1, "<span style='"+styleX+"'>\n")
	if url == "" {
		wstr(buf1, "\t<b>"+linktext+"</b>\n")
	} else {
		wstr(buf1, "\t<a target='_app' href='"+url+"' >"+linktext+"</a>\n")
		if desc != "" {
			wstr(buf1, "<br>\n ")
			wstr(buf1, "<span style='"+styleX+";width:90%;font-size:80%'>\n")
			wstr(buf1, desc)
			wstr(buf1, "</span>\n")
		}
	}
	wstr(buf1, "</span>\n")
}

// CSSColumnsWidth creates CSS classes with w[1...nCols]
//  for any number of screen widths
//  Thus you can write <div class='w2'> ...
//  and you get an element 2/nCols wide
//  and adapting to changes in the browser window
//  CSS class wmax equals the largest wx class.
//  CSS class wn   holds  the simple column minus a system wide margin, currently 10 pix
func CSSColumnsWidth(nCols int) string {

	// this is not really for the vertical scrollbar
	// but some slack is required, I don't know why
	// to prevent too early wrapping of blocks
	const wScrollbar = 16

	// this should not be required either
	// since we use 'box-sizing: border-box;'
	// to internalize all paddings and margins and borders :(
	const wMarginsPaddings = 10

	ret := new(bytes.Buffer)

	// the stepping is the max wasted space
	for w := 600; w <= 1600; w = w + 50 {

		ret.WriteString(sp("\n\t@media (min-width: %vpx) {\n\t\t", w))
		//ret.WriteString(sp("\n\t@media (min-device-width: %vpx) {\n\t\t", w))

		colWidthGross := (w - wScrollbar) / nCols
		colWidthNet := colWidthGross - wMarginsPaddings

		s := sp(".wn  { width: %vpx; } ", colWidthNet)
		ret.WriteString(s)

		for c := 1; c <= nCols; c++ {
			s := sp(".w%v{ width: %vpx; } ", c, c*colWidthGross)
			ret.WriteString(s)
		}
		s2 := sp(".wmax{ width: %vpx; } ", nCols*colWidthGross)
		ret.WriteString(s2)

		ret.WriteString("\n\t}")

	}

	return fmt.Sprintf("  /* generated by htmlfrag.CSSColumnsWidth()*/ \n%v\n", ret)
}

// CSSColumnsHeight is similar to ...Width except:
// 'min-height' attributes are generated
func CSSRowsHeight(nRows int) string {

	// This is so revolting -
	// but since the logic is still in development, it prevents duplicate code
	s := CSSColumnsWidth(nRows)

	var widthToHeight = strings.NewReplacer("wn", "hn",
		"wmax", "hmax",
		"w1", "h1",
		"w2", "h2",
		"w3", "h3",
		"w4", "h4",
		"w5", "h5",
		"w6", "h6",
		"w7", "h7",
		"w8", "h8",
		"min-width", "min-height",
		"width", "min-height")
	s = widthToHeight.Replace(s)

	return s
}

func SetNocacheHeaders(w http.ResponseWriter) {

	w.Header().Set("Expires", "Mon, 26 Jul 1990 05:00:00 GMT")

	tn := time.Now()
	tn = tn.Add(-10 * time.Second)
	tns := tn.Format(http.TimeFormat)
	w.Header().Set("Last-Modified", tns)
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Cache-Control", "post-check=0, pre-check=0")
	w.Header().Set("Pragma", "no-cache")

	// if plainOrHtml {
	// 	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// } else {
	// 	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// }

}
func CacheHeaders(w http.ResponseWriter) {

	w.Header().Set("Cache-control", "max-age=2592000") //30days (60sec * 60min * 24hours * 30days)
	w.Header().Set("Cache-control", "public")

	tn := time.Now()
	tn = tn.Add(-1 * 24 * 365 * time.Hour)
	tns := tn.Format(http.TimeFormat)
	w.Header().Set("Last-Modified", tns)
	// w.Header().Set("Date", tns)

}

func CookieDump(r *http.Request) string {
	str := ""
	c := r.Cookies()
	for _, v := range c {
		str += fmt.Sprintf("%v<br><br>\n\n", v)
	}
	return str
}
