package shellcmd_test

import (
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/adapters/shellcmd"
)

// TestASilentReviewerStopsBeforeTheTotalDeadline: a reviewer that has produced
// nothing is not a slow reviewer, it is a stopped one. Waiting the whole total
// deadline for it is the defect this clock exists to remove.
func TestASilentReviewerStopsBeforeTheTotalDeadline(t *testing.T) {
	t.Parallel()
	start := time.Now()
	_, _, _, timedOut, err := shellcmd.RunInputWithLimits(".", quietCommand(), "", 5,
		shellcmd.Watchdog{FirstOutput: 120 * time.Millisecond, Idle: 120 * time.Millisecond})
	if err != nil {
		t.Fatalf("RunInputWithLimits: %v", err)
	}
	if !timedOut {
		t.Fatal("a silent reviewer ran to the total deadline")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("elapsed %s: the first-output clock did not stop it early", elapsed)
	}
}

// TestAStreamingReviewerSurvivesTheFirstOutputBound: a model that is slow but
// still writing is alive. Killing it at the first-output bound loses a paid call
// that would have answered.
func TestAStreamingReviewerSurvivesTheFirstOutputBound(t *testing.T) {
	t.Parallel()
	_, _, _, timedOut, err := shellcmd.RunInputWithLimits(".", chattyCommand(), "", 15,
		shellcmd.Watchdog{FirstOutput: 3 * time.Second, Idle: 2 * time.Second})
	if err != nil {
		t.Fatalf("RunInputWithLimits: %v", err)
	}
	if timedOut {
		t.Error("a reviewer that kept writing was killed")
	}
}

// TestAnIdleReviewerStopsAtTheIdleBound: the first line re-arms the clock, but
// the silence after it still counts. A model that answers once and stalls must
// not be mistaken for one that is still working.
func TestAnIdleReviewerStopsAtTheIdleBound(t *testing.T) {
	t.Parallel()
	_, _, _, timedOut, err := shellcmd.RunInputWithLimits(".", quietAfterFirstCommand(), "", 5,
		shellcmd.Watchdog{FirstOutput: 100 * time.Millisecond, Idle: 150 * time.Millisecond})
	if err != nil {
		t.Fatalf("RunInputWithLimits: %v", err)
	}
	if !timedOut {
		t.Error("a reviewer that went quiet after its first line ran to the total deadline")
	}
}

// TestASilentReviewerNamesTheSilenceClockAsTheReasonItStopped: a boolean says a
// run was cut and nothing about which clock cut it. The two are different facts:
// silence says the model stopped answering, and the deadline says the work did
// not fit its budget. Only the adapter watched both clocks, so only it can name
// the one that fired.
func TestASilentReviewerNamesTheSilenceClockAsTheReasonItStopped(t *testing.T) {
	t.Parallel()
	_, _, _, stop, err := shellcmd.RunInputStopped(".", quietCommand(), "", 20,
		100*time.Millisecond, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("RunInputStopped: %v", err)
	}
	if stop != shellcmd.StopSilence {
		t.Errorf("stop = %v, want the silence clock", stop)
	}
}

// TestASlowReviewerNamesTheTotalDeadlineAsTheReasonItStopped: with no silence
// clocks the total deadline is the only one left, and the run says so. Reporting
// it as silence would send the human to look at a model that was still working.
func TestASlowReviewerNamesTheTotalDeadlineAsTheReasonItStopped(t *testing.T) {
	t.Parallel()
	_, _, _, stop, err := shellcmd.RunInputStopped(".", quietCommand(), "", 1, 0, 0)
	if err != nil {
		t.Fatalf("RunInputStopped: %v", err)
	}
	if stop != shellcmd.StopDeadline {
		t.Errorf("stop = %v, want the total deadline", stop)
	}
}
