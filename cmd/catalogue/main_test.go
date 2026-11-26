package main

import (
	"net/http/httptest"
	"testing"
)

func TestMuxServesCSSAndCatalogue(t *testing.T) {
	for path, want := range map[string]int{"/_ui/ui.css": 200, "/_ui/catalogue": 200, "/_ui/catalogue/catalogue.css": 200, "/": 302} {
		rec := httptest.NewRecorder()
		mux().ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != want {
			t.Errorf("GET %s: %d, want %d", path, rec.Code, want)
		}
	}
}
