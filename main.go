package main

import (
	"cid_retranslator_walk/adapters"
	"cid_retranslator_walk/cidparser"
	"cid_retranslator_walk/client"
	"cid_retranslator_walk/config"
	"cid_retranslator_walk/core"
	"cid_retranslator_walk/metrics"
	"cid_retranslator_walk/models"
	"cid_retranslator_walk/ui"
	"context"
	"log"
	"log/slog"
	"os"
	"time"
)

func main() {
	const eventFilePath = "events.json"
	jsonData, err := os.ReadFile(eventFilePath)
	if err != nil {
		log.Fatalf("Failed to load %s: %v", eventFilePath, err)
	}

	eventMap, err := cidparser.LoadEvents(jsonData)
	if err != nil {
		log.Fatal(err)
	}

	// 0. Завантажуємо конфігурацію
	cfg := config.New()

	// 1. Ініціалізація Core (TCP сервер/клієнт)
	stats := metrics.New()
	retranslator := core.NewApp(stats)

	// 2. Створюємо моделі UI
	ppkModel := models.NewPPKModel()
	eventModel := models.NewEventModel()
	statsData := models.NewStatsData()

	// 3. Створюємо канали UI з великими буферами для піків навантаження
	ppkChan := make(chan *models.PPKItem, 100)
	eventChan := make(chan *models.EventItem, 100)

	// 4. Підключаємо моделі до каналів (батчинг працює автоматично)
	ppkModel.StartListening(ppkChan)
	eventModel.StartListening(eventChan)

	// 5. Запускаємо TCP сервер/клієнт в окремій горутині
	go func() {
		slog.Info("Starting TCP retranslator...")
		retranslator.Startup()
	}()

	// 6. Ініціалізуємо адаптер
	adapter := adapters.NewAdapter(eventMap)

	// 7. Завантажуємо початковий стан (якщо є збережені дані)
	go func() {
		// 7. Завантажуємо початковий стан

		initialDevices := retranslator.GetInitialDevices()
		adapter.LoadInitialDevices(initialDevices, ppkChan)

		initialEvents := retranslator.GetInitialEvents()
		adapter.LoadInitialEvents(initialEvents, eventChan)

		slog.Info("Initial data loaded")
	}()

	// 8. Запускаємо адаптери - транслюють дані з Server в UI
	go func() {
		slog.Info("Starting device stream adapter...")
		deviceUpdatesChan := retranslator.GetDeviceUpdates()
		adapter.StreamDevicesToUI(deviceUpdatesChan, ppkChan)
	}()

	go func() {
		slog.Info("Starting event stream adapter...")
		eventUpdatesChan := retranslator.GetEventUpdates()
		adapter.StreamEventsToUI(eventUpdatesChan, eventChan)
	}()

	// 9. Створюємо контекст для передачі в UI
	appContext := &ui.AppContext{
		Retranslator: retranslator,
		Adapter:      adapter,
	}

	// 10. Створюємо і запускаємо UI (блокуюча операція)
	slog.Info("Creating main window...")
	mw := ui.CreateMainWindow(ppkModel, eventModel, appContext, statsData, cfg) // Передаємо контекст та конфіг

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startStatsUpdater(ctx, retranslator.GetClient(), statsData, mw)

	// 11. Run блокує виконання до закриття вікна
	slog.Info("Starting UI...")
	mw.Window.Run()

	// 12. Graceful shutdown після закриття UI
	slog.Info("UI closed, initiating shutdown...")
	retranslator.Shutdown(retranslator.Ctx())

	// Закриваємо UI канали
	close(ppkChan)
	close(eventChan)

	slog.Info("Application shutdown complete")
}

func startStatsUpdater(
	ctx context.Context,
	tcpClient *client.Client,
	statsData *models.StatsData,
	mainWindow *ui.MainWindowWithStats,
) {
	slog.Info("Stats updater started")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stats updater stopped")
			return
		case <-ticker.C:
			// Отримуємо статистику з клієнта (неблокуюча операція через канал)
			statsChan := tcpClient.GetQueueStats()

			select {
			case stats := <-statsChan:
				// Оновлюємо модель даних
				statsData.Update(stats)

				// Оновлюємо UI в головному потоці Walk
				mainWindow.Window.Synchronize(func() {
					mainWindow.UpdateStats(stats)
				})

			case <-time.After(500 * time.Millisecond):
				// Якщо клієнт не відповідає за 500мс, логуємо попередження
				slog.Warn("Stats channel timeout")
			}
		}
	}
}
