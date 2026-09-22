package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Minimal dependency-free Prometheus exposition for the platform's standard
// signals: request rate, error rate and latency. Labels are bounded
// (method + normalized path) so cardinality cannot explode.
var buckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5}

type metrics struct {
	mutex    sync.Mutex
	requests map[string]uint64
	errors   map[string]uint64
	hist     map[string][]uint64
	sums     map[string]float64
}

var store = &metrics{
	requests: map[string]uint64{},
	errors:   map[string]uint64{},
	hist:     map[string][]uint64{},
	sums:     map[string]float64{},
}

func normalize(path string) string {
	if path == "/" || strings.HasPrefix(path, "/healthz") || strings.HasPrefix(path, "/readyz") || strings.HasPrefix(path, "/metrics") {
		return path
	}
	return "/other"
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (recorder *statusRecorder) WriteHeader(status int) {
	recorder.status = status
	recorder.ResponseWriter.WriteHeader(status)
}

func instrumented(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		elapsed := time.Since(start).Seconds()
		route := normalize(r.URL.Path)
		key := r.Method + " " + route
		code := strconv.Itoa(recorder.status)
		store.mutex.Lock()
		defer store.mutex.Unlock()
		store.requests[key+" "+code]++
		if recorder.status >= 500 {
			store.errors[key]++
		}
		counts, ok := store.hist[key]
		if !ok {
			counts = make([]uint64, len(buckets)+1)
			store.hist[key] = counts
		}
		placed := false
		for i, bound := range buckets {
			if elapsed <= bound {
				counts[i]++
				placed = true
				break
			}
		}
		if !placed {
			counts[len(buckets)]++
		}
		store.sums[key] += elapsed
	})
}

func exposition() string {
	var builder strings.Builder
	builder.WriteString("# TYPE http_requests_total counter\n")
	keys := make([]string, 0, len(store.requests))
	for key := range store.requests {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		parts := strings.Split(key, " ")
		fmt.Fprintf(&builder, "http_requests_total{method=%q,path=%q,code=%q} %d\n",
			parts[0], parts[1], parts[2], store.requests[key])
	}
	builder.WriteString("# TYPE http_request_errors_total counter\n")
	errorKeys := make([]string, 0, len(store.errors))
	for key := range store.errors {
		errorKeys = append(errorKeys, key)
	}
	sort.Strings(errorKeys)
	for _, key := range errorKeys {
		parts := strings.Split(key, " ")
		fmt.Fprintf(&builder, "http_request_errors_total{method=%q,path=%q} %d\n",
			parts[0], parts[1], store.errors[key])
	}
	builder.WriteString("# TYPE http_request_duration_seconds histogram\n")
	histKeys := make([]string, 0, len(store.hist))
	for key := range store.hist {
		histKeys = append(histKeys, key)
	}
	sort.Strings(histKeys)
	for _, key := range histKeys {
		parts := strings.Split(key, " ")
		cumulative := uint64(0)
		for i, bound := range buckets {
			cumulative += store.hist[key][i]
			fmt.Fprintf(&builder, "http_request_duration_seconds_bucket{method=%q,path=%q,le=%q} %d\n",
				parts[0], parts[1], strconv.FormatFloat(bound, 'g', -1, 64), cumulative)
		}
		cumulative += store.hist[key][len(buckets)]
		fmt.Fprintf(&builder, "http_request_duration_seconds_bucket{method=%q,path=%q,le=\"+Inf\"} %d\n",
			parts[0], parts[1], cumulative)
		fmt.Fprintf(&builder, "http_request_duration_seconds_sum{method=%q,path=%q} %g\n",
			parts[0], parts[1], store.sums[key])
		fmt.Fprintf(&builder, "http_request_duration_seconds_count{method=%q,path=%q} %d\n",
			parts[0], parts[1], cumulative)
	}
	return builder.String()
}

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
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
		fmt.Fprint(w, exposition())
	})
	return instrumented(mux)
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
