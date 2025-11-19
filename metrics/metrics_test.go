package metrics

import (
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	stats := New()
	
	if stats == nil {
		t.Fatal("New() returned nil")
	}

	snap := stats.Snapshot()
	if snap.Accepted != 0 {
		t.Errorf("New() Accepted = %d, want 0", snap.Accepted)
	}
	if snap.Rejected != 0 {
		t.Errorf("New() Rejected = %d, want 0", snap.Rejected)
	}
	if snap.Reconnects != 0 {
		t.Errorf("New() Reconnects = %d, want 0", snap.Reconnects)
	}
	if snap.Connected {
		t.Error("New() Connected should be false")
	}
}

func TestIncrements(t *testing.T) {
	stats := New()

	stats.IncrementAccepted()
	stats.IncrementAccepted()
	stats.IncrementRejected()
	stats.IncrementReconnects()

	snap := stats.Snapshot()

	if snap.Accepted != 2 {
		t.Errorf("Accepted = %d, want 2", snap.Accepted)
	}
	if snap.Rejected != 1 {
		t.Errorf("Rejected = %d, want 1", snap.Rejected)
	}
	if snap.Reconnects != 1 {
		t.Errorf("Reconnects = %d, want 1", snap.Reconnects)
	}
}

func TestConnectionStatus(t *testing.T) {
	stats := New()

	// За замовчуванням false
	if stats.IsConnected() {
		t.Error("Initial IsConnected() should be false")
	}

	// Встановлюємо true
	stats.SetConnected(true)
	if !stats.IsConnected() {
		t.Error("IsConnected() should be true after SetConnected(true)")
	}

	// Встановлюємо false
	stats.SetConnected(false)
	if stats.IsConnected() {
		t.Error("IsConnected() should be false after SetConnected(false)")
	}
}

func TestReset(t *testing.T) {
	stats := New()

	// Збільшуємо лічильники
	stats.IncrementAccepted()
	stats.IncrementRejected()
	stats.IncrementReconnects()
	stats.SetConnected(true)

	time.Sleep(100 * time.Millisecond)

	// Скидаємо
	stats.Reset()

	snap := stats.Snapshot()

	if snap.Accepted != 0 {
		t.Errorf("After Reset() Accepted = %d, want 0", snap.Accepted)
	}
	if snap.Rejected != 0 {
		t.Errorf("After Reset() Rejected = %d, want 0", snap.Rejected)
	}
	if snap.Reconnects != 0 {
		t.Errorf("After Reset() Reconnects = %d, want 0", snap.Reconnects)
	}

	// Uptime має бути скинуто
	if snap.Uptime > 50*time.Millisecond {
		t.Errorf("After Reset() Uptime should be near 0, got %s", snap.Uptime)
	}

	// Connected НЕ скидається при Reset
	if !snap.Connected {
		t.Error("Reset() should not change Connected status")
	}
}

func TestSnapshot(t *testing.T) {
	stats := New()

	stats.IncrementAccepted()
	stats.IncrementRejected()
	stats.SetConnected(true)

	time.Sleep(1100 * time.Millisecond)

	snap := stats.Snapshot()

	if snap.Accepted != 1 {
		t.Errorf("Snapshot Accepted = %d, want 1", snap.Accepted)
	}
	if snap.Rejected != 1 {
		t.Errorf("Snapshot Rejected = %d, want 1", snap.Rejected)
	}
	if !snap.Connected {
		t.Error("Snapshot Connected should be true")
	}
	if snap.Uptime < 1*time.Second {
		t.Errorf("Snapshot Uptime should be >= 1s, got %s", snap.Uptime)
	}
}

func TestSnapshotUptimeString(t *testing.T) {
	snap := Snapshot{
		Uptime: 3665 * time.Second, // 1h 1m 5s
	}

	uptimeStr := snap.UptimeString()
	expected := "01:01:05"

	if uptimeStr != expected {
		t.Errorf("UptimeString() = %s, want %s", uptimeStr, expected)
	}
}

func TestSnapshotString(t *testing.T) {
	snap := Snapshot{
		Accepted:   100,
		Rejected:   10,
		Reconnects: 5,
		Uptime:     3665 * time.Second,
		Connected:  true,
	}

	str := snap.String()

	if str == "" {
		t.Error("String() returned empty string")
	}

	// Перевіряємо, що містить основні елементи
	if !contains(str, "Connected") {
		t.Error("String() should contain 'Connected'")
	}
	if !contains(str, "100") {
		t.Error("String() should contain accepted count")
	}
	if !contains(str, "10") {
		t.Error("String() should contain rejected count")
	}
}

func TestSnapshotTotalMessages(t *testing.T) {
	snap := Snapshot{
		Accepted: 100,
		Rejected: 50,
	}

	total := snap.TotalMessages()
	if total != 150 {
		t.Errorf("TotalMessages() = %d, want 150", total)
	}
}

func TestSnapshotSuccessRate(t *testing.T) {
	tests := []struct {
		name     string
		accepted int64
		rejected int64
		expected float64
	}{
		{
			name:     "100% success",
			accepted: 100,
			rejected: 0,
			expected: 100.0,
		},
		{
			name:     "50% success",
			accepted: 50,
			rejected: 50,
			expected: 50.0,
		},
		{
			name:     "0% success",
			accepted: 0,
			rejected: 100,
			expected: 0.0,
		},
		{
			name:     "No messages",
			accepted: 0,
			rejected: 0,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snap := Snapshot{
				Accepted: tt.accepted,
				Rejected: tt.rejected,
			}

			rate := snap.SuccessRate()
			if rate != tt.expected {
				t.Errorf("SuccessRate() = %.2f, want %.2f", rate, tt.expected)
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	stats := New()
	const goroutines = 100
	const operations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	// Горутини для IncrementAccepted
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				stats.IncrementAccepted()
			}
		}()
	}

	// Горутини для IncrementRejected
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				stats.IncrementRejected()
			}
		}()
	}

	// Горутини для SetConnected
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				stats.SetConnected(j%2 == 0)
			}
		}()
	}

	wg.Wait()

	snap := stats.Snapshot()

	expectedAccepted := int64(goroutines * operations)
	expectedRejected := int64(goroutines * operations)

	if snap.Accepted != expectedAccepted {
		t.Errorf("Accepted = %d, want %d", snap.Accepted, expectedAccepted)
	}

	if snap.Rejected != expectedRejected {
		t.Errorf("Rejected = %d, want %d", snap.Rejected, expectedRejected)
	}
}

// Benchmarks
func BenchmarkIncrementAccepted(b *testing.B) {
	stats := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats.IncrementAccepted()
	}
}

func BenchmarkIncrementRejected(b *testing.B) {
	stats := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats.IncrementRejected()
	}
}

func BenchmarkSetConnected(b *testing.B) {
	stats := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats.SetConnected(i%2 == 0)
	}
}

func BenchmarkSnapshot(b *testing.B) {
	stats := New()
	stats.IncrementAccepted()
	stats.IncrementRejected()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = stats.Snapshot()
	}
}

func BenchmarkConcurrentIncrements(b *testing.B) {
	stats := New()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stats.IncrementAccepted()
			stats.IncrementRejected()
			stats.SetConnected(true)
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr ||
		     containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}