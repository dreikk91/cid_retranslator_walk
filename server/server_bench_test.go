package server

import (
	"cid_retranslator_walk/config"
	"cid_retranslator_walk/metrics"
	"cid_retranslator_walk/queue"
	"context"
	"testing"
	"time"
)

// BenchmarkUpdateDevice вимірює продуктивність UpdateDevice
func BenchmarkUpdateDevice(b *testing.B) {
	stats := metrics.New()
	q := queue.New(100, stats)
	
	cfg := &config.ServerConfig{
		Host: "localhost",
		Port: "20005",
	}
	
	rules := &config.CIDRules{
		RequiredPrefix: "5",
		ValidLength:    20,
	}
	
	server := New(cfg, q, rules)
	
	// Запускаємо сервер в фоні
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go func() {
		// Споживаємо події з каналів
		for {
			select {
			case <-server.deviceUpdates:
			case <-server.eventUpdates:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	eventData := "1234567210011223344"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deviceID := 2000 + (i % 200) // Ротуємо між 2000-2199
		server.UpdateDevice(deviceID, eventData)
	}
}

// BenchmarkUpdateDeviceConcurrent вимірює конкурентну продуктивність
func BenchmarkUpdateDeviceConcurrent(b *testing.B) {
	stats := metrics.New()
	q := queue.New(100, stats)
	
	cfg := &config.ServerConfig{
		Host: "localhost",
		Port: "20005",
	}
	
	rules := &config.CIDRules{
		RequiredPrefix: "5",
		ValidLength:    20,
	}
	
	server := New(cfg, q, rules)
	
	// Запускаємо сервер в фоні
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go func() {
		for {
			select {
			case <-server.deviceUpdates:
			case <-server.eventUpdates:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	eventData := "1234567210011223344"
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		deviceID := 2000
		for pb.Next() {
			server.UpdateDevice(deviceID, eventData)
			deviceID = (deviceID + 1)
			if deviceID > 2200 {
				deviceID = 2000
			}
		}
	})
}

// BenchmarkUpdateDeviceWithChannelConsumer вимірює з реальним споживачем
func BenchmarkUpdateDeviceWithChannelConsumer(b *testing.B) {
	stats := metrics.New()
	q := queue.New(100, stats)
	
	cfg := &config.ServerConfig{
		Host: "localhost",
		Port: "20005",
	}
	
	rules := &config.CIDRules{
		RequiredPrefix: "5",
		ValidLength:    20,
	}
	
	server := New(cfg, q, rules)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Реалістичний споживач з обробкою
	go func() {
		for {
			select {
			case dev := <-server.deviceUpdates:
				_ = dev.ID // Симулюємо обробку
			case ev := <-server.eventUpdates:
				_ = ev.DeviceID
			case <-ctx.Done():
				return
			}
		}
	}()
	
	eventData := "1234567210011223344"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deviceID := 2000 + (i % 10)
		server.UpdateDevice(deviceID, eventData)
	}
}

// BenchmarkGetDevices вимірює продуктивність GetDevices
func BenchmarkGetDevices(b *testing.B) {
	stats := metrics.New()
	q := queue.New(100, stats)
	
	cfg := &config.ServerConfig{
		Host: "localhost",
		Port: "20005",
	}
	
	rules := &config.CIDRules{}
	
	server := New(cfg, q, rules)
	
	// Додаємо тестові пристрої
	for i := 2000; i < 2100; i++ {
		server.devices[i] = &Device{
			ID:            i,
			LastEventTime: time.Now(),
			LastEvent:     "test",
			Events:        make([]Event, 0, maxDeviceEvents),
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.GetDevices()
	}
}

// BenchmarkGetGlobalEvents вимірює продуктивність GetGlobalEvents
func BenchmarkGetGlobalEvents(b *testing.B) {
	stats := metrics.New()
	q := queue.New(100, stats)
	
	cfg := &config.ServerConfig{
		Host: "localhost",
		Port: "20005",
	}
	
	rules := &config.CIDRules{}
	
	server := New(cfg, q, rules)
	
	// Додаємо тестові події
	for i := 0; i < maxGlobalEvents; i++ {
		server.globalEventsRing.Value = GlobalEvent{
			Time:     time.Now(),
			DeviceID: 2000 + i,
			Data:     "test event",
		}
		server.globalEventsRing = server.globalEventsRing.Next()
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.GetGlobalEvents()
	}
}

// BenchmarkExtractDeviceID вимірює продуктивність extractDeviceID
func BenchmarkExtractDeviceID(b *testing.B) {
	message := []byte("1234567210011223344")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractDeviceID(message)
	}
}

// BenchmarkDeviceEventChannel вимірює роботу з device-specific каналами
func BenchmarkDeviceEventChannel(b *testing.B) {
	stats := metrics.New()
	q := queue.New(100, stats)
	
	cfg := &config.ServerConfig{
		Host: "localhost",
		Port: "20005",
	}
	
	rules := &config.CIDRules{}
	
	server := New(cfg, q, rules)
	
	const deviceID = 2100
	ch := server.GetDeviceEventChannel(deviceID)
	
	// Споживач
	go func() {
		for range ch {
			// Обробка
		}
	}()
	
	eventData := "1234567210011223344"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.UpdateDevice(deviceID, eventData)
	}
	
	b.StopTimer()
	server.CloseDeviceEventChannel(deviceID)
}

// BenchmarkCleanupInactiveDevices вимірює продуктивність очищення
func BenchmarkCleanupInactiveDevices(b *testing.B) {
	stats := metrics.New()
	q := queue.New(100, stats)
	
	cfg := &config.ServerConfig{
		Host: "localhost",
		Port: "20005",
	}
	
	rules := &config.CIDRules{}
	
	server := New(cfg, q, rules)
	
	// Додаємо активні та неактивні пристрої
	now := time.Now()
	for i := 2000; i < 2050; i++ {
		server.devices[i] = &Device{
			ID:            i,
			LastEventTime: now,
			LastEvent:     "test",
		}
		// Половина активних, половина старих
		if i%2 == 0 {
			server.lastActive[i] = now
		} else {
			server.lastActive[i] = now.Add(-2 * time.Hour)
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.cleanupInactiveDevices()
		
		// Відновлюємо стан для наступної ітерації
		for id := 2001; id < 2050; id += 2 {
			server.devices[id] = &Device{
				ID:            id,
				LastEventTime: now,
				LastEvent:     "test",
			}
			server.lastActive[id] = now.Add(-2 * time.Hour)
		}
	}
}