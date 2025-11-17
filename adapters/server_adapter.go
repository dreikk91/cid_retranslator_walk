package adapters

import (
	"cid_retranslator_walk/cidparser"
	"cid_retranslator_walk/constants"
	"cid_retranslator_walk/models"
	"cid_retranslator_walk/server"
	"fmt"
	"log/slog"
	"strings"
)

type Adapter struct {
	EventMap cidparser.EventMap
}

func NewAdapter(eventMap cidparser.EventMap) *Adapter {
	return &Adapter{
		EventMap: eventMap,
	}
}

// StreamDevicesToUI транслює оновлення пристроїв з Server в UI
func (ad Adapter) StreamDevicesToUI(serverDeviceChan <-chan server.Device, uiPPKChan chan<- *models.PPKItem) {
	slog.Info("Device adapter started")
	for device := range serverDeviceChan {
		//status := determineDeviceStatus(device.LastEvent)

		uiItem := &models.PPKItem{
			Number: device.ID,
			Name:   fmt.Sprintf("%03d", device.ID),
			Event:  device.LastEvent,
			Date:   device.LastEventTime,
		}

		// Non-blocking send
		select {
		case uiPPKChan <- uiItem:
		default:
			slog.Warn("UI PPK channel full, dropping device update", "deviceID", device.ID)
		}
	}
	slog.Info("Device adapter stopped")
}

// StreamEventsToUI транслює глобальні події з Server в UI
func (ad Adapter) StreamEventsToUI(serverEventChan <-chan server.GlobalEvent, uiEventChan chan<- *models.EventItem) {
	slog.Info("Event adapter started")
	for event := range serverEventChan {

		code := event.Data[11:15]
		group := event.Data[15:17]
		zone := event.Data[17:20]
		eventType, desc, found := ad.EventMap.GetEventDescriptions(code)
		if !found {
			continue
		}
		priority, eventType := ad.determineEventPriority(code, eventType)
		uiEvent := &models.EventItem{
			Time:     event.Time,
			Device:   fmt.Sprint(event.DeviceID),
			Code:     code,
			Type:     eventType,
			Desc:     desc,
			Zone:     fmt.Sprintf("Зона %s|Група %s", zone, group),
			Priority: priority,
		}

		// Non-blocking send
		select {
		case uiEventChan <- uiEvent:
		default:
			slog.Warn("UI Event channel full, dropping event", "deviceID", event.DeviceID)
		}
	}
	slog.Info("Event adapter stopped")
}

// LoadInitialDevices завантажує початковий стан пристроїв в UI
func (ad Adapter) LoadInitialDevices(devices []server.Device, uiPPKChan chan<- *models.PPKItem) {
	slog.Info("Loading initial devices", "count", len(devices))
	for _, device := range devices {
		status := ad.determineDeviceStatus(device.LastEvent)

		uiItem := &models.PPKItem{
			Number: device.ID,
			Name:   fmt.Sprintf("ППК-%03d", device.ID),
			Status: status,
			Date:   device.LastEventTime,
		}

		select {
		case uiPPKChan <- uiItem:
		default:
			slog.Warn("UI PPK channel full during initial load", "deviceID", device.ID)
		}
	}
	slog.Info("Initial devices loaded")
}

// LoadInitialEvents завантажує початковий стан подій в UI
func (ad Adapter) LoadInitialEvents(events []server.GlobalEvent, uiEventChan chan<- *models.EventItem) {
	slog.Info("Loading initial events", "count", len(events))
	for _, event := range events {
		priority, eventType := ad.determineEventPriority(event.Data, "success")

		uiEvent := &models.EventItem{
			Time:     event.Time,
			Type:     eventType,
			Desc:     formatEventDescription(event),
			Priority: priority,
		}

		select {
		case uiEventChan <- uiEvent:
		default:
			slog.Warn("UI Event channel full during initial load")
		}
	}
	slog.Info("Initial events loaded")
}

// determineDeviceStatus визначає статус пристрою на основі останньої події
func (ad Adapter) determineDeviceStatus(lastEvent string) string {
	// Логіка визначення статусу на основі вмісту події
	// Це приклад - адаптуйте під вашу логіку

	if strings.Contains(lastEvent, "error") || strings.Contains(lastEvent, "fail") {
		return "Помилка"
	}

	if strings.Contains(lastEvent, "warning") || strings.Contains(lastEvent, "warn") {
		return "Попередження"
	}

	// За замовчуванням - активний
	return "Активний"
}

// determineEventPriority визначає пріоритет і тип події
func (ad Adapter) determineEventPriority(code, event string) (int, string) {
	// Приклад логіки визначення пріоритету
	// Адаптуйте під ваш протокол CID

	//eventLower := strings.ToLower(eventData)
	eventType := cidparser.GetColorByEvent(code)

	if strings.Contains(eventType, "unknown") {
		return constants.UnknownEvent, event
	}

	if strings.Contains(eventType, "guard") {
		return constants.GuardEvent, event
	}

	if strings.Contains(eventType, "disguard") {
		return constants.DisguardEvent, event
	}

	// Успіх
	if strings.Contains(eventType, "ok") {
		return constants.OkEvent, event
	}

	if strings.Contains(eventType, "alarm") {
		return constants.AlarmEvent, event
	}

	// За замовчуванням - інформація
	return constants.UnknownEvent, event
}

// formatEventDescription форматує опис події для UI
func formatEventDescription(event server.GlobalEvent) string {
	// Можете додати більше логіки форматування
	return fmt.Sprintf("[ППК-%03d] %s", event.DeviceID, event.Data)
}
