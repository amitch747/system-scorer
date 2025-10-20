package main

import (
	_ "encoding/json"

	"log"
	"net/http"

	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metrics struct {
	info         *prometheus.GaugeVec
	userSessions prometheus.Gauge
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		info: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "system-scraper",
			Name:      "info",
			Help:      "Info about env",
		},
			[]string{"version"}),
		userSessions: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "system-scraper",
			Name:      "user_sessions",
			Help:      "Count of user sessions.",
		}),
	}
	reg.MustRegister(m.info, m.userSessions)
	return m
}

var version string

func init() {
	version = "1.0.0"
}

func main() {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.info.With(prometheus.Labels{"version": version}).Set(1)

	sessionCount, err := getUserSessionCount()
	if err == nil {
		m.userSessions.Set(float64(sessionCount))
	}

	pMux := http.NewServeMux()
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	pMux.Handle("/metrics", promHandler)

	go func() {
		log.Fatal(http.ListenAndServe(":8081", pMux))
	}()

	select {}
}

func getUserSessionCount() (int, error) {
	cmd := exec.Command("who")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return len(lines), nil
}
