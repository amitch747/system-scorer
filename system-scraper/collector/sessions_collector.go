package collector

import (
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// Needs to implement Describe() & Collect() methods
type SessionScannerCollector struct {
	// Descriptor for the metric (metadata)
	sessionTotalDesc *prometheus.Desc
}

// Things to collect
/*
	- Total sessions
	- Sessions per user
	- jcpu
*/

type SessionInfo struct {
	User    string
	LoginAt string
	Idle    string
	Jcpu    string
	Pcpu    string
}

func NewSessionScannerCollector() *SessionScannerCollector {
	return &SessionScannerCollector{
		sessionTotalDesc: prometheus.NewDesc(
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
	ch <- ssc.sessionTotalDesc
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
	println(sessionCountPerUser)

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
	// Count number of structs with same [0]
	m := make(map[string]int)
	for _, session := range sessions {
		m[session.User]++
	}

	return m
}

// func countJCPUPerUser() {

// }

// func countPCPUPerUser() {

// }

// func countIdlePerUser() {

// }

// func countDurationPerUser() {

// }
