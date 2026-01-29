package optimizer

import "testing"

func TestBuildPollPlanOrder(t *testing.T) {
	plan := BuildPollPlan([]PollItem{
		{Symbol: "A", ProviderScore: 0.9, ConflictRate: 0.0, WindowWeight: 1.0},
		{Symbol: "B", ProviderScore: 0.1, ConflictRate: 0.9, WindowWeight: 0.0},
	}, SchedulerConfig{})
	if len(plan) != 2 {
		t.Fatalf("len=%d", len(plan))
	}
	if plan[0].Symbol != "A" {
		t.Fatalf("expected A first, got %s", plan[0].Symbol)
	}
	if !(plan[0].NextIn < plan[1].NextIn) {
		t.Fatalf("expected A to be polled more frequently")
	}
}

