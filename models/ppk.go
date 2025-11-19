package models

import (
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/lxn/walk"
)

type PPKItem struct {
	Number int
	Name   string
	Event  string
	Date   time.Time
	Status string
}

type PPKModel struct {
	walk.TableModelBase
	walk.SorterBase
	sortColumn int
	sortOrder  walk.SortOrder
	items      []*PPKItem
	tableView  *walk.TableView
}

func NewPPKModel() *PPKModel {
	m := &PPKModel{
		items:     make([]*PPKItem, 0, 100),
		sortOrder: walk.SortAscending, // За замовчуванням ascending за номером
	}
	slog.Info("PPKModel created", "initialItems", len(m.items))
	return m
}

func (m *PPKModel) RowCount() int {
	return len(m.items)
}

func (m *PPKModel) Value(row, col int) interface{} {
	if row >= len(m.items) {
		return nil
	}
	item := m.items[row]
	switch col {
	case 0:
		return item.Number
	case 1:
		return item.Name
	case 2:
		return item.Event
	case 3:
		return item.Date.Format("15:04:05 2006-01-02")
	}
	return nil
}

func (m *PPKModel) Sort(col int, order walk.SortOrder) error {
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
			return c(a.Number < b.Number)
		case 1:
			return c(a.Name < b.Name)
		case 2:
			return c(a.Event < b.Event)
		case 3:
			return c(a.Date.Before(b.Date))
		}
		return false
	})

	return m.SorterBase.Sort(col, order)
}

func (m *PPKModel) SetTableView(tv *walk.TableView) {
	m.tableView = tv
	slog.Info("PPKModel.SetTableView called", "tableView", tv != nil)
}

func (m *PPKModel) GetItems() []*PPKItem {
	return m.items
}

func (m *PPKModel) GetItem(row int) *PPKItem {
	if row < 0 || row >= len(m.items) {
		return nil
	}
	return m.items[row]
}

func (m *PPKModel) StartListening(dataChan <-chan *PPKItem) {
	slog.Info("PPKModel.StartListening started")

	pendingItems := make([]*PPKItem, 0, 100)
	var mu sync.Mutex

	// Горутина 1: Читання з каналу
	go func() {
		slog.Info("PPKModel reader goroutine started")
		itemCount := 0
		for item := range dataChan {
			itemCount++
			mu.Lock()
			pendingItems = append(pendingItems, item)
			mu.Unlock()
		}
		slog.Info("PPKModel reader goroutine stopped")
	}()

	// Горутина 2: Періодичне оновлення UI
	go func() {
		slog.Info("PPKModel updater goroutine started")
		ticker := time.NewTicker(1000 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			if len(pendingItems) == 0 {
				mu.Unlock()
				continue
			}

			itemsToProcess := make([]*PPKItem, len(pendingItems))
			copy(itemsToProcess, pendingItems)
			pendingItems = pendingItems[:0]
			mu.Unlock()

			if m.tableView == nil {
				slog.Error("PPKModel.tableView is NIL! Cannot update UI")
				continue
			}

			slog.Info("PPKModel processing batch",
				"batchSize", len(itemsToProcess))

			m.tableView.Synchronize(func() {
				updatedRows := make(map[int]bool)
				addedRows := make([]*PPKItem, 0)

				for _, item := range itemsToProcess {
					found := false
					for i, existing := range m.items {
						if existing.Number == item.Number {
							m.items[i] = item
							updatedRows[i] = true
							found = true
							break
						}
					}
					if !found {
						addedRows = append(addedRows, item)
					}
				}

				// Повідомляємо про оновлені рядки
				if len(updatedRows) > 0 {
					// Walk не підтримує PublishRowsChanged для окремих рядків,
					// тому викликаємо для всього діапазону
					m.PublishRowsChanged(0, len(m.items)-1)
				}

				// Додаємо нові рядки в кінець
				if len(addedRows) > 0 {
					oldLen := len(m.items)
					m.items = append(m.items, addedRows...)
					m.PublishRowsInserted(oldLen, len(addedRows))

					// Автоматично сортуємо після додавання
					if err := m.Sort(m.sortColumn, m.sortOrder); err != nil {
						slog.Error("Failed to sort PPK items", "error", err)
					}
				}

				slog.Info("PPKModel update completed", "totalItems", len(m.items))
			})
		}
		slog.Info("PPKModel updater goroutine stopped")
	}()
}
