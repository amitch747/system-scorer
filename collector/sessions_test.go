package collector

import (
	"testing"
)

func TestParseWCommand(t *testing.T) {
	sessions, err := parseWCommand()
	if err != nil {
		t.Fatalf("Failed parseWCommand: %v", err)
	}

	t.Logf("Found %d sessions", len(sessions))
	for _, s := range sessions {
		t.Logf("User: %s, LoginAt: %s Idle: %s, Jcpu: %s, Pcpu: %s",
			s.User, s.LoginAt, s.Idle, s.Jcpu, s.Pcpu)
	}
}

func TestCountSessionsPerUser(t *testing.T) {
	sessions, _ := parseWCommand()
	sessionCountPerUser := countSessionsPerUser(sessions)

	for user, count := range sessionCountPerUser {
		t.Logf("User: %s [%d]", user, count)
	}
}

// t.Logf & t.Fatalf (c style print)
// cd system-scraper/collector
// go test -v
