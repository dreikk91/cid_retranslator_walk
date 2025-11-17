package ui

import (
	"cid_retranslator_walk/constants"
	"cid_retranslator_walk/models"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func CreateEventsTab(eventModel *models.EventModel, eventTableView **walk.TableView) TabPage {
	return TabPage{
		Title:  "Події",
		Layout: VBox{},
		Children: []Widget{
			TableView{
				AssignTo:         eventTableView,
				AlternatingRowBG: true,
				ColumnsOrderable: true,
				Model:            eventModel,
				Columns: []TableViewColumn{
					{Title: "Час", Width: 120},
					{Title: "ППК", Width: 50},
					{Title: "Код", Width: 50},
					{Title: "Тип", Width: 150},
					{Title: "Опис", Width: 300},
					{Title: "Зона|Група", Width: 120},
				},
				StyleCell: func(style *walk.CellStyle) {
					// Використовуємо безпечний метод getItem (всі зміни в UI thread)
					item := eventModel.GetItem(style.Row())
					if item == nil {
						return
					}

					switch item.Priority {
					case constants.UnknownEvent:
						style.BackgroundColor = constants.UnknownEventBG
						style.TextColor = constants.UnknownEventText
					case constants.GuardEvent:
						style.BackgroundColor = constants.GuardEventBG
						style.TextColor = constants.GuardEventText
					case constants.DisguardEvent:
						style.BackgroundColor = constants.DisguardEventBG
						style.TextColor = constants.DisguardEventText
					case constants.OkEvent:
						style.BackgroundColor = constants.OkEventBG
						style.TextColor = constants.OkEventText
					case constants.AlarmEvent:
						style.BackgroundColor = constants.AlarmEventBG
						style.TextColor = constants.AlarmEventText
					case constants.OtherEvent:
						style.BackgroundColor = constants.OtherEventBG
						style.TextColor = constants.OtherEventText
					default:
						style.BackgroundColor = constants.UnknownEventBG
						style.TextColor = constants.UnknownEventText
					}
				},
			},
		},
	}
}
