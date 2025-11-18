package models

import (
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
func (s *StatsData) Update(accepted, rejected, reconnects int, uptime string, status bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.Accepted = accepted
	s.Rejected = rejected
	s.Reconnects = reconnects
	s.Uptime = uptime
	s.ConnectionStatus = status
}

// Get повертає копію статистики потокобезпечно
func (s *StatsData) Get() (int, int, int, string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.Accepted, s.Rejected, s.Reconnects, s.Uptime, s.ConnectionStatus
}