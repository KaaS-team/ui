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
)

func get(t *testing.T, method, path string) (int, string, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	Catalogue().ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	body, _ := io.ReadAll(rec.Body)
	return rec.Code, rec.Header().Get("Content-Type"), string(body)
}

// The catalogue has a section for every partial, in Names order, so a new partial cannot skip it.
func TestCatalogueCoversEveryPartial(t *testing.T) {
	var got []string
	for _, s := range CatalogueSections() {
		got = append(got, s.Name)
		if s.Name != "shell" && len(s.States) < 2 {
			t.Errorf("%s: %d state(s), want at least two", s.Name, len(s.States))
		}
	}
	if !slices.Equal(got, Names) {
		t.Errorf("catalogue sections %v, Names %v", got, Names)
	}
}

func TestCataloguePage(t *testing.T) {
	for _, path := range []string{CataloguePath, CataloguePath + "/"} {
		code, ctype, body := get(t, "GET", path)
		if code != 200 || !strings.HasPrefix(ctype, "text/html") {
			t.Fatalf("GET %s: %d %s", path, code, ctype)
		}
		for _, bad := range []string{"<no value>", "ZgotmplZ"} {
			if strings.Contains(body, bad) {
				t.Errorf("page contains %s", bad)
			}
		}
		if err := balanced(body); err != "" {
			t.Errorf("page: %s", err)
		}
		for _, name := range Names {
			if !strings.Contains(body, `id="`+name+`"`) {
				t.Errorf("page has no section #%s", name)
			}
		}
		for _, want := range []string{CataloguePath + "/catalogue.css", "Loading…", `role="alert"`, "No output yet.", "Nothing has happened yet."} {
			if !strings.Contains(body, want) {
				t.Errorf("page lacks %q", want)
			}
		}
	}
}

// Ids are unique across the whole page: two forms or two nav samples must not collide.
func TestCatalogueIDsAreUnique(t *testing.T) {
	_, _, body := get(t, "GET", CataloguePath)
	seen := map[string]bool{}
	for _, m := range regexp.MustCompile(`\sid="([^"]+)"`).FindAllStringSubmatch(body, -1) {
		if seen[m[1]] {
			t.Errorf("duplicate id %q", m[1])
		}
		seen[m[1]] = true
	}
	for _, m := range regexp.MustCompile(`(?:for|aria-describedby)="([^"]+)"`).FindAllStringSubmatch(body, -1) {
		for _, id := range strings.Fields(m[1]) {
			if !seen[id] {
				t.Errorf("%s points at missing id %q", m[0], id)
			}
		}
	}
}

func TestCatalogueCSSAndErrors(t *testing.T) {
	code, ctype, body := get(t, "GET", CataloguePath+"/catalogue.css")
	if code != 200 || !strings.HasPrefix(ctype, "text/css") || !strings.Contains(body, ".cat-state") {
		t.Errorf("catalogue.css: %d %s", code, ctype)
	}
	if code, _, _ := get(t, "GET", CataloguePath+"/nope"); code != 404 {
		t.Errorf("unknown path: %d, want 404", code)
	}
	if code, _, _ := get(t, "POST", CataloguePath); code != 405 {
		t.Errorf("POST: %d, want 405", code)
	}
}

// catalogue.css reads only declared tokens, like the files in css/ (test/tokens.test.mjs covers those).
func TestCatalogueCSSUsesKnownTokens(t *testing.T) {
	css, _ := catalogueFiles.ReadFile("catalogue/catalogue.css")
	tokens, _ := os.ReadFile("css/tokens.css")
	for _, m := range regexp.MustCompile(`var\((--[\w-]+)`).FindAllStringSubmatch(string(css), -1) {
		if !strings.Contains(string(tokens), m[1]+":") {
			t.Errorf("catalogue.css uses undeclared %s", m[1])
		}
	}
}

// package.json exports every CSS file and tokens.json, and every export target exists.
func TestPackageExportsCSS(t *testing.T) {
	raw, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatal(err)
	}
	var pkg struct {
		Name         string            `json:"name"`
		Exports      map[string]string `json:"exports"`
		Files        []string          `json:"files"`
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(raw, &pkg); err != nil {
		t.Fatal(err)
	}
	if pkg.Name != "@kaas-team/ui" || pkg.Exports["."] != "./css/ui.css" {
		t.Errorf("package %q exports . as %q", pkg.Name, pkg.Exports["."])
	}
	if len(pkg.Dependencies) > 0 {
		t.Errorf("runtime npm dependencies %v; the package is CSS only", pkg.Dependencies)
	}
	if !slices.Contains(pkg.Files, "css") || !slices.Contains(pkg.Files, "tokens.json") {
		t.Errorf("files %v must ship css and tokens.json", pkg.Files)
	}
	for key, target := range pkg.Exports {
		if strings.Contains(target, "*") {
			continue
		}
		if _, err := os.Stat(target); err != nil {
			t.Errorf("export %s: %v", key, err)
		}
	}
	if pkg.Exports["./css/*"] != "./css/*" {
		t.Errorf("layer files are not exported: no ./css/* entry")
	}
}
