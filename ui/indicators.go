package ui

import (
	"cid_retranslator_walk/constants"
	"cid_retranslator_walk/metrics"
	"fmt"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// StatsIndicators зберігає посилання на Label'и для оновлення
type StatsIndicators struct {
	StatusLabel     *walk.Label
	AcceptedLabel   *walk.Label
	RejectedLabel   *walk.Label
	ReconnectsLabel *walk.Label
	UptimeLabel     *walk.Label
	StatusIcon      *walk.Label
}

// NewStatsIndicators створює індикатори з прив'язкою до моделі
func NewStatsIndicators() *StatsIndicators {
	return &StatsIndicators{}
}

// CreateIndicators створює композит з індикаторами у вигляді статус-бару внизу
func (si *StatsIndicators) CreateIndicators() Composite {
	return Composite{
		Background: SolidColorBrush{Color: constants.StatusBarBg},
		MaxSize:    Size{Height: 36},
		MinSize:    Size{Height: 36},
		Layout:     HBox{Margins: Margins{Left: 12, Top: 6, Right: 12, Bottom: 6}, Spacing: 0},
		Children: []Widget{
			si.createStatusIndicator(),
			si.createSeparator(),
			si.createAcceptedIndicator(),
			si.createSeparator(),
			si.createRejectedIndicator(),
			si.createSeparator(),
			si.createReconnectsIndicator(),
			si.createSeparator(),
			si.createUptimeIndicator(),
			HSpacer{},
		},
	}
}

func (si *StatsIndicators) createSeparator() Composite {
	return Composite{
		MaxSize:    Size{Width: 1, Height: 24},
		MinSize:    Size{Width: 1, Height: 24},
		Layout:     HBox{MarginsZero: true},
		Background: SolidColorBrush{Color: constants.StatusBarSeparator},
	}
}

func (si *StatsIndicators) createStatusIndicator() Composite {
	return Composite{
		Layout:  HBox{Margins: Margins{Left: 0, Top: 0, Right: 12, Bottom: 0}, Spacing: 6},
		MinSize: Size{Width: 140},
		Children: []Widget{
			Label{
				AssignTo:  &si.StatusIcon,
				Text:      "●",
				TextColor: constants.StatusConnectedBg,
				Font:      Font{PointSize: 11, Bold: true},
			},
			Label{
				AssignTo:  &si.StatusLabel,
				Text:      "Підключення...",
				TextColor: constants.StatusBarText,
				Font:      Font{PointSize: 9},
			},
		},
	}
}

func (si *StatsIndicators) createAcceptedIndicator() Composite {
	return Composite{
		Layout:  HBox{Margins: Margins{Left: 12, Top: 0, Right: 12, Bottom: 0}, Spacing: 6},
		MinSize: Size{Width: 85},
		Children: []Widget{
			Label{
				Text:      "✓",
				TextColor: constants.CounterAcceptedIcon,
				Font:      Font{PointSize: 10, Bold: true},
			},
			Label{
				AssignTo:  &si.AcceptedLabel,
				Text:      "0",
				TextColor: constants.StatusBarText,
				Font:      Font{PointSize: 9},
			},
		},
	}
}

func (si *StatsIndicators) createRejectedIndicator() Composite {
	return Composite{
		Layout:  HBox{Margins: Margins{Left: 12, Top: 0, Right: 12, Bottom: 0}, Spacing: 6},
		MinSize: Size{Width: 85},
		Children: []Widget{
			Label{
				Text:      "✗",
				TextColor: constants.CounterRejectedIcon,
				Font:      Font{PointSize: 10, Bold: true},
			},
			Label{
				AssignTo:  &si.RejectedLabel,
				Text:      "0",
				TextColor: constants.StatusBarText,
				Font:      Font{PointSize: 9},
			},
		},
	}
}

func (si *StatsIndicators) createReconnectsIndicator() Composite {
	return Composite{
		Layout:  HBox{Margins: Margins{Left: 12, Top: 0, Right: 12, Bottom: 0}, Spacing: 6},
		MinSize: Size{Width: 100},
		Children: []Widget{
			Label{
				Text:      "↻",
				TextColor: constants.CounterReconnectIcon,
				Font:      Font{PointSize: 10, Bold: true},
			},
			Label{
				AssignTo:  &si.ReconnectsLabel,
				Text:      "Повтори: 0",
				TextColor: constants.StatusBarText,
				Font:      Font{PointSize: 9},
			},
		},
	}
}

func (si *StatsIndicators) createUptimeIndicator() Composite {
	return Composite{
		Layout:  HBox{Margins: Margins{Left: 12, Top: 0, Right: 0, Bottom: 0}, Spacing: 6},
		MinSize: Size{Width: 120},
		Children: []Widget{
			Label{
				Text:      "⏱",
				TextColor: constants.CounterUptimeIcon,
				Font:      Font{PointSize: 10, Bold: true},
			},
			Label{
				AssignTo:  &si.UptimeLabel,
				Text:      "00:00:00",
				TextColor: constants.StatusBarText,
				Font:      Font{PointSize: 9},
			},
		},
	}
}

// Update оновлює всі індикатори (викликається з UI-потоку через Synchronize)
func (si *StatsIndicators) Update(snap metrics.Snapshot) {
	accepted := snap.Accepted
	rejected := snap.Rejected
	reconnects := snap.Reconnects
	uptime := snap.UptimeString()
	status := snap.Connected

	// Оновлюємо статус підключення
	if status {
		si.StatusLabel.SetText("Підключено")
		si.StatusIcon.SetTextColor(constants.StatusConnectedBg)
	} else {
		si.StatusLabel.SetText("Відключено")
		si.StatusIcon.SetTextColor(constants.StatusDisconnectedBg)
	}

	// Оновлюємо лічильники з більш компактним форматом
	si.AcceptedLabel.SetText(fmt.Sprintf("%d", accepted))
	si.RejectedLabel.SetText(fmt.Sprintf("%d", rejected))
	si.ReconnectsLabel.SetText(fmt.Sprintf("Повтори: %d", reconnects))
	si.UptimeLabel.SetText(uptime)
}
