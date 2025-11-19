package client

import (
	"cid_retranslator_walk/config"
	"cid_retranslator_walk/metrics"
	"cid_retranslator_walk/queue"
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

const (
	ackByte         = 0x06
	nackByte        = 0x15
	replyTimeout    = 10 * time.Second
	writeTimeout    = 10 * time.Second
	shutdownTimeout = 5 * time.Second
)

// MessageProvider defines the interface for consuming messages
type MessageProvider interface {
	Events() <-chan queue.SharedData
	GetMetrics() *metrics.Stats
}

type Client struct {
	host             string
	port             string
	conn             net.Conn
	queue            MessageProvider
	reconnectInitial time.Duration
	reconnectMax     time.Duration
	cancel           context.CancelFunc
	stopOnce         sync.Once
	metrics          *metrics.Stats
}

func New(cfg *config.ClientConfig, q MessageProvider) *Client {
	return &Client{
		host:             cfg.Host,
		port:             cfg.Port,
		queue:            q,
		reconnectInitial: cfg.ReconnectInitial,
		reconnectMax:     cfg.ReconnectMax,
		metrics:          q.GetMetrics(),
	}
}

// GetQueueStats повертає канал зі статистикою
func (c *Client) GetQueueStats() <-chan metrics.Snapshot {
	ch := make(chan metrics.Snapshot, 1)
	go func() {
		ch <- c.metrics.Snapshot()
		close(ch)
	}()
	return ch
}

// Run запускає клієнта з автоматичним перепідключенням
func (c *Client) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("Panic in client run loop", "panic", r)
			}
		}()

		delay := c.reconnectInitial
		reconnectAttempts := 0

		for {
			select {
			case <-ctx.Done():
				slog.Info("Client stopping due to context cancellation")
				return
			default:
			}

			conn, err := c.dial(ctx)
			if err != nil {
				c.metrics.SetConnected(false)
				reconnectAttempts++
				c.metrics.IncrementReconnects()

				logLevel := slog.LevelError
				if reconnectAttempts > 10 {
					logLevel = slog.LevelWarn
				}

				slog.Log(ctx, logLevel, "Dial failed, retrying",
					"attempt", reconnectAttempts,
					"delay", delay,
					"error", err)

				select {
				case <-time.After(delay):
					delay = c.calculateNextDelay(delay)
				case <-ctx.Done():
					return
				}
				continue
			}

			slog.Info("Connected to target", "target", c.target())
			reconnectAttempts = 0
			c.conn = conn
			c.metrics.SetConnected(true)

			// handleConnection блокується до втрати з'єднання
			c.handleConnection(ctx, conn)

			conn.Close()
			c.conn = nil
			c.metrics.SetConnected(false)
			delay = c.reconnectInitial

			slog.Info("Connection closed, reconnecting...")
		}
	}()

	<-ctx.Done()
	slog.Info("Client run loop stopped")
}

// dial встановлює з'єднання з таймаутом
func (c *Client) dial(ctx context.Context) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}
	return dialer.DialContext(ctx, "tcp", c.target())
}

// calculateNextDelay обчислює наступну затримку з exponential backoff
func (c *Client) calculateNextDelay(current time.Duration) time.Duration {
	next := current * 2
	if next > c.reconnectMax {
		return c.reconnectMax
	}
	return next
}

// target повертає адресу цілі
func (c *Client) target() string {
	return net.JoinHostPort(c.host, c.port)
}

// Stop зупиняє клієнта gracefully
func (c *Client) Stop() {
	c.stopOnce.Do(func() {
		if c.cancel != nil {
			slog.Info("Stopping client...")
			c.cancel()

			if c.conn != nil {
				c.conn.Close()
			}

			slog.Info("Client stopped")
		}
	})
}

// handleConnection обробляє одне з'єднання до його закриття
func (c *Client) handleConnection(ctx context.Context, conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic in handleConnection", "panic", r)
		}
	}()

	for {
		select {
		case data, ok := <-c.queue.Events():
			if !ok {
				slog.Info("DataChannel closed, stopping connection handler")
				return
			}

			if err := c.processMessage(ctx, conn, data); err != nil {
				slog.Error("Failed to process message", "error", err)
				return
			}

		case <-ctx.Done():
			slog.Info("Stopping connection handler due to shutdown")
			return
		}
	}
}

// processMessage обробляє одне повідомлення
func (c *Client) processMessage(ctx context.Context, conn net.Conn, data queue.SharedData) error {
	// Встановлюємо дедлайн на запис
	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	// Відправляємо дані
	if _, err := conn.Write(data.Payload); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	slog.Debug("Wrote to server", "length", len(data.Payload))

	// Чекаємо відповідь
	reply, err := c.readReply(conn)
	if err != nil {
		return fmt.Errorf("read reply failed: %w", err)
	}

	// Обробляємо відповідь
	status := c.parseReply(reply)

	if status {
		c.metrics.IncrementAccepted()
		slog.Debug("Received ACK from server")
	} else {
		c.metrics.IncrementRejected()
		slog.Debug("Received NACK from server")
	}

	// Відправляємо статус назад
	select {
	case data.ReplyCh <- queue.DeliveryData{Status: status}:
		close(data.ReplyCh)
	case <-time.After(replyTimeout):
		slog.Warn("Timeout sending reply to server handler")
	}

	return nil
}

// readReply читає відповідь від сервера з таймаутом
func (c *Client) readReply(conn net.Conn) ([]byte, error) {
	if err := conn.SetReadDeadline(time.Now().Add(replyTimeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	reply := make([]byte, 1024)
	n, err := conn.Read(reply)
	if err != nil {
		return nil, err
	}

	return reply[:n], nil
}

// parseReply визначає статус відповіді (ACK/NACK)
func (c *Client) parseReply(reply []byte) bool {
	if len(reply) == 0 {
		return false
	}
	return reply[0] == ackByte
}
