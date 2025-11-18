package generators

import (
	"cid_retranslator_walk/constants"
	"cid_retranslator_walk/models"
	"math/rand"
	"time"
)

// GenerateEvents - генератор тестових подій
func GenerateEvents(eventChan chan<- *models.EventItem) {
	eventTypes := []struct {
		Type     string
		Priority int
	}{
		{"Критична помилка", constants.PriorityCritical},
		{"Помилка", constants.PriorityError},
		{"Попередження", constants.PriorityWarning},
		{"Інформація", constants.PriorityInfo},
		{"Успіх", constants.PrioritySuccess},
	}

	descriptions := []string{
		"Втрата зв'язку з пристроєм",
		"Перевищення температури",
		"Успішне підключення",
		"Оновлення конфігурації",
		"Низький рівень сигналу",
		"Резервне копіювання завершено",
	}

	ticker := time.NewTicker(100 * time.Millisecond) // 10 разів на секунду
	defer ticker.Stop()

	for range ticker.C {
		eventType := eventTypes[rand.Intn(len(eventTypes))]
		event := &models.EventItem{
			Time:     time.Now(),
			Type:     eventType.Type,
			Desc:     descriptions[rand.Intn(len(descriptions))],
			Priority: eventType.Priority,
		}
		eventChan <- event
	}
}
