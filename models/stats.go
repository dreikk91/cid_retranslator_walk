package models

import (
	"cid_retranslator_walk/metrics"
	"sync"
)

// StatsData - модель статистики клієнта
type StatsData struct {
	Accepted         int
	Rejected         int
	Uptime           string
	Reconnects       int
	ConnectionStatus bool
	mu               sync.RWMutex
}

// NewStatsData створює нову модель статистики
func NewStatsData() *StatsData {
	return &StatsData{
		Uptime: "00:00:00",
	}
}

// Update оновлює статистику потокобезпечно
func (s *StatsData) Update(snap metrics.Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Accepted = int(snap.Accepted)
	s.Rejected = int(snap.Rejected)
	s.Reconnects = int(snap.Reconnects)
	s.Uptime = snap.UptimeString()
	s.ConnectionStatus = snap.Connected
}

// Get повертає копію статистики потокобезпечно
func (s *StatsData) Get() (int, int, int, string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Accepted, s.Rejected, s.Reconnects, s.Uptime, s.ConnectionStatus
}
