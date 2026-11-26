package ui

import (
	"strings"
	"testing"
)

func render(t *testing.T, name string, data any) string {
	t.Helper()
	var b strings.Builder
	if err := Partials.ExecuteTemplate(&b, name, data); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if err := balanced(b.String()); err != "" {
		t.Fatalf("%s: %s:\n%s", name, err, b.String())
	}
	return b.String()
}

// Each case is a state a partial promises; the strings are what a screen reader or a POST handler relies on.
func TestPartialStates(t *testing.T) {
	cases := []struct {
		name, partial string
		data          any
		want, not     []string
	}{
		{"log is a log region", "log", Log{Title: "Build", Lines: []LogLine{{Text: "ok"}}}, []string{`role="log"`, `tabindex="0"`, `aria-label="Build output"`}, []string{"aria-busy"}},
		{"live log is busy", "log", Log{Live: true}, []string{`aria-busy="true"`, "No output yet."}, nil},
		{"table loading", "table", Table{Columns: []string{"A", "B"}, Loading: true}, []string{`aria-busy="true"`, `colspan="2"`, "Loading…"}, []string{"ui-empty"}},
		{"table error", "table", Table{Caption: "Runs", Error: "down"}, []string{`role="alert"`, "Runs: could not load", "down"}, []string{"<table"}},
		{"table empty", "table", Table{Empty: Empty{Title: "None"}}, []string{"ui-empty", "None"}, []string{"<table"}},
		{"fail attention alerts", "attention", Attention{Status: StatusFail, Title: "x"}, []string{`role="alert"`}, nil},
		{"wait attention is status", "attention", Attention{Status: StatusWait, Title: "x", Actions: []Button{{Label: "A"}, {Label: "B"}}}, []string{`role="status"`, `<div class="ui-row ui-attention-actions">`, ">A<", ">B<"}, nil},
		{"approval posts sha", "approval", Approval{Title: "P", ApproveURL: "/a", RejectURL: "/r", SHA: "abc1234", DiffURL: "/d"}, []string{`name="sha" value="abc1234"`, `<code>abc1234</code>`, `href="/d"`, "Approve", "Reject"}, nil},
		{"decided approval has no buttons", "approval", Approval{Title: "P", Decided: "Approved"}, []string{"Approved"}, []string{"<button"}},
		{"form field help and error", "form", Form{ID: "s", Fields: []Field{{Name: "n", Label: "N", Help: "h", Error: "e"}}}, []string{`id="s-n"`, `for="s-n"`, `aria-invalid="true"`, `aria-describedby="s-n-help s-n-error"`, `id="s-n-error"`}, nil},
		{"form error", "form", Form{Error: "bad"}, []string{`role="alert"`, `aria-describedby="f-error"`}, nil},
		{"disabled link is a button", "button", Button{Label: "Go", Href: "/x", Disabled: true}, []string{"<button", "disabled"}, []string{"href"}},
		{"nav label", "nav", Nav{Brand: "B", Label: "Sample"}, []string{`aria-label="Sample"`}, []string{"<ul>"}},
		{"empty timeline", "timeline", Timeline{}, []string{"Nothing has happened yet."}, []string{"<ol"}},
		{"stat without value", "stat", Stat{Label: "L"}, []string{"–"}, nil},
		{"head styles", "head", Head{Title: "T", Styles: []string{"/x.css"}}, []string{`href="/x.css"`, `name="color-scheme"`}, nil},
		{"shell main landmark", "shell", Shell{Head: Head{Title: "T"}}, []string{`<main class="ui-main" id="main">`}, nil},
	}
	for _, c := range cases {
		out := render(t, c.partial, c.data)
		for _, w := range c.want {
			if !strings.Contains(out, w) {
				t.Errorf("%s: lacks %q:\n%s", c.name, w, out)
			}
		}
		for _, n := range c.not {
			if strings.Contains(out, n) {
				t.Errorf("%s: has %q:\n%s", c.name, n, out)
			}
		}
	}
}
