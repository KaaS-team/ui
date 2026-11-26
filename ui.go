// Package ui is the KaaS design system for Go binaries: the CSS embedded from css/, served at /_ui/,
// and the html/template partials every server page is built from.
//
//	mux.Handle(ui.Prefix, ui.Handler())
//	t := template.Must(ui.Partials.Clone())
//	template.Must(t.ParseFS(pages, "*.tmpl"))
package ui

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
)

// Prefix is the URL path the CSS is served under. The head partial links Prefix + "ui.css".
const Prefix = "/_ui/"

//go:embed css
var cssFiles embed.FS

//go:embed templates/*.tmpl
var templateFiles embed.FS

// Names lists every partial, one templates/<name>.tmpl each. Callers execute them by these names;
// partials_test.go fails when a file and this list disagree.
var Names = []string{
	"head", "shell", "nav", "chip", "table", "empty", "log",
	"attention", "approval", "timeline", "form", "button", "stat",
}

// Partials holds every partial. Clone it before adding page templates so the shared set stays untouched.
var Partials = template.Must(template.New("ui").ParseFS(templateFiles, "templates/*.tmpl"))

// CSS returns the stylesheet tree rooted at css/ (ui.css, tokens.css and the layer files).
func CSS() fs.FS {
	sub, err := fs.Sub(cssFiles, "css")
	if err != nil {
		panic(err) // unreachable: css is embedded at build time
	}
	return sub
}

// Handler serves CSS() under Prefix. Mount it at Prefix.
func Handler() http.Handler {
	return http.StripPrefix(Prefix, http.FileServerFS(CSS()))
}
