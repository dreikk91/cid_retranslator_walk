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
					{Title: "Час", Width: 150},
					{Title: "ППК", Width: 80},
					{Title: "Код", Width: 80},
					{Title: "Тип", Width: 200},
					{Title: "Опис", Width: 350},
					{Title: "Зона|Група", Width: 120},
				},
				StyleCell: func(style *walk.CellStyle) {
					items := eventModel.GetItems()
					if style.Row() >= len(items) {
						return
					}
					item := items[style.Row()]

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
