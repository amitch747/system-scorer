package collector

import (
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// Needs to implement Describe() & Collect() methods
type SessionScannerCollector struct {
	sessionTotalDesc        *prometheus.Desc // Descriptors for the metric (metadata)
	sessionCountPerUserDesc *prometheus.Desc
}

func NewSessionScannerCollector() *SessionScannerCollector {
	return &SessionScannerCollector{
		sessionTotalDesc: prometheus.NewDesc(
			"syscraper_session_total",
			"Total number of sessions.",
			nil,
			nil,
		),
		sessionCountPerUserDesc: prometheus.NewDesc(
			"syscraper_session_count_per_user",
			"Number of active sessions per user.",
			[]string{"username"},
			nil,
		),
	}
}

// Prometheus will call this. Need to feed the info into the channel it will call with
func (ssc SessionScannerCollector) Describe(ch chan<- *prometheus.Desc) {
	// Calls Collector() to figure out what descriptors exist
	prometheus.DescribeByCollect(ssc, ch)
}

func (ssc SessionScannerCollector) Collect(ch chan<- prometheus.Metric) {
	sessions, err := parseWCommand()
	if err != nil {
		return
	}

	sessionTotal := len(sessions)
	ch <- prometheus.MustNewConstMetric(
		ssc.sessionTotalDesc,
		prometheus.GaugeValue,
		float64(sessionTotal),
	)

	sessionCountPerUser := countSessionsPerUser(sessions)

	for user, count := range sessionCountPerUser {
		ch <- prometheus.MustNewConstMetric(
			ssc.sessionCountPerUserDesc,
			prometheus.GaugeValue,
			float64(count),
			user,
		)
	}
}

type SessionInfo struct {
	User    string
	LoginAt string
	Idle    string
	Jcpu    string
	Pcpu    string
}

func parseWCommand() ([]SessionInfo, error) {
	cmd := exec.Command("w")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(output), "\n")
	var sessions []SessionInfo

	for i := 2; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		if len(fields) == 8 {
			sessions = append(sessions, SessionInfo{
				User:    fields[0],
				LoginAt: fields[3],
				Idle:    fields[4],
				Jcpu:    fields[5],
				Pcpu:    fields[6],
			})
		}
	}
	return sessions, nil
}

func countSessionsPerUser(sessions []SessionInfo) map[string]int {
	m := make(map[string]int)
	for _, session := range sessions {
		m[session.User]++
	}
	return m
}
