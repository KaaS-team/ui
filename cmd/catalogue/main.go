// Command catalogue serves the KaaS UI catalogue: the CSS at /_ui/ and every partial at /_ui/catalogue.
//
//	go run ./cmd/catalogue -addr 127.0.0.1:8080
//
// It prints the page URL on its first line once it is listening, so scripts can wait for it.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/KaaS-team/ui"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "address to listen on; port 0 picks a free one")
	flag.Parse()
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("http://%s%s\n", ln.Addr(), ui.CataloguePath)
	srv := &http.Server{Handler: mux(), ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(srv.Serve(ln))
}

func mux() *http.ServeMux {
	m := http.NewServeMux()
	m.Handle(ui.Prefix, ui.Handler())
	m.Handle(ui.CataloguePath, ui.Catalogue())
	m.Handle(ui.CataloguePath+"/", ui.Catalogue())
	m.Handle("/{$}", http.RedirectHandler(ui.CataloguePath, http.StatusFound))
	return m
}
