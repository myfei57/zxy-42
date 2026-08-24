package heart_test

import (
	"testing"
	"time"

	"medops/internal/device"
	"medops/internal/heart"
)

// liveDevice is a minimal Beater backed by a real device.Clock so the window
// is exercised exactly as it is in production (Reanchor correction + Advance
// on every accepted beat).
type liveDevice struct {
	dev *device.Device
}

func (l *liveDevice) DeviceID() string              { return l.dev.ID }
func (l *liveDevice) LiveClock() heart.Clock        { return l.dev.Clock }
func (l *liveDevice) LastSeq() int64                { return l.dev.LastSeq() }
func (l *liveDevice) ApplyBeat(b heart.Beat) error { return l.dev.ApplyBeat(b) }

// TestOnlineWindowTracksClockAfterCorrection reproduces the ICU monitor
// incident: a 13:58 clock correction froze the online window on the old
// anchor, so beats from 14:00 onward were rejected as offline even though
// they were live and in sequence.
func TestOnlineWindowTracksClockAfterCorrection(t *testing.T) {
	// 13:00 enrollment; the monitor has been beating every minute.
	enroll := time.Date(2026, 8, 24, 13, 0, 0, 0, time.UTC)
	dev, err := device.NewDevice("SN-ICU-1", "M-1", "监护仪", "H1", "W1", device.ClassificationICU, enroll)
	if err != nil {
		t.Fatal(err)
	}

	svc := heart.NewService(heart.DefaultPolicy().Window(), nil)
	b := &liveDevice{dev: dev}

	// 13:00-13:57: normal per-minute beats, all accepted.
	for m := 0; m <= 57; m++ {
		tick := enroll.Add(time.Duration(m) * time.Minute)
		beat := heart.NewBeat(dev.ID, int64(m+1), tick, tick)
		if _, err := svc.Record(b, beat); err != nil {
			t.Fatalf("beat at %s rejected before correction: %v", tick, err)
		}
	}

	// 13:58: the monitor performs a clock correction. Reanchor moves both
	// the enrollment anchor and the live clock to the corrected time, just
	// like handleClockSync does in production.
	correction := time.Date(2026, 8, 24, 13, 58, 0, 0, time.UTC)
	dev.Clock.Reanchor(correction)

	// 14:00, 14:01, 14:02: beats keep arriving every minute, sequence
	// continues. With the bug, the window stayed frozen on the pre-
	// correction anchor and every beat from 14:01 landed outside it.
	for m, seq := 60, 59; m <= 62; m, seq = m+1, seq+1 {
		tick := enroll.Add(time.Duration(m) * time.Minute) // 14:00, 14:01, 14:02
		beat := heart.NewBeat(dev.ID, int64(seq), tick, tick)
		verdict, err := svc.Record(b, beat)
		if err != nil {
			t.Fatalf("live beat at %s rejected as offline after correction: %v", tick, err)
		}
		if !verdict.Online {
			t.Fatalf("live beat at %s reported offline after correction", tick)
		}
	}
}

// TestOnlineWindowStillRejectsStaleBeatAfterCorrection ensures that anchoring
// the window on Current did not weaken replay protection. After a correction
// and one live beat (which advances Current while Baseline stays put), a
// pre-correction replay is still rejected by RejectLate, and a beat beyond the
// tolerance is still rejected as offline.
func TestOnlineWindowStillRejectsStaleBeatAfterCorrection(t *testing.T) {
	enroll := time.Date(2026, 8, 24, 13, 0, 0, 0, time.UTC)
	dev, err := device.NewDevice("SN-ICU-2", "M-1", "监护仪", "H1", "W1", device.ClassificationICU, enroll)
	if err != nil {
		t.Fatal(err)
	}

	svc := heart.NewService(heart.DefaultPolicy().Window(), nil)
	b := &liveDevice{dev: dev}

	// One accepted beat at 13:01, within tolerance of the enrollment clock.
	tick := enroll.Add(time.Minute)
	if _, err := svc.Record(b, heart.NewBeat(dev.ID, 1, tick, tick)); err != nil {
		t.Fatal(err)
	}

	// Correction to 13:58, then accept the 14:00 beat so the live clock
	// advances past the anchor (Current=14:00, Baseline=13:58).
	dev.Clock.Reanchor(time.Date(2026, 8, 24, 13, 58, 0, 0, time.UTC))
	live := time.Date(2026, 8, 24, 14, 0, 0, 0, time.UTC)
	if _, err := svc.Record(b, heart.NewBeat(dev.ID, 2, live, live)); err != nil {
		t.Fatalf("14:00 beat rejected after correction: %v", err)
	}

	// A pre-correction replay (SentAt before the anchor) is still rejected.
	stale := heart.NewBeat(dev.ID, 3, enroll.Add(10*time.Minute), enroll.Add(10*time.Minute))
	if _, err := svc.Record(b, stale); err != heart.ErrReplayRejected {
		t.Fatalf("expected ErrReplayRejected for pre-correction replay, got %v", err)
	}

	// A beat beyond the tolerance (SentAt far in the future of the live
	// clock) is still rejected as outside the online window.
	future := heart.NewBeat(dev.ID, 4, live.Add(10*time.Minute), live.Add(10*time.Minute))
	if _, err := svc.Record(b, future); err != heart.ErrOfflineWindow {
		t.Fatalf("expected ErrOfflineWindow for future beat, got %v", err)
	}
}
