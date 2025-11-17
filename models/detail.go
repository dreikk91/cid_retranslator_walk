package models

import (
	"time"

	"github.com/lxn/walk"
)

// DetailItem - модель для деталей ППК
type DetailItem struct {
	Param string
	Value string
	Unit  string
	Time  string
}

type DetailModel struct {
	walk.TableModelBase
	items []*DetailItem
}

func NewDetailModel(ppkName string) *DetailModel {
	return &DetailModel{
		items: []*DetailItem{
			{"Температура", "22.5", "°C", time.Now().Format("15:04:05")},
			{"Напруга", "12.3", "В", time.Now().Format("15:04:05")},
			{"Струм", "0.85", "А", time.Now().Format("15:04:05")},
			{"Завантаження CPU", "45", "%", time.Now().Format("15:04:05")},
			{"Вільна пам'ять", "2048", "МБ", time.Now().Format("15:04:05")},
			{"Кількість підключень", "12", "шт", time.Now().Format("15:04:05")},
		},
	}
}

func (m *DetailModel) RowCount() int {
	return len(m.items)
}

func (m *DetailModel) Value(row, col int) interface{} {
	item := m.items[row]
	switch col {
	case 0:
		return item.Param
	case 1:
		return item.Value
	case 2:
		return item.Unit
	case 3:
		return item.Time
	}
	return nil
}
