package ui

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"
)

// CataloguePath is where Catalogue serves its page. Its stylesheet sits under CataloguePath + "/".
const CataloguePath = Prefix + "catalogue"

//go:embed catalogue
var catalogueFiles embed.FS

// Catalogue serves one page that renders every partial in every state it has (empty, loading, error,
// long text, many rows), built from sample data only, for screenshots and visual tests. Mount it next to Handler:
//
//	mux.Handle(ui.Prefix, ui.Handler())
//	mux.Handle(ui.CataloguePath, ui.Catalogue())
//	mux.Handle(ui.CataloguePath+"/", ui.Catalogue())
func Catalogue() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		switch r.URL.Path {
		case CataloguePath, CataloguePath + "/":
			page, err := cataloguePage()
			if err != nil {
				http.Error(w, "catalogue failed to render", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(page)
		case CataloguePath + "/catalogue.css":
			css, _ := catalogueFiles.ReadFile("catalogue/catalogue.css")
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
			_, _ = w.Write(css)
		default:
			http.NotFound(w, r)
		}
	})
}

// CatalogueSection is one partial on the catalogue page and the states it is drawn in.
type CatalogueSection struct {
	Name    string // the partial, one of Names
	Summary string
	States  []CatalogueState
}

// CatalogueState is one rendering of a partial. Data is executed with that partial; Wide spans the row.
type CatalogueState struct {
	Label string
	Data  any
	Wide  bool
	// Many repeats Data side by side, for partials that are usually seen in a row (chips, stats, buttons).
	Many []any
}

var cataloguePage = sync.OnceValues(renderCatalogue)

func renderCatalogue() ([]byte, error) {
	page, err := template.New("catalogue").ParseFS(catalogueFiles, "catalogue/page.tmpl")
	if err != nil {
		return nil, err
	}
	type state struct {
		Label string
		Wide  bool
		HTML  template.HTML
	}
	type section struct {
		Name, Summary string
		States        []state
	}
	var sections []section
	for _, cs := range CatalogueSections() {
		s := section{Name: cs.Name, Summary: cs.Summary}
		for _, st := range cs.States {
			html, err := renderState(cs.Name, st)
			if err != nil {
				return nil, fmt.Errorf("%s / %s: %w", cs.Name, st.Label, err)
			}
			s.States = append(s.States, state{Label: st.Label, Wide: st.Wide, HTML: html})
		}
		sections = append(sections, s)
	}
	var body bytes.Buffer
	if err := page.ExecuteTemplate(&body, "page", sections); err != nil {
		return nil, err
	}
	nav := Nav{Brand: "KaaS UI", Label: "Catalogue"}
	for _, s := range sections {
		nav.Items = append(nav.Items, NavItem{Label: s.Name, Href: "#" + s.Name})
	}
	var out bytes.Buffer
	err = Partials.ExecuteTemplate(&out, "shell", Shell{
		Head: Head{
			Title:       "Catalogue · KaaS UI",
			Description: "Every KaaS UI partial in every state.",
			Styles:      []string{CataloguePath + "/catalogue.css"},
		},
		Nav:  nav,
		Body: template.HTML(body.String()), // rendered by html/template above, so already escaped
	})
	return out.Bytes(), err
}

// renderState executes one partial for a state. head and shell are whole-document partials, so the
// catalogue shows head's output as source and shell as the page itself.
func renderState(name string, st CatalogueState) (template.HTML, error) {
	var b strings.Builder
	data := st.Many
	if data == nil {
		data = []any{st.Data}
	}
	if name == "head" {
		var src strings.Builder
		if err := Partials.ExecuteTemplate(&src, name, st.Data); err != nil {
			return "", err
		}
		b.WriteString(`<pre class="cat-source" tabindex="0" aria-label="Rendered head">`)
		template.HTMLEscape(&b, []byte(src.String()))
		b.WriteString(`</pre>`)
		return template.HTML(b.String()), nil
	}
	if len(data) > 1 {
		b.WriteString(`<div class="cat-row">`)
	}
	for _, d := range data {
		if err := Partials.ExecuteTemplate(&b, name, d); err != nil {
			return "", err
		}
	}
	if len(data) > 1 {
		b.WriteString(`</div>`)
	}
	return template.HTML(b.String()), nil
}

