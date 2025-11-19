package ui

import (
	"cid_retranslator_walk/adapters"
	"cid_retranslator_walk/config"
	"cid_retranslator_walk/core"
	"cid_retranslator_walk/metrics"
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
	statsIndicators := NewStatsIndicators()

	err := MainWindow{
		AssignTo: &mw,
		Title:    "CID Ретранслятор - Система моніторингу",
		MinSize:  Size{Width: 900, Height: 650},
		Size:     Size{Width: 1100, Height: 800},
		Font:     Font{Family: "Segoe UI", PointSize: 9},
		Layout:   VBox{MarginsZero: true, SpacingZero: true},

		Children: []Widget{
			TabWidget{
				AssignTo: &tabWidget,
				Pages: []TabPage{
					CreatePPKTab(ppkModel, &ppkTableView, &mw, appCtx, cfg),
					CreateEventsTab(eventModel, &eventTableView),
					CreateSettingsTab(cfg),
				},
			},
			statsIndicators.CreateIndicators(),
		},
	}.Create()

	if err != nil {
		panic(err)
	}

	// КРИТИЧНО! Встановлюємо tableView ПІСЛЯ Create()
	ppkModel.SetTableView(ppkTableView)
	eventModel.SetTableView(eventTableView)

	// Запускаємо автооновлення таблиці ППК для відображення таймаутів
	StartPPKRefresh(ppkTableView)

	slog.Info("MainWindow created",
		"ppkTableView", ppkTableView != nil,
		"eventTableView", eventTableView != nil)

	return &MainWindowWithStats{
		Window:          mw,
		StatsIndicators: statsIndicators,
	}
}

func (mw *MainWindowWithStats) UpdateStats(snap metrics.Snapshot) {
	mw.StatsIndicators.Update(snap)
}
