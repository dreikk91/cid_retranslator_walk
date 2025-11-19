package queue

import (
	"cid_retranslator_walk/metrics"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	bufferSize := 100
	stats := metrics.New()
	q := New(bufferSize, stats)

	if cap(q.DataChannel) != bufferSize {
		t.Errorf("New() DataChannel capacity = %d, want %d", cap(q.DataChannel), bufferSize)
	}

	if q.metrics == nil {
		t.Error("New() metrics is nil")
	}
}

func TestNewWithNilMetrics(t *testing.T) {
	bufferSize := 100
	q := New(bufferSize, nil)

	if q.metrics == nil {
		t.Error("New() should create metrics if nil is passed")
	}
}

func TestQueue_Close(t *testing.T) {
	stats := metrics.New()
	q := New(10, stats)

	// Перша спроба закриття
	q.Close()

	// Перевіряємо, що канал закритий
	_, ok := <-q.DataChannel
	if ok {
		t.Error("DataChannel should be closed, but it's not")
	}

	// Друга спроба закриття не повинна паніки
	q.Close()
}

func TestQueue_Stats(t *testing.T) {
	stats := metrics.New()
	q := New(10, stats)
	q.UpdateStartTime()
	
	time.Sleep(1100 * time.Millisecond)

	q.IncrementAccepted()
	q.IncrementAccepted()
	q.IncrementRejected()
	q.IncrementReconnects()

	accepted, rejected, reconnects, uptime := q.Stats()

	if accepted != 2 {
		t.Errorf("Expected accepted to be 2, got %d", accepted)
	}
	if rejected != 1 {
		t.Errorf("Expected rejected to be 1, got %d", rejected)
	}
	if reconnects != 1 {
		t.Errorf("Expected reconnects to be 1, got %d", reconnects)
	}
	if uptime < 1*time.Second {
		t.Errorf("Expected uptime to be at least 1 second, got %s", uptime)
	}
}

func TestQueue_ConnectionStatus(t *testing.T) {
	stats := metrics.New()
	q := New(10, stats)

	// За замовчуванням має бути false
	if q.GetConnectionStatus() {
		t.Error("Initial connection status should be false")
	}

	// Встановлюємо true
	q.SetConnectionStatus(true)
	if !q.GetConnectionStatus() {
		t.Error("Connection status should be true after setting")
	}

	// Встановлюємо false
	q.SetConnectionStatus(false)
	if q.GetConnectionStatus() {
		t.Error("Connection status should be false after setting")
	}
}

func TestQueue_ConcurrentIncrements(t *testing.T) {
	stats := metrics.New()
	q := New(10, stats)

	const goroutines = 100
	const incrementsPerGoroutine = 1000

	done := make(chan struct{})
	
	// Запускаємо горутини для accepted
	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < incrementsPerGoroutine; j++ {
				q.IncrementAccepted()
			}
			done <- struct{}{}
		}()
	}

	// Запускаємо горутини для rejected
	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < incrementsPerGoroutine; j++ {
				q.IncrementRejected()
			}
			done <- struct{}{}
		}()
	}

	// Чекаємо завершення
	for i := 0; i < goroutines*2; i++ {
		<-done
	}

	accepted, rejected, _, _ := q.Stats()

	expectedAccepted := goroutines * incrementsPerGoroutine
	expectedRejected := goroutines * incrementsPerGoroutine

	if accepted != expectedAccepted {
		t.Errorf("Expected accepted to be %d, got %d", expectedAccepted, accepted)
	}

	if rejected != expectedRejected {
		t.Errorf("Expected rejected to be %d, got %d", expectedRejected, rejected)
	}
}

func TestQueue_DataChannel(t *testing.T) {
	stats := metrics.New()
	q := New(2, stats)

	// Відправляємо дані
	data1 := SharedData{
		Payload: []byte("test1"),
		ReplyCh: make(chan DeliveryData, 1),
	}
	data2 := SharedData{
		Payload: []byte("test2"),
		ReplyCh: make(chan DeliveryData, 1),
	}

	q.DataChannel <- data1
	q.DataChannel <- data2

	// Читаємо дані
	received1 := <-q.DataChannel
	if string(received1.Payload) != "test1" {
		t.Errorf("Expected payload 'test1', got '%s'", string(received1.Payload))
	}

	received2 := <-q.DataChannel
	if string(received2.Payload) != "test2" {
		t.Errorf("Expected payload 'test2', got '%s'", string(received2.Payload))
	}
}

func TestQueue_UpdateStartTime(t *testing.T) {
	stats := metrics.New()
	q := New(10, stats)

	// Перша ініціалізація
	q.UpdateStartTime()
	time.Sleep(500 * time.Millisecond)

	_, _, _, uptime1 := q.Stats()
	if uptime1 < 500*time.Millisecond {
		t.Errorf("Expected uptime >= 500ms, got %s", uptime1)
	}

	// Скидаємо час
	q.UpdateStartTime()
	_, _, _, uptime2 := q.Stats()

	if uptime2 > 100*time.Millisecond {
		t.Errorf("Expected uptime to be reset, got %s", uptime2)
	}
}

// Benchmark для інкрементів
func BenchmarkQueue_IncrementAccepted(b *testing.B) {
	stats := metrics.New()
	q := New(10, stats)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.IncrementAccepted()
	}
}

func BenchmarkQueue_Stats(b *testing.B) {
	stats := metrics.New()
	q := New(10, stats)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Stats()
	}
}

func BenchmarkQueue_ConcurrentAccess(b *testing.B) {
	stats := metrics.New()
	q := New(10, stats)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			q.IncrementAccepted()
			q.GetConnectionStatus()
			q.Stats()
		}
	})
}