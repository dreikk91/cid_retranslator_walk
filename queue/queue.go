package queue

import (
	"cid_retranslator_walk/metrics"
	"sync"
	"time"
)

// Queue інкапсулює канали для комунікації між сервером і клієнтом
type Queue struct {
	DataChannel chan SharedData
	closeOnce   sync.Once
	metrics     *metrics.Stats
}

// SharedData - структура даних від сервера до клієнта
type SharedData struct {
	Payload []byte
	ReplyCh chan DeliveryData
}

// DeliveryData - структура відповіді про статус доставки
type DeliveryData struct {
	Status bool
}

// New створює та ініціалізує нову чергу
func New(bufferSize int, stats *metrics.Stats) *Queue {
	if stats == nil {
		stats = metrics.New()
	}

	return &Queue{
		DataChannel: make(chan SharedData, bufferSize),
		metrics:     stats,
	}
}

// Close закриває канали черги (можна викликати безпечно кілька разів)
func (q *Queue) Close() {
	q.closeOnce.Do(func() {
		close(q.DataChannel)
	})
}

// UpdateStartTime оновлює час старту (для обчислення uptime)
func (q *Queue) UpdateStartTime() {
	q.metrics.Reset()
}

// IncrementAccepted збільшує лічильник прийнятих повідомлень
func (q *Queue) IncrementAccepted() {
	q.metrics.IncrementAccepted()
}

// IncrementReconnects збільшує лічильник перепідключень
func (q *Queue) IncrementReconnects() {
	q.metrics.IncrementReconnects()
}

// IncrementRejected збільшує лічильник відхилених повідомлень
func (q *Queue) IncrementRejected() {
	q.metrics.IncrementRejected()
}

// Stats повертає статистику черги
func (q *Queue) Stats() (accepted, rejected, reconnects int, uptime time.Duration) {
	snap := q.metrics.Snapshot()
	return int(snap.Accepted), int(snap.Rejected), int(snap.Reconnects), snap.Uptime
}

// SetConnectionStatus встановлює статус підключення
func (q *Queue) SetConnectionStatus(status bool) {
	q.metrics.SetConnected(status)
}

// GetConnectionStatus повертає статус підключення
func (q *Queue) GetConnectionStatus() bool {
	return q.metrics.IsConnected()
}

// GetMetrics повертає посилання на метрики (для прямого доступу)
func (q *Queue) GetMetrics() *metrics.Stats {
	return q.metrics
}

// Enqueue додає дані в чергу (non-blocking). Повертає true, якщо успішно, false якщо черга повна.
func (q *Queue) Enqueue(data SharedData) bool {
	select {
	case q.DataChannel <- data:
		return true
	default:
		return false
	}
}

// Events повертає канал для читання подій (receive-only)
func (q *Queue) Events() <-chan SharedData {
	return q.DataChannel
}
