package server

import (
	"cid_retranslator_gio/config"
	"cid_retranslator_gio/queue"
	"container/ring"
	"fmt"
	"testing"
	"time"
)

// newTestServer creates a server instance for testing.
func newTestServer() *Server {
	cfg := &config.ServerConfig{Host: "localhost", Port: "0"}
	q := queue.New(10)
	rules := &config.CIDRules{}
	return New(cfg, q, rules)
}

func TestExtractDeviceID(t *testing.T) {
	tests := []struct {
		name    string
		message []byte
		want    int
	}{
		{"Valid ID", []byte("...1234..."), 1234},
		{"Invalid ID (non-numeric)", []byte("...abcd..."), 0},
		{"Short message", []byte("...123"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The function expects a message of a specific format.
			// We create a message that fits the expected slice indices.
			var msg []byte
			if len(tt.message) >= 11 {
				msg = make([]byte, len(tt.message))
				copy(msg, tt.message)
			} else {
				msg = make([]byte, 11)
			}
			
			// Manually place the relevant part of the message
			if len(tt.message) >= 11 {
				copy(msg[7:11], tt.message[3:7])
			}


			if got := extractDeviceID(msg); got != tt.want {
				t.Errorf("extractDeviceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateDevice(t *testing.T) {
	s := newTestServer()
	deviceID := 123
	eventData1 := "event1"
	eventData2 := "event2"

	// 1. Test creating a new device
	s.UpdateDevice(deviceID, eventData1)

	s.deviceMu.RLock()
	dev, exists := s.devices[deviceID]
	s.deviceMu.RUnlock()

	if !exists {
		t.Fatal("Device was not created")
	}
	if dev.ID != deviceID {
		t.Errorf("Device ID is incorrect, got %d, want %d", dev.ID, deviceID)
	}
	if dev.LastEvent != eventData1 {
		t.Errorf("Device LastEvent is incorrect, got %s, want %s", dev.LastEvent, eventData1)
	}
	if len(dev.Events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(dev.Events))
	}
	if dev.Events[0].Data != eventData1 {
		t.Errorf("Event data is incorrect, got %s, want %s", dev.Events[0].Data, eventData1)
	}

	// 2. Test adding a second event
	// Sleep to ensure a different timestamp
	time.Sleep(1 * time.Millisecond)
	s.UpdateDevice(deviceID, eventData2)

	s.deviceMu.RLock()
	if len(dev.Events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(dev.Events))
	}
	if dev.LastEvent != eventData2 {
		t.Errorf("Device LastEvent was not updated, got %s, want %s", dev.LastEvent, eventData2)
	}
	if dev.Events[1].Data != eventData2 {
		t.Errorf("Second event data is incorrect, got %s, want %s", dev.Events[1].Data, eventData2)
	}
	s.deviceMu.RUnlock()

	// 3. Test event history trimming
	for i := 0; i < 110; i++ { // Add 110 more events
		s.UpdateDevice(deviceID, fmt.Sprintf("event_%d", i))
	}

	s.deviceMu.RLock()
	if len(dev.Events) != 100 {
		t.Fatalf("Event history should be trimmed to 100, but got %d", len(dev.Events))
	}
	// The last event should be "event_109"
	if dev.Events[99].Data != "event_109" {
		t.Errorf("Last event after trimming is incorrect, got %s", dev.Events[99].Data)
	}
	s.deviceMu.RUnlock()
}

func TestGetDeviceEvents(t *testing.T) {
	s := newTestServer()
	deviceID := 456
	s.UpdateDevice(deviceID, "event1")
	s.UpdateDevice(deviceID, "event2")

	events := s.GetDeviceEvents(deviceID)
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}
	if events[0].Data != "event1" || events[1].Data != "event2" {
		t.Error("GetDeviceEvents returned incorrect data")
	}

	// Test getting events for a non-existent device
	events = s.GetDeviceEvents(999)
	if len(events) != 0 {
		t.Errorf("Expected 0 events for non-existent device, got %d", len(events))
	}
}

func TestGetDevices(t *testing.T) {
	s := newTestServer()
	s.UpdateDevice(102, "eventB")
	s.UpdateDevice(101, "eventA")
	s.UpdateDevice(103, "eventC")

	devicesCh := s.GetDevices()
	
	var receivedDevices []Device
	for dev := range devicesCh {
		receivedDevices = append(receivedDevices, dev)
	}

	if len(receivedDevices) != 3 {
		t.Fatalf("Expected 3 devices, got %d", len(receivedDevices))
	}

	// Check if sorted by ID
	if receivedDevices[0].ID != 101 || receivedDevices[1].ID != 102 || receivedDevices[2].ID != 103 {
		t.Error("Devices are not sorted correctly by ID")
	}
	
	// Check content
	if receivedDevices[0].LastEvent != "eventA" {
		t.Errorf("Incorrect last event for device 101: got %s", receivedDevices[0].LastEvent)
	}
}

func TestGetGlobalEvents(t *testing.T) {
	s := newTestServer()
	s.maxGlobalEvents = 5 // Use a smaller ring for testing
	s.globalEventsRing = s.globalEventsRing.Link(ring.New(s.maxGlobalEvents))


	// Add 7 events to test ring buffer wrapping
	for i := 0; i < 7; i++ {
		s.UpdateDevice(i, fmt.Sprintf("global_event_%d", i))
		time.Sleep(1 * time.Millisecond) // Ensure unique timestamps for sorting
	}

	eventsCh := s.GetGlobalEvents()

	var receivedEvents []GlobalEvent
	for event := range eventsCh {
		receivedEvents = append(receivedEvents, event)
	}

	if len(receivedEvents) != 5 {
		t.Fatalf("Expected 5 global events (due to ring size), got %d", len(receivedEvents))
	}

	// Events should be sorted by time descending (newest first)
	if receivedEvents[0].Data != "global_event_6" {
		t.Errorf("Expected newest event to be first, got %s", receivedEvents[0].Data)
	}
	if receivedEvents[4].Data != "global_event_2" {
		t.Errorf("Expected oldest event to be last, got %s", receivedEvents[4].Data)
	}
}