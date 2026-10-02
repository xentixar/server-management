package models

import (
	"testing"
	"time"

	"api/app/constants"
)

func TestEffectiveAgentStatus(t *testing.T) {
	now := time.Now().UTC()
	recent := now.Add(-20 * time.Second)
	stale := now.Add(-90 * time.Second)
	dead := now.Add(-5 * time.Minute)

	cases := []struct {
		name   string
		server Server
		want   string
	}{
		{name: "pending never seen", server: Server{AgentStatus: constants.AgentStatusPending}, want: constants.AgentStatusPending},
		{name: "online recent", server: Server{AgentStatus: constants.AgentStatusOnline, LastSeenAt: &recent}, want: constants.AgentStatusOnline},
		{name: "degraded", server: Server{AgentStatus: constants.AgentStatusOnline, LastSeenAt: &stale}, want: constants.AgentStatusDegraded},
		{name: "offline", server: Server{AgentStatus: constants.AgentStatusOnline, LastSeenAt: &dead}, want: constants.AgentStatusOffline},
		{name: "was online but never stamped", server: Server{AgentStatus: constants.AgentStatusOnline}, want: constants.AgentStatusOffline},
	}
	for _, tc := range cases {
		if got := tc.server.EffectiveAgentStatus(); got != tc.want {
			t.Fatalf("%s: got %s want %s", tc.name, got, tc.want)
		}
	}
}
