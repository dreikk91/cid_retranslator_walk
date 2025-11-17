package models

import (
	"sync"
	"time"

	"github.com/lxn/walk"
)

// EventItem - модель для таблиці Подій
type EventItem struct {
	Time     string
	Device   string
	Code     string
	Type     string
	Desc     string
	Zone     string
	Priority int // 0-критична, 1-помилка, 2-попередження, 3-інформація, 4-успіх
}

type EventModel struct {
	walk.TableModelBase
	items     []*EventItem
	tableView *walk.TableView
}

func NewEventModel() *EventModel {
	return &EventModel{
		//items: []*EventItem{
		//	{"10:30:15", "Інформація", "Система запущена", 4},
		//	{"10:28:42", "Попередження", "Низький рівень сигналу на ППК-005", 2},
		//	{"10:25:10", "Помилка", "Втрата зв'язку з ППК-003", 1},
		//},
	}
}

func (m *EventModel) RowCount() int {
	return len(m.items)
}

func (m *EventModel) Value(row, col int) interface{} {
	if row >= len(m.items) {
		return nil
	}
	item := m.items[row]
	switch col {
	case 0:
		return item.Time
	case 1:
		return item.Device
	case 2:
		return item.Code
	case 3:
		return item.Type
	case 4:
		return item.Desc
	case 5:
		return item.Zone
	}
	return nil
}

func (m *EventModel) SetTableView(tv *walk.TableView) {
	m.tableView = tv
}

func (m *EventModel) GetItems() []*EventItem {
	return m.items
}

// StartListening - запускає прослуховування каналу з батчингом
func (m *EventModel) StartListening(eventChan <-chan *EventItem) {
	pendingEvents := make([]*EventItem, 0, 100)
	var mu sync.Mutex

	// Горутина для читання з каналу
	go func() {
		for event := range eventChan {
			mu.Lock()
			pendingEvents = append(pendingEvents, event)
			mu.Unlock()
		}
	}()

	// Горутина для періодичного оновлення UI
	go func() {
		ticker := time.NewTicker(1000 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			if len(pendingEvents) == 0 {
				mu.Unlock()
				continue
			}

			eventsToProcess := make([]*EventItem, len(pendingEvents))
			copy(eventsToProcess, pendingEvents)
			pendingEvents = pendingEvents[:0]
			mu.Unlock()

			if m.tableView != nil {
				m.tableView.Synchronize(func() {
					// Додаємо всі події одразу
					for i := len(eventsToProcess) - 1; i >= 0; i-- {
						m.items = append([]*EventItem{eventsToProcess[i]}, m.items...)
					}
					// Обмежуємо кількість подій до 500
					if len(m.items) > 500 {
						m.items = m.items[:500]
					}
					m.PublishRowsReset()
				})
			}
		}
	}()
}
