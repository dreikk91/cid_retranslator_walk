package ui

import (
	"cid_retranslator_walk/constants"
	"cid_retranslator_walk/models"
	"fmt"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func ShowPPKDetails(owner walk.Form, ppkItem *models.PPKItem) {
	var dlg *walk.Dialog
	var tableView *walk.TableView

	model := models.NewDetailModel(ppkItem.Name)

	Dialog{
		AssignTo: &dlg,
		Title:    fmt.Sprintf("Деталі: %s", ppkItem.Name),
		MinSize:  Size{Width: 600, Height: 400},
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "Номер:", Font: Font{Bold: true}},
					Label{Text: fmt.Sprintf("%d", ppkItem.Number)},
					Label{Text: "Назва:", Font: Font{Bold: true}},
					Label{Text: ppkItem.Name},
					Label{Text: "Статус:", Font: Font{Bold: true}},
					Label{Text: ppkItem.Status},
					Label{Text: "Оновлено:", Font: Font{Bold: true}},
					Label{Text: ppkItem.Date},
				},
			},
			Label{
				Text: "Параметри:",
				Font: Font{PointSize: 10, Bold: true},
			},
			TableView{
				AssignTo:         &tableView,
				AlternatingRowBG: true,
				ColumnsOrderable: true,
				Model:            model,
				Columns: []TableViewColumn{
					{Title: "Параметр", Width: 150},
					{Title: "Значення", Width: 100},
					{Title: "Одиниці", Width: 80},
					{Title: "Час", Width: 100},
				},
				StyleCell: func(style *walk.CellStyle) {
					if style.Row()%2 == 0 {
						style.BackgroundColor = constants.ColorGray
					}
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "Закрити",
						OnClicked: func() {
							dlg.Accept()
						},
					},
				},
			},
		},
	}.Run(owner)
}