const longText = "This line is long on purpose: it checks that a partial wraps a sentence that runs well past the width of its column instead of clipping it, overflowing the card or pushing the layout sideways on a phone."

// CatalogueSections is the sample data the catalogue draws, one section per name in Names, in that order.
func CatalogueSections() []CatalogueSection {
	many := make([][]string, 30)
	for i := range many {
		state := []string{"passed", "waiting", "failed"}[i%3]
		many[i] = []string{fmt.Sprintf("#%d", 1200+i), "kaas-team/factory", state, fmt.Sprintf("%dm %02ds", 1+i%7, (i*17)%60)}
	}
	var manyLines []LogLine
	for i := range 40 {
		l := LogLine{Time: fmt.Sprintf("10:%02d:%02d", i/6, (i*10)%60), Text: fmt.Sprintf("step %d of 40: ok", i+1)}
		if i == 39 {
			l.Status, l.Text = StatusPass, "PASS all 40 steps"
		}
		manyLines = append(manyLines, l)
	}
	var manyEvents []TimelineEvent
	for i, t := range []string{"Opened", "Triaged", "Planned", "Plan approved", "Built", "Gate passed", "Reviewed", "Merged"} {
		manyEvents = append(manyEvents, TimelineEvent{Time: fmt.Sprintf("09:%02d", i*7), Title: t, Status: []Status{StatusInfo, "", StatusInfo, StatusPass, "", StatusPass, StatusWait, StatusPass}[i]})
	}
	review := &Button{Label: "Review", Href: "#attention"}
	return []CatalogueSection{
		{Name: "head", Summary: "The head element of every page. Shown as the HTML it renders.", States: []CatalogueState{
			{Label: "Title and description", Data: Head{Title: "Runs · Factory", Description: "Every run today"}, Wide: true},
			{Label: "Extra stylesheet", Data: Head{Title: "Runs", Styles: []string{"/static/runs.css"}}, Wide: true},
		}},
		{Name: "shell", Summary: "A whole page: the nav rail beside the main column. This page is drawn with it.", States: nil},
		{Name: "nav", Summary: "The product name and its links. The current page carries aria-current.", States: []CatalogueState{
			{Label: "With a current page", Data: Nav{Brand: "Factory", Label: "Sample nav", Items: []NavItem{{Label: "Runs", Href: "#nav", Current: true}, {Label: "Plans", Href: "#nav"}, {Label: "Settings", Href: "#nav"}}}},
			{Label: "Brand only", Data: Nav{Brand: "Kaasbot", Label: "Sample nav, brand only"}},
			{Label: "Long labels", Data: Nav{Brand: "Factory for a very long organisation name", Label: "Sample nav, long labels", Items: []NavItem{{Label: "Runs that are waiting for a person to look at them", Href: "#nav"}}}},
		}},
		{Name: "chip", Summary: "A short status label. Colour comes from the status tokens only.", States: []CatalogueState{
			{Label: "Every status", Many: []any{Chip{Label: "Neutral"}, Chip{Label: "Passed", Status: StatusPass}, Chip{Label: "Waiting", Status: StatusWait}, Chip{Label: "Failed", Status: StatusFail}, Chip{Label: "Info", Status: StatusInfo}}},
			{Label: "Long label", Data: Chip{Label: "Waiting for review from the platform team", Status: StatusWait}},
		}},
		{Name: "table", Summary: "A data table. With no rows it shows its empty state; it also has loading and error states.", States: []CatalogueState{
			{Label: "Rows", Data: Table{Caption: "Runs today", Columns: []string{"Run", "Repository", "State", "Took"}, Rows: many[:4]}, Wide: true},
			{Label: "Loading", Data: Table{Caption: "Runs yesterday", Columns: []string{"Run", "Repository", "State", "Took"}, Loading: true}},
			{Label: "Error", Data: Table{Caption: "Runs today", Columns: []string{"Run"}, Error: "The run list did not load. Reload the page to try again."}},
			{Label: "Empty", Data: Table{Caption: "Runs today", Columns: []string{"Run"}, Empty: Empty{Title: "No runs yet", Body: "Label an issue to start one.", Action: &Button{Label: "Open issues", Href: "#table"}}}},
			{Label: "Long text", Data: Table{Caption: "Plans", Columns: []string{"Plan", "Summary"}, Rows: [][]string{{"#12", longText}, {"#13", "Short."}}}},
			{Label: "Many rows", Data: Table{Caption: "Runs this week", Columns: []string{"Run", "Repository", "State", "Took"}, Rows: many}, Wide: true},
		}},
		{Name: "empty", Summary: "What a list says when it has nothing yet, and the one action that fills it.", States: []CatalogueState{
			{Label: "With an action", Data: Empty{Title: "No plans yet", Body: "Plans appear here when an issue is triaged.", Action: &Button{Label: "Triage an issue", Href: "#empty", Variant: "primary"}}},
			{Label: "Title only", Data: Empty{Title: "Nothing waits for you"}},
			{Label: "Long text", Data: Empty{Title: "No runs match these filters", Body: longText}},
		}},
		{Name: "log", Summary: "Run output in an ARIA log region, so new lines are announced. Scrolls on its own and takes keyboard focus.", States: []CatalogueState{
			{Label: "Statuses", Data: Log{Title: "Build", Lines: []LogLine{
				{Time: "10:01", Text: "go test ./..."}, {Time: "10:01", Status: StatusInfo, Text: "running 42 tests"},
				{Time: "10:02", Status: StatusWait, Text: "retrying flaky test once"}, {Time: "10:02", Status: StatusFail, Text: "FAIL ui 0.01s"},
				{Time: "10:03", Status: StatusPass, Text: "ok  term 0.01s"},
			}}},
			{Label: "Live, still writing", Data: Log{Title: "Deploy", Live: true, Lines: []LogLine{{Time: "11:00", Text: "pushing image"}}}},
			{Label: "Empty", Data: Log{Title: "Lint"}},
			{Label: "Long lines", Data: Log{Title: "Trace", Lines: []LogLine{{Time: "12:00", Status: StatusFail, Text: "error: " + strings.Repeat("a long unbroken path/segment/", 8)}}}},
			{Label: "Many lines", Data: Log{Title: "Test", Lines: manyLines}, Wide: true},
		}},
		{Name: "attention", Summary: "An item that needs a person: what it is, why, and what they can do. Fail is announced as an alert.", States: []CatalogueState{
			{Label: "Waiting, with actions", Data: Attention{Status: StatusWait, Title: "Two plans wait for you", Body: "Factory will not build them until someone approves.", Action: review, Actions: []Button{{Label: "Snooze", Href: "#attention"}}}},
			{Label: "Failed", Data: Attention{Status: StatusFail, Title: "The gate failed on main", Body: "Three tests failed after the last merge.", Action: &Button{Label: "Open the run", Href: "#attention", Variant: "danger"}}},
			{Label: "Info, no action", Data: Attention{Status: StatusInfo, Title: "Maintenance tonight at 22:00 UTC"}},
			{Label: "Long text", Data: Attention{Status: StatusWait, Title: "A plan for kaas-team/factory waits for your review before it can build", Body: longText, Action: review}},
		}},
		{Name: "approval", Summary: "A decision pinned to a commit. Each button is its own POST form carrying the CSRF token and the SHA.", States: []CatalogueState{
			{Label: "Waiting", Data: Approval{Title: "Plan for #12", Summary: "Adds the chip partial.", ApproveURL: "#approval", RejectURL: "#approval", CSRF: "sample", SHA: "3e1ba70", DiffURL: "#approval"}},
			{Label: "Decided", Data: Approval{Title: "Plan for #11", Summary: "Moves CI off Node 20.", SHA: "a7d9da8", DiffURL: "#approval", Decided: "Approved by ana at 09:14"}},
			{Label: "No commit yet", Data: Approval{Title: "Scope for #14", ApproveURL: "#approval", RejectURL: "#approval"}},
			{Label: "Long text", Data: Approval{Title: "Plan for #15: rework the token pipeline so every consumer reads one generated file", Summary: longText, ApproveURL: "#approval", RejectURL: "#approval", SHA: "0123456789abcdef0123456789abcdef01234567", DiffURL: "#approval"}},
		}},
		{Name: "timeline", Summary: "Events oldest first. Status colours the marker only.", States: []CatalogueState{
			{Label: "Events", Data: Timeline{Label: "Run 42", Events: manyEvents[:3]}},
			{Label: "Empty", Data: Timeline{Label: "Run 43"}},
			{Label: "Long text", Data: Timeline{Label: "Run 44", Events: []TimelineEvent{{Time: "09:00", Title: "Review asked for changes on the plan", Body: longText, Status: StatusWait}}}},
			{Label: "Many events", Data: Timeline{Label: "Run 45", Events: manyEvents}},
		}},
		{Name: "form", Summary: "Labelled fields and one submit button. Help and errors are tied to their field with aria-describedby.", States: []CatalogueState{
			{Label: "Default", Data: Form{ID: "settings", Action: "#form", CSRF: "sample", Submit: Button{Label: "Save", Type: "submit", Variant: "primary"}, Fields: []Field{
				{Name: "repo", Label: "Repository", Required: true, Placeholder: "owner/name"},
				{Name: "notes", Label: "Notes", Type: "textarea", Help: "Shown on the plan."},
				{Name: "tier", Label: "Tier", Type: "select", Value: "b", Options: []string{"a", "b"}},
			}}},
			{Label: "Errors", Data: Form{ID: "invalid", Action: "#form", Error: "Two fields need a fix before this saves.", Submit: Button{Label: "Save", Type: "submit", Variant: "primary"}, Fields: []Field{
				{Name: "repo", Label: "Repository", Required: true, Value: "factory", Error: "Use owner/name, for example kaas-team/factory."},
				{Name: "budget", Label: "Monthly budget", Type: "number", Value: "-5", Help: "In euros.", Error: "Enter zero or more."},
			}}},
			{Label: "Long text", Data: Form{ID: "long", Action: "#form", Submit: Button{Label: "Send the request to the platform team", Type: "submit"}, Fields: []Field{
				{Name: "why", Label: "Why this repository needs a higher tier than the one it has today", Help: longText},
			}}},
		}},
		{Name: "button", Summary: "A link when it has an href, otherwise a button. A disabled link becomes a disabled button.", States: []CatalogueState{
			{Label: "Variants", Many: []any{Button{Label: "Default"}, Button{Label: "Primary", Variant: "primary"}, Button{Label: "Danger", Variant: "danger"}, Button{Label: "Link", Href: "#button"}}},
			{Label: "Disabled", Many: []any{Button{Label: "Disabled", Disabled: true}, Button{Label: "Disabled link", Href: "#button", Disabled: true, Variant: "primary"}}},
			{Label: "Long label", Data: Button{Label: "Approve every waiting plan in this repository at once"}},
		}},
		{Name: "stat", Summary: "One number with its label and an optional change.", States: []CatalogueState{
			{Label: "Statuses", Many: []any{Stat{Label: "Runs today", Value: "12", Delta: "+3", Status: StatusPass}, Stat{Label: "Waiting", Value: "2", Status: StatusWait}, Stat{Label: "Failed", Value: "1", Delta: "+1", Status: StatusFail}, Stat{Label: "Median run", Value: "4m 10s"}}, Wide: true},
			{Label: "No value yet", Data: Stat{Label: "Plans this month"}},
			{Label: "Long text", Data: Stat{Label: "Plans approved without a single change requested", Value: "1,284,019", Delta: "+12% on last month"}},
		}},
	}
}
