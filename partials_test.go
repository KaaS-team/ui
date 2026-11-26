package ui

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"text/template/parse"
)

// samples holds data for every partial, so a template that breaks on real fields fails here
// rather than in factory, kaasbot or dx. A new name in Names needs an entry.
var samples = map[string]any{
	"head":  Head{Title: "Runs", Description: "Every run today"},
	"shell": Shell{Head: Head{Title: "Runs"}, Nav: sampleNav, Body: "<p>Body</p>"},
	"nav":   sampleNav,
	"chip":  Chip{Label: "Passed", Status: StatusPass},
	"table": Table{Caption: "Runs", Columns: []string{"Run", "State"}, Rows: [][]string{{"42", "passed"}, {"43", "waiting"}}},
	"empty": Empty{Title: "No runs yet", Body: "Label an issue to start one.", Action: &Button{Label: "Open issues", Href: "/issues"}},
	"log": Log{Title: "Build", Lines: []LogLine{
		{Time: "10:01", Text: "go test ./..."}, {Time: "10:02", Status: StatusFail, Text: "FAIL ui"},
	}},
	"attention": Attention{Status: StatusWait, Title: "Two plans wait for you", Action: &Button{Label: "Review", Href: "/plans"}},
	"approval":  Approval{Title: "Plan for #12", Summary: "Adds the chip partial.", ApproveURL: "/approve/12", RejectURL: "/reject/12", CSRF: "t0k"},
	"timeline": Timeline{Events: []TimelineEvent{
		{Time: "09:00", Title: "Triaged", Status: StatusInfo}, {Time: "09:20", Title: "Built", Body: "Gate passed.", Status: StatusPass},
	}},
	"form": Form{Action: "/settings", CSRF: "t0k", Submit: Button{Label: "Save", Type: "submit", Variant: "primary"}, Fields: []Field{
		{Name: "repo", Label: "Repository", Required: true, Placeholder: "owner/name"},
		{Name: "notes", Label: "Notes", Type: "textarea", Help: "Shown on the plan."},
		{Name: "tier", Label: "Tier", Type: "select", Value: "b", Options: []string{"a", "b"}},
	}},
	"button": Button{Label: "Retry", Name: "action", Value: "retry"},
	"stat":   Stat{Label: "Runs today", Value: "12", Delta: "+3", Status: StatusPass},
}

var sampleNav = Nav{Brand: "Factory", Items: []NavItem{{Label: "Runs", Href: "/runs", Current: true}, {Label: "Plans", Href: "/plans"}}}

func TestEveryPartialRendersValidHTML(t *testing.T) {
	for _, name := range Names {
		data, ok := samples[name]
		if !ok {
			t.Errorf("%s: no sample data in partials_test.go", name)
			continue
		}
		var b strings.Builder
		if err := Partials.ExecuteTemplate(&b, name, data); err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		out := b.String()
		for _, bad := range []string{"<no value>", "ZgotmplZ"} {
			if strings.Contains(out, bad) {
				t.Errorf("%s: output contains %s:\n%s", name, bad, out)
			}
		}
		if err := balanced(out); err != "" {
			t.Errorf("%s: %s:\n%s", name, err, out)
		}
	}
}

// Each templates/<name>.tmpl defines exactly {{define "<name>"}}, and the files are exactly Names,
// so a caller can trust the list and a partial is found by its file name.
func TestFilesMatchNames(t *testing.T) {
	entries, err := templateFiles.ReadDir("templates")
	if err != nil {
		t.Fatal(err)
	}
	var files []string
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".tmpl")
		files = append(files, name)
		src, _ := templateFiles.ReadFile("templates/" + e.Name())
		trees, err := parse.Parse(name, string(src), "{{", "}}", map[string]any{"or": true, "eq": true})
		if err != nil {
			t.Fatal(err)
		}
		var defined []string
		for n, tree := range trees {
			if n != name || tree.Root != nil && len(tree.Root.Nodes) > 0 {
				defined = append(defined, n)
			}
		}
		if len(defined) != 1 || defined[0] != name {
			t.Errorf("%s defines %v, want exactly [%s]", e.Name(), defined, name)
		}
	}
	want := slices.Clone(Names)
	slices.Sort(want)
	if !slices.Equal(files, want) {
		t.Errorf("templates/ has %v, Names has %v", files, want)
	}
}

// Every Status has its two tokens, so a chip with a status is never drawn with an undefined colour.
func TestStatusesHaveTokens(t *testing.T) {
	raw, err := os.ReadFile("tokens.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Groups map[string]struct{ Names []string } `json:"groups"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	names := doc.Groups["status"].Names
	for _, s := range []Status{StatusPass, StatusWait, StatusFail, StatusInfo} {
		for _, tok := range []string{"--ui-" + string(s), "--ui-" + string(s) + "-bg"} {
			if !slices.Contains(names, tok) {
				t.Errorf("status %q: tokens.json has no %s", s, tok)
			}
		}
	}
}

func TestHandlerServesCSS(t *testing.T) {
	for path, want := range map[string]string{"/_ui/ui.css": "@import", "/_ui/tokens.css": "--ui-bg"} {
		rec := httptest.NewRecorder()
		Handler().ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		body, _ := io.ReadAll(rec.Body)
		if rec.Code != 200 || !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/css") || !strings.Contains(string(body), want) {
			t.Errorf("GET %s: %d %q, body lacks %q", path, rec.Code, rec.Header().Get("Content-Type"), want)
		}
	}
}

var (
	tagRe = regexp.MustCompile(`<(/?)([a-zA-Z][a-zA-Z0-9]*)\b[^>]*>`)
	void  = map[string]bool{"meta": true, "link": true, "input": true, "br": true, "hr": true, "img": true}
)

// balanced reports the first unclosed or mismatched element; html/template escapes text but not structure.
func balanced(html string) string {
	var stack []string
	for _, m := range tagRe.FindAllStringSubmatch(html, -1) {
		name := strings.ToLower(m[2])
		switch {
		case void[name]:
		case m[1] == "":
			stack = append(stack, name)
		case len(stack) == 0 || stack[len(stack)-1] != name:
			return "unexpected </" + name + "> with open " + strings.Join(stack, ">")
		default:
			stack = stack[:len(stack)-1]
		}
	}
	if len(stack) > 0 {
		return "unclosed " + strings.Join(stack, ">")
	}
	return ""
}
