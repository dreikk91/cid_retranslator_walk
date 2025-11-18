package main

import (
	"cid_retranslator_walk/adapters"
	"cid_retranslator_walk/cidparser"
	"cid_retranslator_walk/core"
	"cid_retranslator_walk/models"
	"cid_retranslator_walk/ui"
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

	// 1. Ініціалізація Core (TCP сервер/клієнт)
	retranslator := core.NewApp()

	// 2. Створюємо моделі UI
	ppkModel := models.NewPPKModel()
	eventModel := models.NewEventModel()

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

	// 6. Чекаємо трохи, щоб TCP сервер встиг запуститись
	time.Sleep(500 * time.Millisecond)
	adapter := adapters.NewAdapter(eventMap)

	// 7. Завантажуємо початковий стан (якщо є збережені дані)
	go func() {
		time.Sleep(1 * time.Second) // Чекаємо поки сервер повністю запуститься

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
	mw := ui.CreateMainWindow(ppkModel, eventModel, appContext) // Передаємо контекст

	// 11. Run блокує виконання до закриття вікна
	slog.Info("Starting UI...")
	mw.Run()

	// 12. Graceful shutdown після закриття UI
	slog.Info("UI closed, initiating shutdown...")
	retranslator.Shutdown(retranslator.Ctx())

	// Закриваємо UI канали
	close(ppkChan)
	close(eventChan)

	slog.Info("Application shutdown complete")
}
