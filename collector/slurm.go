package collector

import (
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type slurmCollector struct {
	slurmStateDesc    *prometheus.Desc
	slurmJobCountDesc *prometheus.Desc
	slurmReservedDesc *prometheus.Desc
}

func NewSlurmCollector() *slurmCollector {
	return &slurmCollector{
		slurmStateDesc: prometheus.NewDesc(
			"syscore_slurm_state_info",
			"Current Slurm node state",
			[]string{"state"},
			nil,
		),
		slurmJobCountDesc: prometheus.NewDesc(
			"syscore_slurm_job_count",
			"Number of active jobs on this node",
			nil,
			nil,
		),
		slurmReservedDesc: prometheus.NewDesc(
			"syscore_slurm_reserved",
			"Binary indicator if node is reserved",
			nil,
			nil,
		),
	}
}

func (sc *slurmCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(sc, ch)
}

func (sc *slurmCollector) Collect(ch chan<- prometheus.Metric) {
	hostname := getShortHostname()

	state := getSlurmNodeState(hostname)

	jobCount := getActiveJobCount(hostname)

	isReserved := getNodeReservationStatus(hostname)

	ch <- prometheus.MustNewConstMetric(
		sc.slurmStateDesc,
		prometheus.GaugeValue,
		1.0,
		state,
	)
	ch <- prometheus.MustNewConstMetric(
		sc.slurmJobCountDesc,
		prometheus.GaugeValue,
		float64(jobCount),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.slurmReservedDesc,
		prometheus.GaugeValue,
		float64(isReserved),
	)

}

func getShortHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return strings.Split(hostname, ".")[0]
}

func getSlurmNodeState(hostname string) string {
	cmd := exec.Command("scontrol", "show", "node", hostname, "-o")
	output, err := cmd.Output()
	if err != nil {
		// Slurm not available or node not in Slurm config
		return "UNKNOWN"
	}

	// State can have modifiers like "IDLE+DRAIN" want the base state
	re := regexp.MustCompile(`State=([A-Z]+)`)
	matches := re.FindSubmatch(output)
	if len(matches) > 1 {
		return string(matches[1])
	}
	return "UNKNOWN"
}

func getActiveJobCount(hostname string) int {
	cmd := exec.Command("squeue", "-w", hostname, "-h", "-o", "%i")
	output, err := cmd.Output()
	if err != nil {
		// squeue failed or no jobs
		return 0
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

func getNodeReservationStatus(hostname string) int64 {
	cmd := exec.Command("scontrol", "show", "reservation", "-o")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	// ensure "node1" doesn't match "node10"
	hostnamePattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(hostname) + `\b`)

	for _, line := range lines {
		// Check if specific hostname appears
		if !hostnamePattern.MatchString(line) {
			continue
		}

		// if !strings.Contains(line, "State=ACTIVE") {
		// 	continue // Node has inactive reservation
		// }

		return 1
	}
	return 0
}
