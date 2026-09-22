package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

var requests atomic.Uint64

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"service":"payment-api","status":"ok"}`)
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "ready")
	})
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# TYPE http_requests_total counter\nhttp_requests_total %d\n", requests.Load())
	})
	return mux
}

func main() {
	address := ":8080"
	if value := os.Getenv("LISTEN_ADDRESS"); value != "" {
		address = value
	}
	log.Printf("listening on %s", address)
	if err := http.ListenAndServe(address, handler()); err != nil {
		log.Fatal(err)
	}
}
