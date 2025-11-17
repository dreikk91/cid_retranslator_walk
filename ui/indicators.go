package ui

import (
	"cid_retranslator_walk/constants"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func CreateIndicators() Composite {
	return Composite{
		Layout: HBox{Margins: Margins{Left: 10, Top: 10, Right: 10, Bottom: 10}},
		Children: []Widget{
			createIndicator("Статус: Активний", constants.ColorGreen),
			createIndicator("Помилки: 1", constants.ColorRed),
			createIndicator("Попередження: 1", constants.ColorOrange),
			createIndicator("Підключено: 100", constants.ColorBlue),
		},
	}
}

func createIndicator(text string, bgColor walk.Color) Composite {
	return Composite{
		Layout:     HBox{},
		Background: SolidColorBrush{Color: bgColor},
		MinSize:    Size{Width: 80, Height: 30},
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
