package models

import (
	"log/slog"
	"sort"
	"time"

	"github.com/lxn/walk"
)

type DetailItem struct {
	Time     time.Time
	Device   string
	Code     string
	Type     string
	Desc     string
	Zone     string
	Priority int
}

type DetailModel struct {
	walk.TableModelBase
	walk.SorterBase
	items        []*DetailItem
	tableView    *walk.TableView
	sortOrder    walk.SortOrder
	sortColumn   int
	uiDetailChan chan *DetailItem
	stopChan     chan struct{}
}

func NewDetailModel(ppkName string) *DetailModel {
	return &DetailModel{
		items:        make([]*DetailItem, 0, 100),
		sortOrder:    walk.SortDescending,
		sortColumn:   0,
		uiDetailChan: make(chan *DetailItem, 200),
		stopChan:     make(chan struct{}),
	}
}

func (m *DetailModel) Value(row, col int) interface{} {
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

func (m *DetailModel) RowCount() int {
	return len(m.items)
}

func (m *DetailModel) Sort(col int, order walk.SortOrder) error {
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
			return c(a.Time.Before(b.Time))
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

func (m *DetailModel) SetTableView(tv *walk.TableView) {
	m.tableView = tv
}

func (m *DetailModel) GetItem(row int) *DetailItem {
	if row < 0 || row >= len(m.items) {
		return nil
	}
	return m.items[row]
}

func (m *DetailModel) GetChannel() chan<- *DetailItem {
	return m.uiDetailChan
}

func (m *DetailModel) Stop() {
	close(m.stopChan)
}

func (m *DetailModel) GetStopChannel() <-chan struct{} {
	return m.stopChan
}

func (m *DetailModel) LoadInitialEvents(events []*DetailItem) {
	if m.tableView == nil {
		m.items = events
		return
	}

	m.tableView.Synchronize(func() {
		m.items = events
		err := m.Sort(m.sortColumn, m.sortOrder)
		if err != nil {
			slog.Debug("Initial sort error", "err", err)
		}
		m.PublishRowsReset()
	})
}

func (m *DetailModel) StartListening() {
	go func() {
		for ev := range m.uiDetailChan {
			if m.tableView == nil {
				continue
			}

			m.tableView.Synchronize(func() {
				m.items = append(m.items, ev)

				// Обмежуємо до 100 останніх подій
				if len(m.items) > 100 {
					m.items = m.items[len(m.items)-100:]
				}

				err := m.Sort(m.sortColumn, m.sortOrder)
				if err != nil {
					slog.Debug("Device event sort error", "err", err)
				}
				m.PublishRowsReset()
			})
		}
		slog.Info("DetailModel listening stopped")
	}()
}
