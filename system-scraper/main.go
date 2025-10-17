package main

import (
	"encoding/json"

	"net/http"
	"log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Device struct {
	ID int `json:"id"`
	Mac string `json:"mac"`
	Firmware string `json:"firmware"`
}

type metrics struct {
	devices prometheus.Gauge
	info *prometheus.GaugeVec
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		devices: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "system-scraper",
			Name: "connected_devices",
			Help: "Number of currently connected devices.",
		}),
		info: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "system-scraper",
			Name: "info",
			Help: "Info about env",
		},
			[]string{"version"}),
	}
	reg.MustRegister(m.devices, m.info)
	return m
}


var dvs []Device
var version string

func init() {
	version = "1.0.0"
	dvs = []Device{
		{ID: 1, Mac: "00:00:00:00:01", Firmware: "1.0.0"},
		{ID: 2, Mac: "00:00:00:00:02", Firmware: "1.0.1"},
		{ID: 3, Mac: "00:00:00:00:03", Firmware: "1.0.2"},
	}
}

func main() {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.devices.Set(float64(len(dvs)))
	m.info.With(prometheus.Labels{"version": version}).Set(1)

	dMux := http.NewServeMux()
	dMux.HandleFunc("/devices", getDevices)

	pMux := http.NewServeMux()
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	pMux.Handle("/metrics", promHandler)

	go func() {
		log.Fatal(http.ListenAndServe(":8080", dMux))
	}()

	go func() {
		log.Fatal(http.ListenAndServe(":8081", pMux))
	}()

	select {}
}


func getDevices(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(dvs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}