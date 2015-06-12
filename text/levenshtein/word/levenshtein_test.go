package word

import (
	"fmt"

	ls_core "github.com/pbberlin/tools/text/levenshtein"
	"github.com/pbberlin/tools/util"

	"testing"
)

type TestCase struct {
	src      []Token
	dst      []Token
	distance int
}

var testCases = []TestCase{

	{[]Token{}, []Token{"wd1"}, 1},
	{[]Token{"wd1"}, []Token{"wd1", "wd1"}, 1},
	{[]Token{"wd1"}, []Token{"wd1", "wd1", "wd1"}, 2},

	{[]Token{}, []Token{}, 0},
	{[]Token{"wd1"}, []Token{"wd2"}, 2},
	{[]Token{"wd1", "wd1", "wd1"}, []Token{"wd1", "wd2", "wd1"}, 2},
	{[]Token{"wd1", "wd1", "wd1"}, []Token{"wd1", "wd2"}, 3},

	{[]Token{"wd1"}, []Token{"wd1"}, 0},
	{[]Token{"wd1", "wd2"}, []Token{"wd1", "wd2"}, 0},
	{[]Token{"wd1"}, []Token{}, 1},

	{[]Token{"wd1", "wd2"}, []Token{"wd1"}, 1},
	{[]Token{"wd1", "wd2", "wd3"}, []Token{"wd1"}, 2},

	{[]Token{"wd1", "wd1", "wd1"}, []Token{"wd1", "wd2", "wd1", "wd3"}, 3},
}

func init() {
	sss := [][]string{
		[]string{"wd1", "wd2", "up"},
		[]string{"trink", "nicht", "so", "viel", "Kaffee"},
		[]string{"nicht", "fuer", "Kinder", "ist", "Tuerkentrank"},
	}

	for i := 0; i < len(sss); i++ {
		st := make([]Token, 0, len(sss[i]))
		for j := 0; j < len(sss[i]); j++ {
			st = append(st, Token(sss[i][j]))
		}

		prev := testCases[len(testCases)-1]
		// log.Printf("%v", prev)
		testCases = append(testCases, TestCase{src: st, dst: prev.src, distance: 2})
	}

}

func TestLevenshtein(t *testing.T) {
	for i, tc := range testCases {

		mx := ls_core.New(convertToCore(tc.src), convertToCore(tc.dst), ls_core.DefaultOptions)
		got := mx.Distance()

		ssrc := fmt.Sprintf("%v", tc.src)
		sdst := fmt.Sprintf("%v", tc.dst)
		if got != tc.distance {
			t.Logf(
				"%2v: Distance between %20v and %20v should be %v - but got %v ",
				i, util.Ellipsoider(ssrc, 8), util.Ellipsoider(sdst, 8), tc.distance, got)
			t.Fail()
		}

		if i != 2 && i != 4 && i != 5 {
			continue
		}

		if i == 5 || i == 12 || i == 13 || true {
			mx.Print()

			es := mx.EditScript()
			es.Print()
		}

	}

}
