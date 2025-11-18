package constants

import "github.com/lxn/walk"

// Кольори для статусів ППК
var (
	ColorGreen  = walk.RGB(76, 175, 80)   // Активний
	ColorRed    = walk.RGB(244, 67, 54)   // Помилка
	ColorOrange = walk.RGB(255, 152, 0)   // Попередження
	ColorBlue   = walk.RGB(33, 150, 243)  // Інформаціяcolor.NRGBA
	ColorWhite  = walk.RGB(255, 255, 255) // Білий текст
	ColorBlack  = walk.RGB(0, 0, 0)       // Чорний текст
	ColorGray   = walk.RGB(240, 240, 240) // Сірий фон
)

// Кольори для подій
var (
	EventCriticalBg   = walk.RGB(183, 28, 28)   // Критична помилка - фон
	EventCriticalText = walk.RGB(255, 205, 210) // Критична помилка - текст
	EventErrorBg      = walk.RGB(239, 154, 154) // Помилка - фон
	EventWarningBg    = walk.RGB(255, 235, 59)  // Попередження - фон
	EventInfoBg       = walk.RGB(100, 181, 246) // Інформація - фон
	EventSuccessBg    = walk.RGB(129, 199, 132) // Успіх - фон
)

var (
	UnknownEventBG   = walk.RGB(156, 163, 175)
	UnknownEventText = walk.RGB(0, 0, 0)

	GuardEventBG   = walk.RGB(41, 128, 185)
	GuardEventText = walk.RGB(236, 240, 241)

	DisguardEventBG   = walk.RGB(39, 174, 96)
	DisguardEventText = walk.RGB(236, 240, 241)

	OkEventBG   = walk.RGB(241, 196, 15)
	OkEventText = walk.RGB(44, 62, 80)

	AlarmEventBG   = walk.RGB(231, 76, 60)
	AlarmEventText = walk.RGB(236, 240, 241)

	OtherEventBG   = walk.RGB(236, 240, 241)
	OtherEventText = walk.RGB(44, 62, 80)
)

var (
	UnknownEvent  = 0
	GuardEvent    = 1
	DisguardEvent = 3
	OkEvent       = 3
	AlarmEvent    = 4
	OtherEvent    = 5
)

// Пріоритети подій
const (
	PriorityCritical = 0
	PriorityError    = 1
	PriorityWarning  = 2
	PriorityInfo     = 3
	PrioritySuccess  = 4
)

var (
	GreenBrush, _ = walk.NewSolidColorBrush(ColorGreen)
	RedBrush, _   = walk.NewSolidColorBrush(ColorRed)
)


var (
	ModernBgPrimary   = walk.RGB( 248,  250,  252)
	ModernBgCard      = walk.RGB( 255,  255,  255)
	ModernBorder      = walk.RGB( 226,  232,  240)
	StatAcceptedBg    = walk.RGB( 209,  250,  229)
	StatAcceptedText  = walk.RGB( 22,  101,  52)
	StatRejectedBg    = walk.RGB( 254,  226,  226)
	StatRejectedText  = walk.RGB( 127,  29,  29)
	StatReconnectBg   = walk.RGB( 254,  243,  199)
	StatReconnectText = walk.RGB( 120,  53,  15)
	StatUptimeBg      = walk.RGB( 219,  234,  254)
	StatUptimeText    = walk.RGB( 30,  58,  138)
	TabActive         = walk.RGB( 59,  130,  246)
	TabInactive       = walk.RGB( 148,  163,  184)
	TabActiveBg       = walk.RGB( 239,  246,  255)
	TextPrimary       = walk.RGB( 15,  23,  42)
	TextSecondary     = walk.RGB( 100,  116,  139)

)