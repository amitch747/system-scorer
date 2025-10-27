package collector

import (
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// Replacing functionality of https://github.com/stfsy/prometheus-what-active-users-exporter/tree/main
// Instead of using `w` command, we check `/run/user/`
// The metrics themselves keep identical names to ensure old dashboards still work

type userCollector struct {
	userSessionsDesc *prometheus.Desc
	eachSessionDesc  *prometheus.Desc
}

func NewUserCollector() *userCollector {
	return &userCollector{
		userSessionsDesc: prometheus.NewDesc(
			"what_user_sessions_currently_active",
			"Number of sessions per user",
			[]string{"user"},
			nil,
		),
		eachSessionDesc: prometheus.NewDesc(
			"what_each_session_currently_active",
			"Individual sessions per user",
			[]string{"user", "ip", "tty"},
			nil,
		),
	}
}

// Prometheus will call this. Need to feed the info into the channel it will call with
func (uc *userCollector) Describe(ch chan<- *prometheus.Desc) {
	// Calls Collector() to figure out what descriptors exist
	prometheus.DescribeByCollect(uc, ch)
}

func (uc *userCollector) Collect(ch chan<- prometheus.Metric) {

	//users, err := os.ReadDir("/run/user")

	users, err := parseWCommand()
	if err != nil {
		return
	}

	userTotal := len(users)
	ch <- prometheus.MustNewConstMetric(
		uc.userTotalDesc,
		prometheus.GaugeValue,
		float64(userTotal),
	)

	userSessionCount := countSessionsPerUser(users)

	for user, count := range userSessionCount {
		ch <- prometheus.MustNewConstMetric(
			uc.userSessionCountDesc,
			prometheus.GaugeValue,
			float64(count),
			user,
		)
	}
}

type UserInfo struct {
	Name      string
	LoginDate string
	LoginTime string
}

func parseWCommand() ([]UserInfo, error) {
	cmd := exec.Command("who")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(output), "\n")
	var sessions []UserInfo

	for i := 2; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		if len(fields) == 4 {
			sessions = append(sessions, UserInfo{
				Name:      fields[0],
				LoginDate: fields[2],
				LoginTime: fields[3],
			})
		}
	}
	return sessions, nil
}

func countSessionsPerUser(users []UserInfo) map[string]int {
	m := make(map[string]int)
	for _, user := range users {
		m[user.Name]++
	}
	return m
}

// ls /run/user/ | xargs id -nu
// look into slurm to get start and end time
// squeue (login time)
// idle
// for processes
// tty from /user
