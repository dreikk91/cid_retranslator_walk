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
	"github.com/lxn/win"
)

// AppContext тримає залежності для UI
type AppContext struct {
	Retranslator *core.App
	Adapter      *adapters.Adapter
}

type MainWindowWithStats struct {
	Window          *walk.MainWindow
	StatsIndicators *StatsIndicators
	NotifyIcon      *walk.NotifyIcon
	Config          *config.Config
}

func CreateMainWindow(ppkModel *models.PPKModel, eventModel *models.EventModel, appCtx *AppContext, statsData *models.StatsData, cfg *config.Config) *MainWindowWithStats {
	var mw *walk.MainWindow
	var tabWidget *walk.TabWidget
	var ppkTableView *walk.TableView
	var eventTableView *walk.TableView
	var notifyIcon *walk.NotifyIcon

	// Створюємо індикатори зі зв'язком з моделлю статистики
	statsIndicators := NewStatsIndicators()

	err := MainWindow{
		AssignTo: &mw,
		Title:    "CID Ретранслятор - Система моніторингу",
		MinSize:  Size{Width: 600, Height: 480},
		Size:     Size{Width: 800, Height: 480},
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

	// Створюємо іконку в системному треї
	icon, err := walk.Resources.Icon("APPICON")
	if err != nil {
		// Спроба 2: завантажити з файлу icon.ico
		icon, err = walk.NewIconFromFile("icon.ico")
		if err != nil {
			// Спроба 3: використовуємо стандартну іконку Windows
			icon, err = walk.NewIconFromResourceId(32512) // IDI_APPLICATION
			if err != nil {
				slog.Warn("Failed to load icon from all sources", "error", err)
				icon = nil
			} else {
				slog.Info("Using default Windows icon")
			}
		} else {
			slog.Info("Icon loaded from file icon.ico")
		}
	} else {
		slog.Info("Icon loaded from embedded resources")
	}

	notifyIcon, err = walk.NewNotifyIcon(mw)
	if err != nil {
		slog.Warn("Failed to create notify icon", "error", err)
	} else {
		// Встановлюємо іконку тільки якщо вона успішно створена
		if icon != nil {
			if err := notifyIcon.SetIcon(icon); err != nil {
				slog.Warn("Failed to set notify icon", "error", err)
			}
		}
		if err := notifyIcon.SetToolTip("CID Ретранслятор"); err != nil {
			slog.Warn("Failed to set notify icon tooltip", "error", err)
		}

		// Створюємо контекстне меню для іконки в треї
		if err := notifyIcon.SetVisible(true); err != nil {
			slog.Warn("Failed to show notify icon", "error", err)
		}

		// Подвійний клік на іконці - показати/сховати вікно
		notifyIcon.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
			if button == walk.LeftButton {
				toggleWindowVisibility(mw)
			}
		})

		// Контекстне меню
		showHideAction := walk.NewAction()
		showHideAction.SetText("Показати/Сховати")
		showHideAction.Triggered().Attach(func() {
			toggleWindowVisibility(mw)
		})

		exitAction := walk.NewAction()
		exitAction.SetText("Вихід")
		exitAction.Triggered().Attach(func() {
			mw.Close()
		})

		notifyIcon.ContextMenu().Actions().Add(showHideAction)
		notifyIcon.ContextMenu().Actions().Add(exitAction)
	}

	// Обробка події закриття вікна
	mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if cfg.UI.CloseToTray && reason == walk.CloseReasonUser {
			*canceled = true
			mw.Hide()
		}
	})

	// Обробка події мінімізації вікна
	if cfg.UI.MinimizeToTray {
		mw.VisibleChanged().Attach(func() {
			if !mw.Visible() {
				// Вікно приховане
				return
			}
		})

		// Для обробки мінімізації використовуємо StateChanged
		// Walk не має прямого обробника для мінімізації, тому використовуємо SizeChanged
		mw.SizeChanged().Attach(func() {
			// Перевіряємо чи вікно мінімізоване через WinAPI
			if isMinimized(mw) && cfg.UI.MinimizeToTray {
				mw.Hide()
			}
		})
	}

	// КРИТИЧНО! Встановлюємо tableView ПІСЛЯ Create()
	ppkModel.SetTableView(ppkTableView)
	eventModel.SetTableView(eventTableView)

	// Запускаємо автооновлення таблиці ППК для відображення таймаутів
	StartPPKRefresh(ppkTableView)

	slog.Info("MainWindow created",
		"ppkTableView", ppkTableView != nil,
		"eventTableView", eventTableView != nil)

	result := &MainWindowWithStats{
		Window:          mw,
		StatsIndicators: statsIndicators,
		NotifyIcon:      notifyIcon,
		Config:          cfg,
	}

	// Якщо потрібно запустити згорнутим
	if cfg.UI.StartMinimized {
		mw.Hide()
	}

	return result
}

func (mw *MainWindowWithStats) UpdateStats(snap metrics.Snapshot) {
	mw.StatsIndicators.Update(snap)
}

// toggleWindowVisibility показує/ховає головне вікно
func toggleWindowVisibility(mw *walk.MainWindow) {
	if mw.Visible() {
		mw.Hide()
	} else {
		mw.Show()
		// Відновлюємо вікно якщо воно було згорнуте
		restoreWindow(mw)
	}
}

// restoreWindow відновлює вікно з мінімізованого стану
func restoreWindow(mw *walk.MainWindow) {
	mw.Synchronize(func() {
		// Використовуємо Windows API для відновлення вікна
		hwnd := mw.Handle()
		const SW_RESTORE = 9
		win.ShowWindow(hwnd, SW_RESTORE)
		win.SetForegroundWindow(hwnd)
	})
}

// isMinimized перевіряє чи вікно мінімізоване
func isMinimized(mw *walk.MainWindow) bool {
	hwnd := mw.Handle()
	return win.IsIconic(hwnd)
}
