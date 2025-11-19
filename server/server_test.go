package server

import (
	"cid_retranslator_walk/config"
	"cid_retranslator_walk/metrics"
	"cid_retranslator_walk/queue"
	"context"
	"net"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	cfg := &config.ServerConfig{
		Host: "localhost",
		Port: "8080",
	}
	mockQ := queue.NewMockQueue()
	rules := &config.CIDRules{}

	s := New(cfg, mockQ, rules)

	if s.host != cfg.Host {
		t.Errorf("expected host %s, got %s", cfg.Host, s.host)
	}
	if s.port != cfg.Port {
		t.Errorf("expected port %s, got %s", cfg.Port, s.port)
	}
	if s.queue != mockQ {
		t.Error("expected queue to be set")
	}
}

func TestServer_UpdateDevice(t *testing.T) {
	stats := metrics.New()
	q := queue.New(10, stats)
	s := New(&config.ServerConfig{}, q, &config.CIDRules{})

	deviceID := 1234
	eventData := "test_event"

	s.UpdateDevice(deviceID, eventData)

	// Check internal state
	s.deviceMu.RLock()
	dev, exists := s.devices[deviceID]
	s.deviceMu.RUnlock()

	if !exists {
		t.Fatal("device should exist")
	}
	if dev.LastEvent != eventData {
		t.Errorf("expected event %s, got %s", eventData, dev.LastEvent)
	}
	if len(dev.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(dev.Events))
	}

	// Check channels
	select {
	case d := <-s.GetDeviceUpdatesChannel():
		if d.ID != deviceID {
			t.Errorf("expected device ID %d, got %d", deviceID, d.ID)
		}
	default:
		t.Error("expected device update in channel")
	}

	select {
	case e := <-s.GetEventUpdatesChannel():
		if e.DeviceID != deviceID {
			t.Errorf("expected event device ID %d, got %d", deviceID, e.DeviceID)
		}
	default:
		t.Error("expected global event in channel")
	}
}

func TestServer_GetDevices(t *testing.T) {
	s := New(&config.ServerConfig{}, nil, nil)

	s.UpdateDevice(1, "event1")
	s.UpdateDevice(2, "event2")

	devs := s.GetDevices()
	if len(devs) != 2 {
		t.Errorf("expected 2 devices, got %d", len(devs))
	}
}

func TestServer_handleRequest(t *testing.T) {
	// Mock connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	mockQ := queue.NewMockQueue()

	// Channel to signal that Enqueue was called
	enqueueCalled := make(chan struct{})

	mockQ.EnqueueFunc = func(data queue.SharedData) bool {
		// ChangeAccountNumber adds the terminator back
		expectedPayload := "5010 182100R57516331" + "\x14"
		if string(data.Payload) != expectedPayload {
			t.Errorf("unexpected payload: %q, want %q", string(data.Payload), expectedPayload)
		}
		// Simulate successful processing
		go func() {
			data.ReplyCh <- queue.DeliveryData{Status: true}
		}()
		close(enqueueCalled)
		return true
	}

	rules := &config.CIDRules{
		RequiredPrefix: "5",
		ValidLength:    20,
	}

	s := New(&config.ServerConfig{}, mockQ, rules)
	s.wg.Add(1) // handleRequest calls Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connHandler := &connection{
		conn:   serverConn,
		queue:  mockQ,
		rules:  rules,
		server: s,
	}

	go connHandler.handleRequest(ctx)

	// Send valid message
	// Format: 5 + 1234 + 18 + 2100 + R575 + 16331 + terminator
	msg := "5010 182100R57516331" + "\x14"
	go func() {
		clientConn.Write([]byte(msg))
	}()

	// We expect Enqueue to be called
	select {
	case <-enqueueCalled:
		// Success
	case <-time.After(time.Second):
		t.Error("timeout waiting for Enqueue call")
	}

	// Read ACK
	buf := make([]byte, 1)
	_, err := clientConn.Read(buf)
	if err != nil {
		t.Errorf("failed to read ACK: %v", err)
	}
	if buf[0] != ackByte {
		t.Errorf("expected ACK, got %x", buf[0])
	}
}

func TestExtractDeviceID(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected int
	}{
		{"Valid", "5010 182100R57516331", 2100},
		{"Short", "123", 0},
		{"Invalid Number", "5010 18ABCDR57516331", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractDeviceID([]byte(tt.message)); got != tt.expected {
				t.Errorf("extractDeviceID() = %v, want %v", got, tt.expected)
			}
		})
	}
}
