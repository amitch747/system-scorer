package main

import (
	"log"
	"net/http"

	"github.com/amitch747/system-scorer/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Create registry
	reg := prometheus.NewRegistry()

	// Register metrics collectors
	if amdGPUCollector, err := collector.NewAMDGPUCollector(); err == nil {
		reg.MustRegister(amdGPUCollector)
	} else {
		log.Printf("Warning: GPU collector not available: %v", err)
	}
	reg.MustRegister(collector.NewUserCollector())
	reg.MustRegister(collector.NewCPUCollector())
	reg.MustRegister(collector.NewMemCollector())
	reg.MustRegister(collector.NewIoCollector())
	reg.MustRegister(collector.NewNetworkCollector())
	reg.MustRegister(collector.NewSlurmCollector())
	reg.MustRegister(collector.NewScoreCollector())

	// Expose metrics
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe((":8081"), mux))
}
