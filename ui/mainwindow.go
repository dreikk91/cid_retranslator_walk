package ui

import (
	"cid_retranslator_walk/adapters"
	"cid_retranslator_walk/core"
	"cid_retranslator_walk/models"
	"log/slog"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// AppContext тримає залежності для UI
type AppContext struct {
	Retranslator *core.App
	Adapter      *adapters.Adapter
}

type MainWindowWithStats struct {
	Window          *walk.MainWindow
	StatsIndicators *StatsIndicators
}

func CreateMainWindow(ppkModel *models.PPKModel, eventModel *models.EventModel, appCtx *AppContext, statsData *models.StatsData,) *MainWindowWithStats{
	var mw *walk.MainWindow
	var tabWidget *walk.TabWidget
	var ppkTableView *walk.TableView
	var eventTableView *walk.TableView

	// Створюємо індикатори зі зв'язком з моделлю статистики
	statsIndicators := NewStatsIndicators(statsData)

	err := MainWindow{
		AssignTo: &mw,
		Title:    "Система моніторингу (модульна)",
		MinSize:  Size{Width: 600, Height: 400},
		Layout:   VBox{},
		
		Children: []Widget{
			statsIndicators.CreateIndicators(),
			TabWidget{
				AssignTo: &tabWidget,
				Pages: []TabPage{
					CreatePPKTab(ppkModel, &ppkTableView, &mw, appCtx),
					CreateEventsTab(eventModel, &eventTableView),
					CreateSettingsTab(),
				},
			},
		},
	}.Create()

	if err != nil {
		panic(err)
	}

	// КРИТИЧНО! Встановлюємо tableView ПІСЛЯ Create()
	ppkModel.SetTableView(ppkTableView)
	eventModel.SetTableView(eventTableView)

	slog.Info("MainWindow created",
		"ppkTableView", ppkTableView != nil,
		"eventTableView", eventTableView != nil)

	return &MainWindowWithStats{
		Window:          mw,
		StatsIndicators: statsIndicators,
	}
}
