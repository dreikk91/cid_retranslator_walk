package client

import (
	"cid_retranslator_gio/config"
	"cid_retranslator_gio/queue"
	"context"
	"net"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	cfg := &config.ClientConfig{
		Host: "localhost",
		Port: "8080",
	}
	q := queue.New(10)
	c := New(cfg, q)

	if c.host != cfg.Host {
		t.Errorf("Expected host %s, got %s", cfg.Host, c.host)
	}
	if c.port != cfg.Port {
		t.Errorf("Expected port %s, got %s", cfg.Port, c.port)
	}
	if c.queue == nil {
		t.Error("Expected queue to be initialized, got nil")
	}
}

func TestFormatDuration(t *testing.T) {
	d := 1*time.Hour + 2*time.Minute + 3*time.Second
	expected := "01:02:03"
	if got := formatDuration(d); got != expected {
		t.Errorf("Expected duration %s, got %s", expected, got)
	}
}

func TestHandleConnection_ACK(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	q := queue.New(10)
	c := New(&config.ClientConfig{}, q)
	c.conn = clientConn

	go func() {
		// Simulate server reading the payload and sending an ACK
		buf := make([]byte, 1024)
		serverConn.Read(buf)
		serverConn.Write([]byte{0x06})
	}()

	replyCh := make(chan queue.DeliveryData, 1)
	q.DataChannel <- queue.SharedData{
		Payload: []byte("test payload"),
		ReplyCh: replyCh,
	}

	go c.handleConnection(context.Background(), clientConn)

	select {
	case reply := <-replyCh:
		if !reply.Status {
			t.Error("Expected status to be true on ACK")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for reply")
	}
}

func TestClientRunStop(t *testing.T) {
	// Create a listener for the mock server
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer ln.Close()

	// Start the mock server
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return // Listener closed, just exit
		}
		defer conn.Close()
		// Just accept the connection and do nothing
	}()

	addr := ln.Addr().String()
	host, port, _ := net.SplitHostPort(addr)

	cfg := &config.ClientConfig{
		Host:             host,
		Port:             port,
		ReconnectInitial: 10 * time.Millisecond,
	}
	q := queue.New(10)
	c := New(cfg, q)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go c.Run(ctx)

	// Give the client a moment to connect
	time.Sleep(100 * time.Millisecond)

	c.Stop()

	// Give the client a moment to stop
	time.Sleep(100 * time.Millisecond)

	if c.conn != nil {
		// Check if the connection is actually closed
		c.conn.SetReadDeadline(time.Now())
		var one []byte
		if _, err := c.conn.Read(one); err == nil {
			t.Error("Expected connection to be closed")
		}
	}
}

func TestHandleConnection_NACK(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	q := queue.New(10)
	c := New(&config.ClientConfig{}, q)
	c.conn = clientConn

	go func() {
		// Simulate server reading the payload and sending a NACK
		buf := make([]byte, 1024)
		serverConn.Read(buf)
		serverConn.Write([]byte{0x15})
	}()

	replyCh := make(chan queue.DeliveryData, 1)
	q.DataChannel <- queue.SharedData{
		Payload: []byte("test payload"),
		ReplyCh: replyCh,
	}

	go c.handleConnection(context.Background(), clientConn)

	select {
	case reply := <-replyCh:
		if reply.Status {
			t.Error("Expected status to be false on NACK")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for reply")
	}
}