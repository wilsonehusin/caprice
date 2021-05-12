package scribetag

import (
	"fmt"
	"testing"
	"time"
)

func batchAdd(tracker *ScribeTracker, now time.Time, count int) error {
	var timeDiff time.Duration

	for i := 0; i < count; i++ {
		timeDiff = time.Duration(i) * 5 * time.Second
		if err := tracker.Add(fmt.Sprintf("name-%d", i), now.Add(timeDiff)); err != nil {
			return err
		}
	}
	return nil
}

func expectTag(t *testing.T, tag *ScribeTag, name string) {
	if tag == nil {
		t.Fatalf("expected \"%s\", received nil", name)
	}
	if tag.Name != name {
		t.Fatalf("expected \"%s\", received %s", name, tag.Name)
	}
}

func TestGetOldestNoPulse(t *testing.T) {
	tracker := NewTracker()
	startTime := time.Now()

	if err := batchAdd(tracker, startTime, 3); err != nil {
		t.Fatal(err)
	}

	expectTag(t, tracker.Oldest(), "name-0")
}

func TestWithPulse(t *testing.T) {
	tracker := NewTracker()
	startTime := time.Now()

	if err := batchAdd(tracker, startTime, 5); err != nil {
		t.Fatal(err)
	}

	if err := tracker.Pulse("name-0", startTime.Add(60*time.Second)); err != nil {
		t.Fatal(err)
	}

	expectTag(t, tracker.Oldest(), "name-1")
}

func TestReLink(t *testing.T) {
	tracker := NewTracker()
	startTime := time.Now()

	if err := batchAdd(tracker, startTime, 6); err != nil {
		t.Fatal(err)
	}

	// Roughly this is what's expected:
	// 0-1-2-3-4-5     Oldest() == 0
	// 0---2-3-4-5     Oldest() == 0
	//     2-3-4-5-0   Oldest() == 2
	expectTag(t, tracker.Take("name-1"), "name-1")
	expectTag(t, tracker.Oldest(), "name-0")
	if err := tracker.Pulse("name-0", startTime.Add(60*time.Second)); err != nil {
		t.Fatal(err)
	}
	expectTag(t, tracker.Oldest(), "name-2")
}
