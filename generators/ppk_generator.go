package generators

import (
	"cid_retranslator_walk/models"
	"fmt"
	"math/rand"
	"time"
)

// GeneratePPKData - генератор тестових даних для ППК
func GeneratePPKData(ppkChan chan<- *models.PPKItem) {
	statuses := []string{"Активний", "Помилка", "Попередження"}

	ticker := time.NewTicker(50 * time.Millisecond) // 20 разів на секунду
	defer ticker.Stop()

	for range ticker.C {
		// Оновлюємо випадковий ППК з діапазону 1-100
		ppkNumber := rand.Intn(100) + 1
		item := &models.PPKItem{
			Number: ppkNumber,
			Name:   fmt.Sprintf("ППК-%03d", ppkNumber),
			Status: statuses[rand.Intn(len(statuses))],
			Date:   time.Now().Format("2006-01-02 15:04:05.000"),
		}
		ppkChan <- item
	}
}
