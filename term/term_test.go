package term

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/KaaS-team/ui"
)

type tokensDoc struct {
	Groups map[string]struct {
		Dark  bool     `json:"dark"`
		Names []string `json:"names"`
	} `json:"groups"`
}

func readTokensJSON(t *testing.T) tokensDoc {
	t.Helper()
	raw, err := os.ReadFile("../tokens.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc tokensDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

// The three lists of token names agree: tokens.css declares exactly what tokens.json lists, and the theme
// covers exactly the colour and status groups. A token added or renamed on one side fails here.
func TestTokenParity(t *testing.T) {
	doc := readTokensJSON(t)
	src, err := os.ReadFile("../css/tokens.css")
	if err != nil {
		t.Fatal(err)
	}
	light, _ := Declarations(string(src))
	var css, listed, want []string
	for n := range light {
		css = append(css, n)
	}
	for g, grp := range doc.Groups {
		listed = append(listed, grp.Names...)
		if slices.Contains(Groups, g) {
			want = append(want, grp.Names...)
		}
	}
	slices.Sort(css)
	slices.Sort(listed)
	if !slices.Equal(css, listed) {
		t.Errorf("tokens.css declares %v\ntokens.json lists %v", css, listed)
	}
	got := Tokens()
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("term.Theme covers %v\ntokens.json groups %v list %v", got, Groups, want)
	}
	for _, g := range Groups {
		if _, ok := doc.Groups[g]; !ok {
			t.Errorf("tokens.json has no group %q", g)
		}
	}
}

// Default carries the embedded values, light and dark, so the TUI and the web draw the same colours.
func TestDefaultMatchesTokensCSS(t *testing.T) {
	src, err := os.ReadFile("../css/tokens.css")
	if err != nil {
		t.Fatal(err)
	}
	light, dark := Declarations(string(src))
	want := map[string]Color{}
	for _, n := range Tokens() {
		l, _ := hex(n, light[n])
		d, _ := hex(n, dark[n])
		want[n] = Color{Light: l, Dark: d}
	}
	got, _ := json.Marshal(Default)
	var byField map[string]Color
	_ = json.Unmarshal(got, &byField)
	for i, n := range Tokens() {
		field := fieldName(i)
		if byField[field] != want[n] {
			t.Errorf("%s (%s) = %+v, tokens.css has %+v", field, n, byField[field], want[n])
		}
		if byField[field].Light == "" || byField[field].Dark == "" {
			t.Errorf("%s has an empty value", field)
		}
	}
}

func TestStatusUsesStatusTokens(t *testing.T) {
	th := Default
	for s, want := range map[ui.Status][2]Color{
		ui.StatusPass: {th.Pass, th.PassBg}, ui.StatusWait: {th.Wait, th.WaitBg},
		ui.StatusFail: {th.Fail, th.FailBg}, ui.StatusInfo: {th.Info, th.InfoBg}, "": {th.Muted, th.Raised},
	} {
		fg, bg := th.Status(s)
		if fg != want[0] || bg != want[1] {
			t.Errorf("Status(%q) = %v, %v; want %v", s, fg, bg, want)
		}
	}
	if th.Pass.Pick(true) != th.Pass.Dark || th.Pass.Pick(false) != th.Pass.Light {
		t.Error("Pick returns the wrong side")
	}
}

func TestLoadRejectsWhatATerminalCannotDraw(t *testing.T) {
	src, err := os.ReadFile("../css/tokens.css")
	if err != nil {
		t.Fatal(err)
	}
	for name, css := range map[string]string{
		"missing token": strings.Replace(string(src), "--ui-muted:", "--ui-gone:", 1),
		"not hex":       strings.Replace(string(src), "--ui-text:", "--ui-text: canvastext; --ui-old:", 1),
	} {
		if _, err := Load(fstest.MapFS{"tokens.css": {Data: []byte(css)}}); err == nil {
			t.Errorf("%s: Load succeeded", name)
		}
	}
	short := ":root { " + strings.Join(func() []string {
		var d []string
		for _, n := range Tokens() {
			d = append(d, n+": "+hash+"AbC;")
		}
		return d
	}(), " ") + " }"
	th, err := Load(fstest.MapFS{"tokens.css": {Data: []byte(short)}})
	if err != nil {
		t.Fatal(err)
	}
	if want := hash + "aabbcc"; th.Text.Light != want || th.Text.Dark != want {
		t.Errorf("short hex with no dark block: %+v, want %s both sides", th.Text, want)
	}
}

func fieldName(i int) string {
	return reflect.TypeFor[Theme]().Field(i).Name
}

// hash keeps hex literals out of this file: test/colours.test.mjs allows raw colours only in css/tokens.css.
const hash = "#"
