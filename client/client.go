package client

import (
	"cid_retranslator_walk/config"
	"cid_retranslator_walk/queue"
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

type Client struct {
	host             string
	port             string
	conn             net.Conn
	queue            *queue.Queue
	reconnectInitial time.Duration
	reconnectMax     time.Duration
	cancel           context.CancelFunc
	stopOnce         sync.Once
	startTime        time.Time
}

type Stats struct {
	Accepted         int    `json:"accepted"`
	Rejected         int    `json:"rejected"`
	Uptime           string `json:"uptime"`
	Reconnects       int    `json:"reconnects"`
	ConnectionStatus bool   `json:"connectionStatus"`
}

func New(cfg *config.ClientConfig, q *queue.Queue) *Client {
	return &Client{
		host:             cfg.Host,
		port:             cfg.Port,
		queue:            q,
		reconnectInitial: cfg.ReconnectInitial,
		reconnectMax:     cfg.ReconnectMax,
		startTime:        time.Now(),
	}
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// GetQueueStats повертає статистику з черги
func (c *Client) GetQueueStats() <-chan Stats {
	ch := make(chan Stats)
	go func() {
		accepted, rejected, reconnects, _ := c.queue.Stats()
		uptime := time.Since(c.startTime).Truncate(time.Second)
		status := c.queue.GetConnectionStatus()
		stats := Stats{
			Accepted:         accepted,
			Rejected:         rejected,
			Reconnects:       reconnects,
			Uptime:           formatDuration(uptime),
			ConnectionStatus: status,
		}

		ch <- stats
		close(ch)
	}()

	return ch
}

func (c *Client) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	go func() {
		delay := c.reconnectInitial
		reconnectAttempts := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			conn, err := net.Dial("tcp", c.host+":"+c.port)
			if err != nil {
				c.queue.SetConnectionStatus(false)
				reconnectAttempts++
				c.queue.IncrementReconnects()
				logMessage := fmt.Sprintf("Dial failed (attempt %d), retrying in %s", reconnectAttempts, delay)
				if reconnectAttempts > 10 { // After 10 attempts, log as a warning
					slog.Warn(logMessage, "target", c.host+":"+c.port, "error", err)
				} else {
					slog.Error(logMessage, "target", c.host+":"+c.port, "error", err)
				}

				time.Sleep(delay)
				delay *= 2
				if delay > c.reconnectMax {
					delay = c.reconnectMax
				}
				continue
			}

			slog.Info("Connected to target", "target", c.host+":"+c.port)
			reconnectAttempts = 0 // Reset on successful connection
			c.conn = conn
			c.queue.SetConnectionStatus(true)

			// handleConnection blocks until connection is lost or shutdown
			c.handleConnection(ctx, conn)

			conn.Close()
			c.conn = nil
			delay = c.reconnectInitial
			slog.Info("Connection closed, reconnecting...")
		}
	}()

	// Wait for stop signal
	<-ctx.Done()
	slog.Info("Client stopping...")
}

func (c *Client) Stop() {
	c.stopOnce.Do(func() {
		if c.cancel != nil {
			slog.Info("Stopping client...")
			c.cancel()
			if c.conn != nil {
				c.conn.Close()
			}
		}
	})
}

func (c *Client) handleConnection(ctx context.Context, conn net.Conn) {
	for {
		select {
		case data, ok := <-c.queue.DataChannel:
			if !ok {
				slog.Info("DataChannel closed, stopping connection handler.")
				return
			}

			_, err := conn.Write(data.Payload)
			if err != nil {
				slog.Error("Write to server failed", "error", err)
				// Don't close the reply channel, server will timeout
				return // Exit to reconnect
			}
			slog.Debug("Wrote to server", "data", string(data.Payload))

			reply := make([]byte, 1024)
			n, err := conn.Read(reply)
			if err != nil {
				slog.Error("Read from server failed", "error", err)
				// Don't close the reply channel, server will timeout
				return // Exit to reconnect
			}

			slog.Debug("Reply from server", "reply", string(reply[:n]))
			if reply[0] == 0x06 {
				slog.Info("Received ACK")
				data.ReplyCh <- queue.DeliveryData{Status: true}
				c.queue.IncrementAccepted()
			} else {
				slog.Warn("Received NACK or other non-ACK response")
				data.ReplyCh <- queue.DeliveryData{Status: false}
				c.queue.IncrementRejected()
			}
			close(data.ReplyCh)

		case <-ctx.Done():
			slog.Info("Stopping connection handler due to shutdown signal.")
			return
		}
	}
}
