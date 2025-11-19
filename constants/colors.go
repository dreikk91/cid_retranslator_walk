package constants

import "github.com/lxn/walk"

// Windows 11 Color Scheme
// Основні кольори інтерфейсу в стилі Windows 11
var (
	// Background colors
	Win11Background     = walk.RGB(243, 243, 243) // Світло-сірий фон (#F3F3F3)
	Win11CardBackground = walk.RGB(255, 255, 255) // Білі картки (#FFFFFF)
	Win11Border         = walk.RGB(229, 229, 229) // Тонкі бордери (#E5E5E5)
	Win11Divider        = walk.RGB(237, 237, 237) // Розділювачі (#EDEDED)

	// Text colors
	Win11TextPrimary   = walk.RGB(31, 31, 31)    // Основний текст (#1F1F1F)
	Win11TextSecondary = walk.RGB(96, 96, 96)    // Вторинний текст (#606060)
	Win11TextDisabled  = walk.RGB(161, 161, 161) // Неактивний текст (#A1A1A1)

	// Accent colors (Windows 11 blue)
	Win11Accent      = walk.RGB(0, 120, 212)   // Акцентний синій (#0078D4)
	Win11AccentLight = walk.RGB(204, 228, 247) // Світлий акцент (#CCE4F7)
	Win11AccentDark  = walk.RGB(0, 99, 177)    // Темний акцент (#0063B1)

	// Status colors
	Win11Success = walk.RGB(16, 124, 16) // Зелений успіх (#107C10)
	Win11Error   = walk.RGB(196, 43, 28) // Червона помилка (#C42B1C)
	Win11Warning = walk.RGB(255, 185, 0) // Жовте попередження (#FFB900)
	Win11Info    = walk.RGB(0, 120, 212) // Синя інформація (#0078D4)
)

// Кольори для статусів ППК (оновлені під Windows 11)
var (
	ColorGreen  = walk.RGB(16, 124, 16)   // Активний (Win11 Success)
	ColorRed    = walk.RGB(196, 43, 28)   // Помилка (Win11 Error)
	ColorOrange = walk.RGB(255, 185, 0)   // Попередження (Win11 Warning)
	ColorBlue   = walk.RGB(0, 120, 212)   // Інформація (Win11 Accent)
	ColorWhite  = walk.RGB(255, 255, 255) // Білий текст
	ColorBlack  = walk.RGB(31, 31, 31)    // Чорний текст (Win11 style)
	ColorGray   = walk.RGB(250, 250, 250) // Світлий сірий фон (#FAFAFA)
)

// Кольори для подій (оновлені під Windows 11)
var (
	EventCriticalBg   = walk.RGB(253, 231, 233) // Світло-червоний фон (#FDE7E9)
	EventCriticalText = walk.RGB(196, 43, 28)   // Червоний текст (#C42B1C)
	EventErrorBg      = walk.RGB(253, 237, 237) // Дуже світло-червоний (#FDEDED)
	EventWarningBg    = walk.RGB(255, 244, 206) // Світло-жовтий (#FFF4CE)
	EventInfoBg       = walk.RGB(243, 250, 255) // Світло-синій (#F3FAFF)
	EventSuccessBg    = walk.RGB(223, 246, 221) // Світло-зелений (#DFF6DD)
)

var (
	UnknownEventBG   = walk.RGB(243, 243, 243)
	UnknownEventText = walk.RGB(96, 96, 96)

	GuardEventBG   = walk.RGB(232, 244, 255) // Світло-синій (#E8F4FF)
	GuardEventText = walk.RGB(0, 99, 177)

	DisguardEventBG   = walk.RGB(223, 246, 221) // Світло-зелений (#DFF6DD)
	DisguardEventText = walk.RGB(16, 124, 16)

	OkEventBG   = walk.RGB(255, 251, 230) // Світло-жовтий (#FFFBE6)
	OkEventText = walk.RGB(148, 101, 0)

	AlarmEventBG   = walk.RGB(253, 231, 233) // Світло-червоний (#FDE7E9)
	AlarmEventText = walk.RGB(196, 43, 28)

	OtherEventBG   = walk.RGB(250, 250, 250)
	OtherEventText = walk.RGB(31, 31, 31)
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

// Modern design colors (updated for Windows 11 style)
var (
	ModernBgPrimary   = walk.RGB(243, 243, 243) // Win11 background
	ModernBgCard      = walk.RGB(255, 255, 255) // White cards
	ModernBorder      = walk.RGB(229, 229, 229) // Light borders
	StatAcceptedBg    = walk.RGB(223, 246, 221) // Light green
	StatAcceptedText  = walk.RGB(16, 124, 16)   // Green text
	StatRejectedBg    = walk.RGB(253, 237, 237) // Light red
	StatRejectedText  = walk.RGB(196, 43, 28)   // Red text
	StatReconnectBg   = walk.RGB(255, 244, 206) // Light yellow
	StatReconnectText = walk.RGB(148, 101, 0)   // Dark yellow
	StatUptimeBg      = walk.RGB(232, 244, 255) // Light blue
	StatUptimeText    = walk.RGB(0, 99, 177)    // Blue text
	TabActive         = walk.RGB(0, 120, 212)   // Win11 accent
	TabInactive       = walk.RGB(96, 96, 96)    // Gray
	TabActiveBg       = walk.RGB(243, 250, 255) // Very light blue
	TextPrimary       = walk.RGB(31, 31, 31)    // Dark gray
	TextSecondary     = walk.RGB(96, 96, 96)    // Medium gray
)

// Windows 11 Status Bar Colors
var (
	StatusBarBg        = walk.RGB(249, 249, 249) // Very light gray (#F9F9F9)
	StatusBarText      = walk.RGB(31, 31, 31)    // Dark text (#1F1F1F)
	StatusBarBorder    = walk.RGB(229, 229, 229) // Light border (#E5E5E5)
	StatusBarSeparator = walk.RGB(237, 237, 237) // Divider (#EDEDED)

	// Status indicator colors (Windows 11 semantic colors)
	StatusConnectedBg    = walk.RGB(16, 124, 16) // Win11 Success green
	StatusDisconnectedBg = walk.RGB(196, 43, 28) // Win11 Error red

	// Counter colors (Windows 11 semantic colors)
	CounterAcceptedIcon  = walk.RGB(16, 124, 16) // Success green
	CounterRejectedIcon  = walk.RGB(196, 43, 28) // Error red
	CounterReconnectIcon = walk.RGB(202, 80, 16) // Orange (#CA5010)
	CounterUptimeIcon    = walk.RGB(0, 120, 212) // Accent blue
)
