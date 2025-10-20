package main

import (
	_ "encoding/json"

	"log"
	"net/http"

	"github.com/amitch747/prometheus-system-scraper/system-scraper/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Create non-global registry
	reg := prometheus.NewRegistry()
	// Create new metrics and register
	reg.MustRegister(collector.NewSessionScannerCollector())

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	log.Fatal(http.ListenAndServe((":8081"), mux))
}
