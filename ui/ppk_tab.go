package ui

import (
	"cid_retranslator_walk/constants"
	"cid_retranslator_walk/models"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func CreatePPKTab(
	ppkModel *models.PPKModel,
	ppkTableView **walk.TableView,
	mw **walk.MainWindow,
	appCtx *AppContext, // Додаємо контекст
) TabPage {
	return TabPage{
		Title:  "ППК",
		Layout: VBox{},
		Children: []Widget{
			TableView{
				AssignTo:            ppkTableView,
				AlternatingRowBG:    true,
				ColumnsOrderable:    true,
				LastColumnStretched: true,
				Model:               ppkModel,
				Columns: []TableViewColumn{
					{Title: "№", Width: 60},
					{Title: "ППК", Width: 80},
					{Title: "Остання подія", Width: 160}, // Ця колонка розтягується
					{Title: "Дата/Час", Width: 160},
				},
				OnItemActivated: func() {
					if (*ppkTableView).CurrentIndex() >= 0 {
						item := ppkModel.GetItem((*ppkTableView).CurrentIndex())
						if item != nil {
							ShowPPKDetails(*mw, item, appCtx) // Передаємо контекст
						}
					}
				},
				StyleCell: func(style *walk.CellStyle) {
					item := ppkModel.GetItem(style.Row())
					if item == nil {
						return
					}

					if style.Row()%2 == 0 {
						style.BackgroundColor = constants.ColorGray
					}

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
