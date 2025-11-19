package queue

import (
	"cid_retranslator_walk/metrics"
)

// MockQueue is a mock implementation of both MessageEnqueuer and MessageProvider
type MockQueue struct {
	EnqueueFunc    func(data SharedData) bool
	EventsFunc     func() <-chan SharedData
	GetMetricsFunc func() *metrics.Stats
	Stats          *metrics.Stats
}

func NewMockQueue() *MockQueue {
	return &MockQueue{
		Stats: metrics.New(),
	}
}

func (m *MockQueue) Enqueue(data SharedData) bool {
	if m.EnqueueFunc != nil {
		return m.EnqueueFunc(data)
	}
	return true
}

func (m *MockQueue) Events() <-chan SharedData {
	if m.EventsFunc != nil {
		return m.EventsFunc()
	}
	return make(chan SharedData)
}

func (m *MockQueue) GetMetrics() *metrics.Stats {
	if m.GetMetricsFunc != nil {
		return m.GetMetricsFunc()
	}
	return m.Stats
}
