package ui

import (
	. "github.com/lxn/walk/declarative"
)

func CreateSettingsTab() TabPage {
	return TabPage{
		Title:  "Налаштування",
		Layout: VBox{},
		Children: []Widget{
			Label{
				Text: "Налаштування будуть додані пізніше",
				Font: Font{PointSize: 10},
			},
			VSpacer{},
		},
	}
}
