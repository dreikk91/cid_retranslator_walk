package ui

import (
	"cid_retranslator_walk/adapters"
	"cid_retranslator_walk/config"
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

func CreateMainWindow(ppkModel *models.PPKModel, eventModel *models.EventModel, appCtx *AppContext, statsData *models.StatsData, cfg *config.Config) *MainWindowWithStats {
	var mw *walk.MainWindow
	var tabWidget *walk.TabWidget
	var ppkTableView *walk.TableView
	var eventTableView *walk.TableView

	// Створюємо індикатори зі зв'язком з моделлю статистики
	statsIndicators := NewStatsIndicators(statsData)

	err := MainWindow{
		AssignTo: &mw,
		Title:    "CID Ретранслятор - Система моніторингу",
		MinSize:  Size{Width: 800, Height: 600},
		Size:     Size{Width: 1024, Height: 768},
		Font:     Font{Family: "Segoe UI", PointSize: 10},
		Layout:   VBox{Margins: Margins{Left: 10, Top: 10, Right: 10, Bottom: 10}},

		Children: []Widget{
			statsIndicators.CreateIndicators(),
			TabWidget{
				AssignTo: &tabWidget,
				Pages: []TabPage{
					CreatePPKTab(ppkModel, &ppkTableView, &mw, appCtx),
					CreateEventsTab(eventModel, &eventTableView),
					CreateSettingsTab(cfg),
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
