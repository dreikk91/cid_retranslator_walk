package ui

import (
	"cid_retranslator_walk/constants"
	"cid_retranslator_walk/models"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func CreatePPKTab(ppkModel *models.PPKModel, ppkTableView **walk.TableView, mw **walk.MainWindow) TabPage {
	return TabPage{
		Title:  "ППК",
		Layout: VBox{},
		Children: []Widget{
			TableView{
				AssignTo:         ppkTableView,
				AlternatingRowBG: true,
				ColumnsOrderable: true,
				Model:            ppkModel,
				Columns: []TableViewColumn{
					{Title: "№", Width: 50},
					{Title: "ППК", Width: 50},
					{Title: "Остання подія", Width: 250},
					{Title: "Дата/Час", Width: 200},
				},
				OnItemActivated: func() {
					if (*ppkTableView).CurrentIndex() >= 0 && (*ppkTableView).CurrentIndex() < len(ppkModel.GetItems()) {
						item := ppkModel.GetItems()[(*ppkTableView).CurrentIndex()]
						ShowPPKDetails(*mw, item)
					}
				},
				StyleCell: func(style *walk.CellStyle) {
					items := ppkModel.GetItems()
					if style.Row() >= len(items) {
						return
					}
					if style.Row()%2 == 0 {
						style.BackgroundColor = constants.ColorGray
					}
					item := items[style.Row()]
					if style.Col() == 2 {
						switch item.Status {
						case "Помилка":
							style.TextColor = constants.ColorRed
						case "Попередження":
							style.TextColor = constants.ColorOrange
						case "Активний":
							style.TextColor = constants.ColorGreen
						}
					}
				},
			},
		},
	}
}
