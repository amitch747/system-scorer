package collector

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"
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

var SharedUserCount float64

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

	usernameUID := make(map[string]string)
	sessionSet := make(map[string]struct{})
	userSessionCount := make(map[string]int)

	// Read all processes
	proc, err := os.ReadDir("/proc")
	if err != nil {
		return
	}

	for _, entry := range proc {
		// Ensure entry is a directory
		if !entry.IsDir() {
			continue
		}
		pid := entry.Name()
		// Ensure entry is a numerical directory
		if _, err := strconv.Atoi(pid); err != nil {
			continue
		}

		// Find uid from /proc/pid/status
		uid := ReadUID(pid)
		if uid == "" {
			continue
		}

		// Filter for actual users with /run/user
		stat, err := os.Stat(filepath.Join("/run/user", uid))
		if err != nil || !stat.IsDir() {
			continue
		}

		// At this point we know we have a user
		// Find username
		username, ok := usernameUID[uid]
		if !ok {
			if userObj, err := user.LookupId(uid); err == nil {
				username = userObj.Username
				usernameUID[uid] = username
			} else {
				continue
			}
		}

		// Find ttys
		ttys := ReadTTYs(pid)
		if len(ttys) == 0 {
			continue
		}

		// Find SSH client IP
		ip := readSSHClient(pid)

		// Merge pids from same user sessions
		for _, tty := range ttys {
			key := username + "|" + tty + "|" + ip
			if _, exists := sessionSet[key]; exists {
				// already accounted for, move on
				continue
			}

			sessionSet[key] = struct{}{}
			userSessionCount[username]++

			ch <- prometheus.MustNewConstMetric(
				uc.eachSessionDesc,
				prometheus.GaugeValue,
				1,
				username, ip, tty,
			)
		}
	}

	SharedUserCount = float64(len(userSessionCount))

	for user, count := range userSessionCount {
		ch <- prometheus.MustNewConstMetric(
			uc.userSessionsDesc,
			prometheus.GaugeValue,
			float64(count),
			user,
		)
	}
}

// GetActiveUserCount returns the number of unique users with active TTY sessions
func GetActiveUserCount() float64 {
	usernameUID := make(map[string]string)
	activeUsers := make(map[string]struct{})

	proc, err := os.ReadDir("/proc")
	if err != nil {
		return 0
	}

	for _, entry := range proc {
		if !entry.IsDir() {
			continue
		}
		pid := entry.Name()
		if _, err := strconv.Atoi(pid); err != nil {
			continue
		}

		uid := ReadUID(pid)
		if uid == "" {
			continue
		}

		stat, err := os.Stat(filepath.Join("/run/user", uid))
		if err != nil || !stat.IsDir() {
			continue
		}

		ttys := ReadTTYs(pid)
		if len(ttys) == 0 {
			continue
		}

		username, ok := usernameUID[uid]
		if !ok {
			if userObj, err := user.LookupId(uid); err == nil {
				username = userObj.Username
				usernameUID[uid] = username
			} else {
				continue
			}
		}

		activeUsers[username] = struct{}{}
	}

	return float64(len(activeUsers))
}

func ReadUID(pid string) string {
	data, err := os.ReadFile(filepath.Join("/proc", pid, "status"))
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1]
			}
			break
		}
	}
	return ""
}

func ReadTTYs(pid string) []string {
	fdDir := filepath.Join("/proc", pid, "fd")
	entries, err := os.ReadDir(fdDir)
	if err != nil {
		return nil
	}
	// Mimics the behaviour of a set (Why does Go not have sets)
	seen := make(map[string]struct{})
	var ttys []string

	for _, e := range entries {
		target, err := os.Readlink(filepath.Join(fdDir, e.Name()))
		if err != nil {
			continue
		}
		if strings.Contains(target, "(deleted)") {
			continue
		}

		// Only filter for /dev/pts/ - maybe include /dev/tty*
		if strings.HasPrefix(target, "/dev/pts/") {
			tty := filepath.Base(target)
			label := "pts/" + tty
			if _, ok := seen[label]; !ok {
				seen[label] = struct{}{} // Zero bytes
				ttys = append(ttys, label)
			}
		}
	}
	return ttys
}

func readSSHClient(pid string) string {
	data, err := os.ReadFile(filepath.Join("/proc", pid, "environ"))
	if err != nil {
		return "unknown"
	}
	// proc/pid/environ is null seperated
	for _, v := range strings.Split(string(data), "\x00") {
		if strings.HasPrefix(v, "SSH_CONNECTION=") {
			v := strings.TrimPrefix(v, "SSH_CONNECTION=")
			fields := strings.Fields(v)
			if len(fields) > 1 {
				return fields[0]
			}
		}
	}
	return "unknown"
}
