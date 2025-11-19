package metrics

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Stats - потокобезпечна статистика додатку
type Stats struct {
	accepted   atomic.Int64
	rejected   atomic.Int64
	reconnects atomic.Int64
	startTime  time.Time
	connected  atomic.Bool
}

// Snapshot - знімок статистики на момент часу
type Snapshot struct {
	Accepted   int64         `json:"accepted"`
	Rejected   int64         `json:"rejected"`
	Reconnects int64         `json:"reconnects"`
	Uptime     time.Duration `json:"uptime"`
	Connected  bool          `json:"connected"`
}

// New створює новий екземпляр статистики
func New() *Stats {
	return &Stats{
		startTime: time.Now(),
	}
}

// IncrementAccepted збільшує лічильник прийнятих повідомлень
func (s *Stats) IncrementAccepted() {
	s.accepted.Add(1)
}

// IncrementRejected збільшує лічильник відхилених повідомлень
func (s *Stats) IncrementRejected() {
	s.rejected.Add(1)
}

// IncrementReconnects збільшує лічильник перепідключень
func (s *Stats) IncrementReconnects() {
	s.reconnects.Add(1)
}

// SetConnected встановлює статус підключення
func (s *Stats) SetConnected(status bool) {
	s.connected.Store(status)
}

// IsConnected повертає поточний статус підключення
func (s *Stats) IsConnected() bool {
	return s.connected.Load()
}

// Reset скидає всі лічильники (окрім startTime)
func (s *Stats) Reset() {
	s.accepted.Store(0)
	s.rejected.Store(0)
	s.reconnects.Store(0)
	s.startTime = time.Now()
}

// Snapshot повертає знімок поточної статистики
func (s *Stats) Snapshot() Snapshot {
	return Snapshot{
		Accepted:   s.accepted.Load(),
		Rejected:   s.rejected.Load(),
		Reconnects: s.reconnects.Load(),
		Uptime:     time.Since(s.startTime),
		Connected:  s.connected.Load(),
	}
}

// UptimeString повертає uptime у форматі HH:MM:SS
func (snap Snapshot) UptimeString() string {
	d := snap.Uptime.Truncate(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// String повертає текстове представлення статистики
func (snap Snapshot) String() string {
	status := "Disconnected"
	if snap.Connected {
		status = "Connected"
	}
	
	return fmt.Sprintf(
		"Status: %s | Accepted: %d | Rejected: %d | Reconnects: %d | Uptime: %s",
		status,
		snap.Accepted,
		snap.Rejected,
		snap.Reconnects,
		snap.UptimeString(),
	)
}

// TotalMessages повертає загальну кількість оброблених повідомлень
func (snap Snapshot) TotalMessages() int64 {
	return snap.Accepted + snap.Rejected
}

// SuccessRate повертає процент успішних повідомлень (0-100)
func (snap Snapshot) SuccessRate() float64 {
	total := snap.TotalMessages()
	if total == 0 {
		return 0
	}
	return float64(snap.Accepted) / float64(total) * 100
}