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
}

// NewStatsIndicators створює індикатори з прив'язкою до моделі
func NewStatsIndicators() *StatsIndicators {
	return &StatsIndicators{}
}

// CreateIndicators створює композит з індикаторами
func (si *StatsIndicators) CreateIndicators() Composite {
	return Composite{
		Layout: HBox{Margins: Margins{Left: 5, Top: 5, Right: 5, Bottom: 5}, Spacing: 8},
		Children: []Widget{
			si.createStatusIndicator(),
			si.createAcceptedIndicator(),
			si.createRejectedIndicator(),
			si.createReconnectsIndicator(),
			si.createUptimeIndicator(),
		},
	}
}

func (si *StatsIndicators) createStatusIndicator() Composite {
	return Composite{
		Layout:     HBox{Margins: Margins{Left: 8, Top: 4, Right: 8, Bottom: 4}},
		Background: SolidColorBrush{Color: constants.ColorGreen},
		MinSize:    Size{Width: 150, Height: 25},
		Children: []Widget{
			HSpacer{},
			Label{
				AssignTo:  &si.StatusLabel,
				Text:      "Статус: Підключення...",
				TextColor: constants.ColorWhite,
				Font:      Font{PointSize: 9, Bold: true},
			},
			HSpacer{},
		},
	}
}

func (si *StatsIndicators) createAcceptedIndicator() Composite {
	return Composite{
		Layout:     HBox{Margins: Margins{Left: 8, Top: 4, Right: 8, Bottom: 4}},
		Background: SolidColorBrush{Color: constants.StatAcceptedBg},
		MinSize:    Size{Width: 130, Height: 25},
		Children: []Widget{
			HSpacer{},
			Label{
				AssignTo:  &si.AcceptedLabel,
				Text:      "Прийнято: 0",
				TextColor: constants.StatAcceptedText,
				Font:      Font{PointSize: 9, Bold: true},
			},
			HSpacer{},
		},
	}
}

func (si *StatsIndicators) createRejectedIndicator() Composite {
	return Composite{
		Layout:     HBox{Margins: Margins{Left: 8, Top: 4, Right: 8, Bottom: 4}},
		Background: SolidColorBrush{Color: constants.StatRejectedBg},
		MinSize:    Size{Width: 130, Height: 25},
		Children: []Widget{
			HSpacer{},
			Label{
				AssignTo:  &si.RejectedLabel,
				Text:      "Відхилено: 0",
				TextColor: constants.StatRejectedText,
				Font:      Font{PointSize: 9, Bold: true},
			},
			HSpacer{},
		},
	}
}

func (si *StatsIndicators) createReconnectsIndicator() Composite {
	return Composite{
		Layout:     HBox{Margins: Margins{Left: 8, Top: 4, Right: 8, Bottom: 4}},
		Background: SolidColorBrush{Color: constants.StatReconnectBg},
		MinSize:    Size{Width: 160, Height: 25},
		Children: []Widget{
			HSpacer{},
			Label{
				AssignTo:  &si.ReconnectsLabel,
				Text:      "Переподключення: 0",
				TextColor: constants.StatReconnectText,
				Font:      Font{PointSize: 9, Bold: true},
			},
			HSpacer{},
		},
	}
}

func (si *StatsIndicators) createUptimeIndicator() Composite {
	return Composite{
		Layout:     HBox{Margins: Margins{Left: 8, Top: 4, Right: 8, Bottom: 4}},
		Background: SolidColorBrush{Color: constants.StatUptimeBg},
		MinSize:    Size{Width: 150, Height: 25},
		Children: []Widget{
			HSpacer{},
			Label{
				AssignTo:  &si.UptimeLabel,
				Text:      "Uptime: 00:00:00",
				TextColor: constants.StatUptimeText,
				Font:      Font{PointSize: 9, Bold: true},
			},
			HSpacer{},
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
		si.StatusLabel.SetText("Статус: Підключено")
		green, _ := walk.NewSolidColorBrush(constants.ColorGreen)
		si.StatusLabel.Parent().SetBackground(green)
	} else {
		si.StatusLabel.SetText("Статус: Відключено")
		red, _ := walk.NewSolidColorBrush(constants.ColorRed)
		si.StatusLabel.Parent().SetBackground(red)
	}

	// Оновлюємо лічильники
	si.AcceptedLabel.SetText(fmt.Sprintf("Прийнято: %d", accepted))
	si.RejectedLabel.SetText(fmt.Sprintf("Відхилено: %d", rejected))
	si.ReconnectsLabel.SetText(fmt.Sprintf("Переподключення: %d", reconnects))
	si.UptimeLabel.SetText(fmt.Sprintf("Uptime: %s", uptime))
}
