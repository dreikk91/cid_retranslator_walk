package ui

import (
	"cid_retranslator_walk/constants"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func CreateIndicators() Composite {
	return Composite{
		Layout: HBox{Margins: Margins{Left: 10, Top: 0, Right: 10, Bottom: 0}},
		Children: []Widget{
			createIndicator("Прийнято:", constants.ColorGreen),
			createIndicator("Відхилено:", constants.ColorRed),
			createIndicator("Перепідключень:", constants.ColorOrange),
			createIndicator("Аптайм:", constants.ColorBlue),
		},
	}
}

func createIndicator(text string, bgColor walk.Color) Composite {
	return Composite{
		Layout:     HBox{},
		Background: SolidColorBrush{Color: bgColor},
		MinSize:    Size{Width: 60, Height: 30},
		Children: []Widget{
			HSpacer{},
			Label{
				Text:      text,
				TextColor: constants.ColorWhite,
				Font:      Font{PointSize: 9, Bold: true},
			},
			HSpacer{},
		},
	}
}
