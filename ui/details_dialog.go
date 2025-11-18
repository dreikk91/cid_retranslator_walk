package ui

import (
	"cid_retranslator_walk/constants"
	"cid_retranslator_walk/models"
	"fmt"
	"log/slog"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func ShowPPKDetails(
	owner walk.Form,
	ppkItem *models.PPKItem,
	appCtx *AppContext,
) {
	var dlg *walk.Dialog
	var tableView *walk.TableView

	model := models.NewDetailModel(ppkItem.Name)

	Dialog{
		AssignTo: &dlg,
		Title:    fmt.Sprintf("Деталі: %s", ppkItem.Name),
		MinSize:  Size{Width: 800, Height: 500},
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "Номер:", Font: Font{Bold: true}},
					Label{Text: fmt.Sprintf("%d", ppkItem.Number)},
					Label{Text: "Назва:", Font: Font{Bold: true}},
					Label{Text: ppkItem.Name},
					Label{Text: "Остання подія:", Font: Font{Bold: true}},
					Label{Text: ppkItem.Event},
				},
			},
			Label{
				Text: "Останні події:",
				Font: Font{PointSize: 10, Bold: true},
			},
			TableView{
				AssignTo:         &tableView,
				AlternatingRowBG: true,
				ColumnsOrderable: true,
				Model:            model,
				Columns: []TableViewColumn{
					{Title: "Час", Width: 130},
					{Title: "ППК", Width: 60},
					{Title: "Код", Width: 60},
					{Title: "Тип", Width: 120},
					{Title: "Опис", Width: 200},
					{Title: "Зона|Група", Width: 120},
				},
				StyleCell: func(style *walk.CellStyle) {
					item := model.GetItem(style.Row())
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
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "Закрити",
						OnClicked: func() {
							model.Stop()
							appCtx.Retranslator.CloseDeviceEventChannel(ppkItem.Number)
							dlg.Accept()
						},
					},
				},
			},
		},
	}.Create(owner)

	// Встановлюємо tableView в модель
	model.SetTableView(tableView)

	// Запускаємо слухання каналу
	model.StartListening()

	// Завантажуємо початкові події
	go func() {
		events := appCtx.Retranslator.GetDeviceEvents(ppkItem.Number)
		for _, ev := range events {
			if len(ev.Data) < 20 {
				continue
			}

			devID := ev.Data[7:11]
			code := ev.Data[11:15]
			group := ev.Data[15:17]
			zone := ev.Data[17:20]

			eventType, desc, found := appCtx.Adapter.EventMap.GetEventDescriptions(code)
			if !found {
				continue
			}

			priority, eventType := appCtx.Adapter.DetermineEventPriority(code, eventType)

			uiEvent := &models.DetailItem{
				Time:     ev.Time,
				Device:   fmt.Sprint(devID),
				Code:     code,
				Type:     eventType,
				Desc:     desc,
				Zone:     fmt.Sprintf("Зона %s|Група %s", zone, group),
				Priority: priority,
			}

			select {
			case model.GetChannel() <- uiEvent:
			default:
				slog.Warn("UI detail channel full during initial load")
			}
		}
		slog.Info("Initial device events loaded", "deviceID", ppkItem.Number, "count", len(events))
	}()

	// Отримуємо канал для нових подій
	deviceEventChan := appCtx.Retranslator.GetDeviceEventChannel(ppkItem.Number)

	// Запускаємо стрімінг нових подій
	go appCtx.Adapter.StreamDeviceEventsToUI(
		ppkItem.Number,
		deviceEventChan,
		model.GetChannel(),
		model.GetStopChannel(),
	)

	slog.Info("PPK details dialog opened", "deviceID", ppkItem.Number)

	// Відкриваємо діалог (блокуюче)
	dlg.Run()

	slog.Info("PPK details dialog closed", "deviceID", ppkItem.Number)
}
