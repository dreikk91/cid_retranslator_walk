package queue

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	bufferSize := 100
	q := New(bufferSize)

	if cap(q.DataChannel) != bufferSize {
		t.Errorf("New() DataChannel capacity = %d, want %d", cap(q.DataChannel), bufferSize)
	}
}

func TestQueue_Close(t *testing.T) {
	q := New(10)

	// Closing once should be fine
	q.Close()

	// Verify channels are closed
	_, ok := <-q.DataChannel
	if ok {
		t.Error("DataChannel should be closed, but it's not")
	}

	// Closing a second time should not panic
	q.Close()
}

func TestQueue_Stats(t *testing.T) {
	q := New(10)
	q.UpdateStartTime()
	time.Sleep(1 * time.Second)

	q.IncrementAccepted()
	q.IncrementAccepted()
	q.IncrementRejected()
	q.IncrementReconnects()

	accepted, rejected, reconnects, uptime := q.Stats()

	if accepted != 2 {
		t.Errorf("Expected accepted to be 2, got %d", accepted)
	}
	if rejected != 1 {
		t.Errorf("Expected rejected to be 1, got %d", rejected)
	}
	if reconnects != 1 {
		t.Errorf("Expected reconnects to be 1, got %d", reconnects)
	}
	if uptime < 1*time.Second {
		t.Errorf("Expected uptime to be at least 1 second, got %s", uptime)
	}
}
