// Package term is the KaaS design system for terminal UIs: the colour and status tokens of css/tokens.css as
// light and dark pairs, read from the embedded stylesheet at start-up, so a TUI matches the web with no copy.
//
// Color has the same shape as Lip Gloss v1's AdaptiveColor, so it converts directly:
//
//	style := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor(term.Default.Text))
//
// With Lip Gloss v2, pick by background: lipgloss.Color(term.Default.Text.Pick(hasDarkBackground)).
// The package itself has no dependencies beyond the ui module.
package term

import (
	"fmt"
	"io/fs"
	"reflect"
	"regexp"
	"strings"

	"github.com/KaaS-team/ui"
)

// Color is one token's value in light and dark mode, each a #rrggbb hex string.
type Color struct {
	Light string
	Dark  string
}

// Pick returns the value for a dark or light terminal background.
func (c Color) Pick(dark bool) string {
	if dark {
		return c.Dark
	}
	return c.Light
}

// Theme holds every colour and status token. The token tag on each field names the CSS custom property it
// is read from; Groups lists the tokens.json groups those tags must cover exactly.
type Theme struct {
	Bg      Color `token:"--ui-bg"`
	Surface Color `token:"--ui-surface"`
	Raised  Color `token:"--ui-raised"`
	Border  Color `token:"--ui-border"`
	// BorderStrong is the border of inputs and hovered rows.
	BorderStrong Color `token:"--ui-border-strong"`
	Hover        Color `token:"--ui-hover"`
	Text         Color `token:"--ui-text"`
	Muted        Color `token:"--ui-muted"`
	Accent       Color `token:"--ui-accent"`
	AccentHover  Color `token:"--ui-accent-hover"`
	OnAccent     Color `token:"--ui-on-accent"`
	Focus        Color `token:"--ui-focus"`

	Pass   Color `token:"--ui-pass"`
	PassBg Color `token:"--ui-pass-bg"`
	Wait   Color `token:"--ui-wait"`
	WaitBg Color `token:"--ui-wait-bg"`
	Fail   Color `token:"--ui-fail"`
	FailBg Color `token:"--ui-fail-bg"`
	Info   Color `token:"--ui-info"`
	InfoBg Color `token:"--ui-info-bg"`
}

// Groups are the tokens.json groups a terminal can draw. Space, radius, type, font, shadow and motion
// have no terminal meaning and are left out on purpose.
var Groups = []string{"colour", "status"}

// Default is the theme read from the embedded css/tokens.css.
var Default = mustLoad()

func mustLoad() Theme {
	t, err := Load(ui.CSS())
	if err != nil {
		panic(err) // css is embedded at build time; term_test.go fails first
	}
	return t
}

// Status returns the foreground and background for a status. The empty Status is neutral: muted on raised.
func (t Theme) Status(s ui.Status) (fg, bg Color) {
	switch s {
	case ui.StatusPass:
		return t.Pass, t.PassBg
	case ui.StatusWait:
		return t.Wait, t.WaitBg
	case ui.StatusFail:
		return t.Fail, t.FailBg
	case ui.StatusInfo:
		return t.Info, t.InfoBg
	}
	return t.Muted, t.Raised
}

// Tokens lists the CSS custom property behind each Theme field, in field order.
func Tokens() []string {
	rt := reflect.TypeFor[Theme]()
	names := make([]string, rt.NumField())
	for i := range names {
		names[i] = rt.Field(i).Tag.Get("token")
	}
	return names
}

var (
	declRe    = regexp.MustCompile(`(--[\w-]+)\s*:\s*([^;]+);`)
	commentRe = regexp.MustCompile(`(?s)/\*.*?\*/`)
	hexRe     = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
)

// Load reads tokens.css from css (the tree ui.CSS returns) and fills a Theme. A token that is missing or is
// not a hex colour is an error, since a terminal cannot draw it.
func Load(css fs.FS) (Theme, error) {
	src, err := fs.ReadFile(css, "tokens.css")
	if err != nil {
		return Theme{}, err
	}
	light, dark := Declarations(string(src))
	var t Theme
	v := reflect.ValueOf(&t).Elem()
	for i, name := range Tokens() {
		l, ok := light[name]
		if !ok {
			return Theme{}, fmt.Errorf("term: tokens.css does not declare %s", name)
		}
		d, ok := dark[name]
		if !ok {
			d = l
		}
		var c Color
		if c.Light, err = hex(name, l); err != nil {
			return Theme{}, err
		}
		if c.Dark, err = hex(name, d); err != nil {
			return Theme{}, err
		}
		v.Field(i).Set(reflect.ValueOf(c))
	}
	return t, nil
}

// Declarations returns the custom properties of a tokens.css: those before the prefers-color-scheme: dark
// block, and those inside it.
func Declarations(css string) (light, dark map[string]string) {
	css = commentRe.ReplaceAllString(css, "")
	lightSrc, darkSrc := css, ""
	if i := strings.Index(css, "@media (prefers-color-scheme: dark)"); i >= 0 {
		lightSrc, darkSrc = css[:i], css[i:]
	}
	read := func(s string) map[string]string {
		m := map[string]string{}
		for _, d := range declRe.FindAllStringSubmatch(s, -1) {
			m[d[1]] = strings.TrimSpace(d[2])
		}
		return m
	}
	return read(lightSrc), read(darkSrc)
}

// hex checks a value is #rgb or #rrggbb and returns it as lower-case #rrggbb.
func hex(name, v string) (string, error) {
	if !hexRe.MatchString(v) {
		return "", fmt.Errorf("term: %s is %q, want a hex colour a terminal can draw", name, v)
	}
	v = strings.ToLower(v)
	if len(v) == 4 {
		v = string([]byte{'#', v[1], v[1], v[2], v[2], v[3], v[3]})
	}
	return v, nil
}
