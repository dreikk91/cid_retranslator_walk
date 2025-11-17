// models/ppk.go - VERSION WITH DETAILED LOGS

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
	Date   string
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
		items: make([]*PPKItem, 0, 100),
	}

	//statuses := []string{"Активний", "Помилка", "Попередження"}
	//for i := 1; i <= 100; i++ {
	//	m.items = append(m.items, &PPKItem{
	//		Number: i,
	//		Name:   fmt.Sprintf("ППК-%03d", i),
	//		Status: statuses[i%3],
	//		Date:   time.Now().Format("2006-01-02 15:04:05"),
	//	})
	//}

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
		return item.Date
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
			return c(a.Date < b.Date)
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
			slog.Debug("PPKModel received item",
				"number", item.Number,
				"name", item.Name,
				"totalReceived", itemCount)

			mu.Lock()
			pendingItems = append(pendingItems, item)
			pendingCount := len(pendingItems)
			mu.Unlock()

			slog.Debug("PPKModel item buffered",
				"pendingCount", pendingCount)
		}
		slog.Info("PPKModel reader goroutine stopped")
	}()

	// Горутина 2: Періодичне оновлення UI
	go func() {
		slog.Info("PPKModel updater goroutine started")
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		updateCount := 0
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

			updateCount++
			slog.Info("PPKModel processing batch",
				"batchSize", len(itemsToProcess),
				"updateNumber", updateCount,
				"tableView", m.tableView != nil)

			if m.tableView == nil {
				slog.Error("PPKModel.tableView is NIL! Cannot update UI")
				continue
			}

			m.tableView.Synchronize(func() {
				slog.Debug("PPKModel inside Synchronize",
					"itemsToProcess", len(itemsToProcess))

				for _, item := range itemsToProcess {
					found := false
					for i, existing := range m.items {
						if existing.Number == item.Number {
							slog.Debug("PPKModel updating existing item",
								"number", item.Number)
							m.items[i] = item
							found = true
							break
						}
					}
					if !found {
						slog.Debug("PPKModel adding new item",
							"number", item.Number)
						m.items = append(m.items, item)
					}
				}

				slog.Info("PPKModel calling PublishRowsReset",
					"totalItems", len(m.items))
				m.PublishRowsReset()
				slog.Info("PPKModel PublishRowsReset completed")
			})
		}
		slog.Info("PPKModel updater goroutine stopped")
	}()
}

func (m *PPKModel) GetItemUnsafe(row int) *PPKItem {
	if row < 0 || row >= len(m.items) {
		return nil
	}
	return m.items[row]
}
