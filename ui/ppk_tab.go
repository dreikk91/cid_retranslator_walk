package ui

import (
	"cid_retranslator_walk/config"
	"cid_retranslator_walk/constants"
	"cid_retranslator_walk/models"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func CreatePPKTab(
	ppkModel *models.PPKModel,
	ppkTableView **walk.TableView,
	mw **walk.MainWindow,
	appCtx *AppContext, // Додаємо контекст
	cfg *config.Config, // Додаємо конфіг
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

					// Перевірка на таймаут
					if time.Since(item.Date) > cfg.Monitoring.PPKTimeout {
						style.BackgroundColor = constants.ColorRed
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

// StartPPKRefresh запускає періодичне оновлення таблиці ППК для відображення таймаутів
func StartPPKRefresh(tv *walk.TableView) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if tv != nil {
				tv.Synchronize(func() {
					tv.Invalidate()
				})
			}
		}
	}()
}
