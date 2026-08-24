package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"medops/internal/console"
	"medops/internal/ns"
	"medops/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	data := flag.String("data", "medops-data", "data directory for file persistence")
	flag.Parse()

	fs, err := store.NewFileStore(*data)
	if err != nil {
		log.Fatalf("medops: init store: %v", err)
	}
	namespaces := ns.NewWardRegistry()
	if err := namespaces.Load(fs); err != nil {
		log.Fatalf("medops: init namespace: %v", err)
	}
	deps, err := console.WireDeps(fs, namespaces)
	if err != nil {
		log.Fatalf("medops: wire services: %v", err)
	}
	api := console.NewAPI(deps)

	log.Printf("medops listening on %s (data=%s)", *addr, *data)
	server := &http.Server{
		Addr:              *addr,
		Handler:           api.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("medops: serve: %v", err)
	}
}
