package ui

import "html/template"

// Status is the meaning a chip, stat, log line or timeline event carries. The empty Status is neutral.
type Status string

// The four statuses, each drawn with the --ui-<status> and --ui-<status>-bg tokens.
const (
	StatusPass Status = "pass"
	StatusWait Status = "wait"
	StatusFail Status = "fail"
	StatusInfo Status = "info"
)

// Head is the data for "head": the <head> element of every page.
type Head struct {
	Title       string   // the <title>; required
	Description string   // meta description; omitted when empty
	Styles      []string // extra stylesheet hrefs, linked after ui.css
}

// Shell is the data for "shell": a whole page, the nav rail beside the main column.
type Shell struct {
	Head Head
	Nav  Nav
	Body template.HTML // the rendered main column; the caller is responsible for its safety
}

// Nav is the data for "nav": the product name and its links.
type Nav struct {
	Brand string
	Items []NavItem
	Label string // accessible name of the nav landmark; defaults to "Main"
}

// NavItem is one nav link. Current marks the page being shown (aria-current="page").
type NavItem struct {
	Label   string
	Href    string
	Current bool
}

// Chip is the data for "chip": a short status label.
type Chip struct {
	Label  string
	Status Status
}

// Table is the data for "table". With no Rows it renders Empty instead of an empty table.
// Loading draws the header with one busy row; Error replaces the rows with an alert.
type Table struct {
	Caption string
	Columns []string
	Rows    [][]string // one string per column
	Empty   Empty
	Loading bool
	Error   string
}

// Empty is the data for "empty": what a list says when it has nothing yet, and the one action that fills it.
type Empty struct {
	Title  string
	Body   string
	Action *Button // optional
}

// Log is the data for "log": run output, one line each. It is an ARIA log region, so screen readers
// announce lines appended while the page is open. Live adds aria-busy while the run still writes.
type Log struct {
	Title string
	Lines []LogLine
	Live  bool
	Empty string // shown when Lines is empty; defaults to "No output yet."
}

// LogLine is one line of a Log. Status colours the line; the empty Status is plain.
type LogLine struct {
	Time   string
	Status Status
	Text   string
}

// Attention is the data for "attention": an item that needs a person. Title says what, Body says why,
// and Action or Actions are what they can do about it. A fail status is announced as an alert.
type Attention struct {
	Status  Status
	Title   string
	Body    string
	Action  *Button  // optional
	Actions []Button // optional, drawn after Action
}

// Approval is the data for "approval": a decision a person approves or rejects with a POST.
type Approval struct {
	Title      string
	Summary    string
	ApproveURL string
	RejectURL  string
	CSRF       string // sent as the csrf field; omitted when empty
	SHA        string // the commit being approved; shown and sent as the sha field when set
	DiffURL    string // link to the diff; omitted when empty
	Decided    string // set once decided ("Approved by ana"); replaces the buttons
}

// Timeline is the data for "timeline": events oldest first. With no Events it shows Empty.
type Timeline struct {
	Label  string // accessible name of the list; defaults to "Timeline"
	Events []TimelineEvent
	Empty  string // defaults to "Nothing has happened yet."
}

// TimelineEvent is one entry on a Timeline.
type TimelineEvent struct {
	Time   string
	Title  string
	Body   string
	Status Status
}

// Form is the data for "form". Method defaults to post when empty. ID prefixes every field id
// (default "f"), so two forms on one page keep unique ids. Error is a form-level alert.
type Form struct {
	ID     string
	Error  string
	Action string
	Method string
	CSRF   string // sent as the csrf field; omitted when empty
	Fields []Field
	Submit Button
}

// Field is one labelled input. Type is an input type, or "textarea" or "select" (which uses Options).
type Field struct {
	Name        string
	Label       string
	Type        string
	Value       string
	Placeholder string
	Help        string
	Required    bool
	Options     []string
	Error       string // shown under the field and marks it aria-invalid
}

// Button is the data for "button": a link when Href is set, otherwise a <button>.
// Variant is "primary", "danger" or empty for the default look; Type defaults to button.
type Button struct {
	Label    string
	Href     string
	Type     string
	Variant  string
	Name     string
	Value    string
	Disabled bool // a disabled link is drawn as a disabled button
}

// Stat is the data for "stat": one number with its label and an optional change.
type Stat struct {
	Label  string
	Value  string
	Delta  string
	Status Status
}
