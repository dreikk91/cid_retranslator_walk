package queue

import (
	"sync"
	"time"
)

// Queue encapsulates the channels used for communication between the server and client.
type Queue struct {
	DataChannel chan SharedData
	closeOnce   sync.Once
	accepted int
	rejected int
	reconnects int
	mu sync.RWMutex
	StartTime   time.Time
	ConnectionStatus bool
}

// SharedData is the data structure sent from the server to the client.
type SharedData struct {
	Payload []byte
	ReplyCh chan DeliveryData
}

// DeliveryData is the data structure for delivery status replies.
type DeliveryData struct {
	Status bool
}

// New creates and initializes a new Queue.
func New(bufferSize int) *Queue {
	return &Queue{
		DataChannel: make(chan SharedData, bufferSize),
	}
}

// Close closes the channels in the queue.
func (q *Queue) Close() {
	q.closeOnce.Do(func() {
		close(q.DataChannel)
	})
}

func (q *Queue) UpdateStartTime() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.StartTime = time.Now()
}

func (q *Queue) IncrementAccepted() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.accepted ++
}

func (q *Queue) IncrementReconnects() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.reconnects ++
}

func (q *Queue) IncrementRejected() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.rejected ++
}

func (q *Queue) Stats() (accepted, rejected, reconnects int, uptime time.Duration) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	up := time.Since(q.StartTime).Round(time.Second)
	return q.accepted, q.rejected, q.reconnects, up
}

func (q *Queue) SetConnectionStatus(status bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.ConnectionStatus = status
}

func (q *Queue) GetConnectionStatus() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.ConnectionStatus
	
}


