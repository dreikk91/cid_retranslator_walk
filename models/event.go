package models

import (
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/lxn/walk"
)

type EventItem struct {
	ID       int64
	Time     time.Time
	Device   string
	Code     string
	Type     string
	Desc     string
	Zone     string
	Priority int
}

type EventModel struct {
	walk.TableModelBase
	walk.SorterBase
	items      []*EventItem
	tableView  *walk.TableView
	sortColumn int
	sortOrder  walk.SortOrder
}

func NewEventModel() *EventModel {
	m := &EventModel{
		items:     make([]*EventItem, 0, 520), // preallocate
		sortOrder: walk.SortDescending,        // За замовчуванням descending за часом
	}
	return m
}

func (m *EventModel) RowCount() int {
	return len(m.items)
}

func (m *EventModel) Value(row, col int) interface{} {
	if row < 0 || row >= len(m.items) {
		return nil
	}
	item := m.items[row]
	switch col {
	case 0:
		return item.Time.Format("15:04:05 2006-01-02")
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

func (m *EventModel) Sort(col int, order walk.SortOrder) error {
	m.sortColumn = col
	m.sortOrder = order

	sort.SliceStable(m.items, func(i, j int) bool {
		a, b := m.items[i], m.items[j]
		c := func(ls bool) bool {
			if m.sortOrder == walk.SortAscending {
				return ls
			}
			return !ls
		}

		switch col {
		case 0:
			return a.Time.After(b.Time)
		case 1:
			return c(a.Device < b.Device)
		case 2:
			return c(a.Code < b.Code)
		case 3:
			return c(a.Type < b.Type)
		case 4:
			return c(a.Desc < b.Desc)
		case 5:
			return c(a.Zone < b.Zone)
		}
		return false
	})

	return m.SorterBase.Sort(col, order)
}

func (m *EventModel) SetTableView(tv *walk.TableView) {
	m.tableView = tv
}

func (m *EventModel) GetItem(row int) *EventItem {
	if row < 0 || row >= len(m.items) {
		return nil
	}
	return m.items[row]
}

func (m *EventModel) StartListening(eventChan <-chan *EventItem) {
	pendingEvents := make([]*EventItem, 0, 400)
	var mu sync.Mutex

	// Читання з каналу
	go func() {
		for event := range eventChan {
			mu.Lock()
			pendingEvents = append(pendingEvents, event)
			mu.Unlock()
		}
	}()

	// Батчинг оновлення UI
	go func() {
		ticker := time.NewTicker(400 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			if len(pendingEvents) == 0 {
				mu.Unlock()
				continue
			}
			batch := make([]*EventItem, len(pendingEvents))
			copy(batch, pendingEvents)
			pendingEvents = pendingEvents[:0]
			mu.Unlock()

			if m.tableView == nil {
				continue
			}

			m.tableView.Synchronize(func() {
				oldLen := len(m.items)
				added := len(batch)

				// Append в кінець (швидко, без зсуву)
				m.items = append(m.items, batch...)

				// Сповіщаємо про вставку в кінець
				m.PublishRowsInserted(oldLen, added)

				// Обрізаємо старі, якщо > 500 (видаляємо з початку)
				if len(m.items) > 500 {
					excess := len(m.items) - 500
					if excess > 400 {
						excess = 400
					
					}
					itemToDelete := 500 - excess
					m.items = m.items[0:itemToDelete]
					m.PublishRowsRemoved(0, excess)
				}

				// Сортуємо (новіші зверху) - це сповістить UI про зміни гладко
				err := m.Sort(0, walk.SortAscending)
				if err != nil {
					slog.Debug("Event sort error", "err", err)
				}

			})
		}
	}()
}
