package collector

import (
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// Needs to implement Describe() & Collect() methods
type SessionScannerCollector struct {
	// Descriptor for the metric (metadata)
	sessionCountDesc *prometheus.Desc
}

func NewSessionScannerCollector() *SessionScannerCollector {
	return &SessionScannerCollector{
		sessionCountDesc: prometheus.NewDesc(
			"sessionscanner_sessions_total",
			"Total number of sessions.",
			nil,
			nil,
		),
	}
}

// Prometheus will call this. Need to feed the info into the channel it will call with
func (ssc SessionScannerCollector) Describe(ch chan<- *prometheus.Desc) {
	// Calls Collector() to figure out what descriptors exist. Less efficient
	//prometheus.DescribeByCollect(ssc, ch)
	ch <- ssc.sessionCountDesc
}

func (ssc SessionScannerCollector) Collect(ch chan<- prometheus.Metric) {
	sessionCount, err := ssc.getUserSessionCount()

	if err != nil {
		return
	}

	ch <- prometheus.MustNewConstMetric(
		ssc.sessionCountDesc,
		prometheus.GaugeValue,
		float64(sessionCount),
	)

}

func (s *SessionScannerCollector) getUserSessionCount() (int, error) {
	cmd := exec.Command("who")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return len(lines), nil
}
