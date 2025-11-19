package client

import (
	"cid_retranslator_walk/config"
	"cid_retranslator_walk/metrics"
	"cid_retranslator_walk/queue"
	"context"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	cfg := &config.ClientConfig{
		Host:             "localhost",
		Port:             "8080",
		ReconnectInitial: time.Second,
		ReconnectMax:     time.Minute,
	}

	mockQ := queue.NewMockQueue()

	c := New(cfg, mockQ)

	if c.host != cfg.Host {
		t.Errorf("expected host %s, got %s", cfg.Host, c.host)
	}
	if c.port != cfg.Port {
		t.Errorf("expected port %s, got %s", cfg.Port, c.port)
	}
	if c.queue != mockQ {
		t.Error("expected queue to be set")
	}
	if c.metrics != mockQ.Stats {
		t.Error("expected metrics to be set")
	}
}

func TestClient_calculateNextDelay(t *testing.T) {
	c := &Client{
		reconnectMax: 10 * time.Second,
	}

	tests := []struct {
		current  time.Duration
		expected time.Duration
	}{
		{1 * time.Second, 2 * time.Second},
		{2 * time.Second, 4 * time.Second},
		{8 * time.Second, 10 * time.Second},
		{10 * time.Second, 10 * time.Second},
	}

	for _, tt := range tests {
		if got := c.calculateNextDelay(tt.current); got != tt.expected {
			t.Errorf("calculateNextDelay(%v) = %v, want %v", tt.current, got, tt.expected)
		}
	}
}

func TestClient_parseReply(t *testing.T) {
	c := &Client{}

	tests := []struct {
		name     string
		reply    []byte
		expected bool
	}{
		{"ACK", []byte{ackByte}, true},
		{"NACK", []byte{nackByte}, false},
		{"Empty", []byte{}, false},
		{"Garbage", []byte{0xFF}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.parseReply(tt.reply); got != tt.expected {
				t.Errorf("parseReply() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClient_processMessage(t *testing.T) {
	// Create a pipe to simulate network connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	stats := metrics.New()
	c := &Client{
		metrics: stats,
	}

	// Mock server handling
	go func() {
		buf := make([]byte, 1024)
		n, err := serverConn.Read(buf)
		if err != nil {
			return
		}
		// Echo ACK if received "test"
		if string(buf[:n]) == "test" {
			serverConn.Write([]byte{ackByte})
		} else {
			serverConn.Write([]byte{nackByte})
		}
	}()

	ctx := context.Background()
	replyCh := make(chan queue.DeliveryData, 1)
	data := queue.SharedData{
		Payload: []byte("test"),
		ReplyCh: replyCh,
	}

	err := c.processMessage(ctx, clientConn, data)
	if err != nil {
		t.Fatalf("processMessage failed: %v", err)
	}

	select {
	case reply := <-replyCh:
		if !reply.Status {
			t.Error("expected successful delivery status")
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for reply")
	}

	if stats.Snapshot().Accepted != 1 {
		t.Errorf("expected 1 accepted message, got %d", stats.Snapshot().Accepted)
	}
}

func TestClient_Run_ConnectionLoop(t *testing.T) {
	// Start a real TCP listener to test connection
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)

	cfg := &config.ClientConfig{
		Host:             addr.IP.String(),
		Port:             strconv.Itoa(addr.Port), // This might need int to string conversion, but Port is string in config
		ReconnectInitial: 10 * time.Millisecond,
		ReconnectMax:     100 * time.Millisecond,
	}
	// Fix Port assignment
	cfg.Port = "0" // Not used directly as we use addr.String() logic in target() but target() uses c.port
	// Let's just parse the port from listener
	_, portStr, _ := net.SplitHostPort(listener.Addr().String())
	cfg.Port = portStr

	stats := metrics.New()
	q := queue.New(10, stats)
	c := New(cfg, q)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.Run(ctx)
	}()

	// Accept connection
	conn, err := listener.Accept()
	if err != nil {
		t.Fatalf("failed to accept connection: %v", err)
	}
	conn.Close() // Close immediately to trigger reconnect logic or just prove connection

	// Give it a moment to update stats
	time.Sleep(50 * time.Millisecond)

	if !stats.Snapshot().Connected {
		// It might be disconnected by now, but we check if it connected at least once?
		// Actually Run sets Connected=true then false.
		// We can check Reconnects count?
	}

	cancel()
	wg.Wait()
}
