package ui

import (
	"cid_retranslator_walk/config"
	"fmt"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// SettingsTab holds the configuration form widgets
type SettingsTab struct {
	cfg *config.Config

	// Server fields
	serverHost *walk.LineEdit
	serverPort *walk.LineEdit

	// Client fields
	clientHost       *walk.LineEdit
	clientPort       *walk.LineEdit
	reconnectInitial *walk.LineEdit
	reconnectMax     *walk.LineEdit

	// Queue fields
	bufferSize *walk.NumberEdit

	// Logging fields
	logFilename   *walk.LineEdit
	logMaxSize    *walk.NumberEdit
	logMaxBackups *walk.NumberEdit
	logMaxAge     *walk.NumberEdit
	logCompress   *walk.CheckBox

	// CID Rules fields
	requiredPrefix *walk.LineEdit
	validLength    *walk.NumberEdit
	accNumOffset   *walk.NumberEdit
	accNumAdd      *walk.NumberEdit
}

// NewSettingsTab creates a new settings tab with the given configuration
func NewSettingsTab(cfg *config.Config) *SettingsTab {
	return &SettingsTab{cfg: cfg}
}

// CreateSettingsTab creates the settings tab page
func (st *SettingsTab) CreateSettingsTab() TabPage {
	return TabPage{
		Title:  "Налаштування",
		Layout: VBox{},
		Children: []Widget{
			Composite{
				Layout: VBox{},
				Children: []Widget{
					// Title
					Label{
						Text: "Налаштування додатку",
						Font: Font{PointSize: 12, Bold: true},
					},

					// Scrollable container
					ScrollView{
						Layout: VBox{},
						Children: []Widget{
							// Server Configuration
							GroupBox{
								Title:  "Сервер (Вхідні підключення)",
								Layout: Grid{Columns: 2},
								Children: []Widget{
									Label{Text: "Хост:"},
									LineEdit{AssignTo: &st.serverHost, Text: st.cfg.Server.Host},

									Label{Text: "Порт:"},
									LineEdit{AssignTo: &st.serverPort, Text: st.cfg.Server.Port},
								},
							},

							// Client Configuration
							GroupBox{
								Title:  "Клієнт (Вихідне підключення)",
								Layout: Grid{Columns: 2},
								Children: []Widget{
									Label{Text: "Хост:"},
									LineEdit{AssignTo: &st.clientHost, Text: st.cfg.Client.Host},

									Label{Text: "Порт:"},
									LineEdit{AssignTo: &st.clientPort, Text: st.cfg.Client.Port},

									Label{Text: "Початкова затримка перепідключення:"},
									LineEdit{
										AssignTo:    &st.reconnectInitial,
										Text:        st.cfg.Client.ReconnectInitial.String(),
										ToolTipText: "Формат: 1s, 5s, 1m тощо",
									},

									Label{Text: "Максимальна затримка перепідключення:"},
									LineEdit{
										AssignTo:    &st.reconnectMax,
										Text:        st.cfg.Client.ReconnectMax.String(),
										ToolTipText: "Формат: 1s, 5s, 1m тощо",
									},
								},
							},

							// Queue Configuration
							GroupBox{
								Title:  "Черга",
								Layout: Grid{Columns: 2},
								Children: []Widget{
									Label{Text: "Розмір буфера:"},
									NumberEdit{
										AssignTo: &st.bufferSize,
										Value:    float64(st.cfg.Queue.BufferSize),
										Decimals: 0,
										MinValue: 1,
										MaxValue: 10000,
									},
								},
							},

							// Logging Configuration
							GroupBox{
								Title:  "Логування",
								Layout: Grid{Columns: 2},
								Children: []Widget{
									Label{Text: "Файл логу:"},
									LineEdit{AssignTo: &st.logFilename, Text: st.cfg.Logging.Filename},

									Label{Text: "Максимальний розмір (МБ):"},
									NumberEdit{
										AssignTo: &st.logMaxSize,
										Value:    float64(st.cfg.Logging.MaxSize),
										Decimals: 0,
										MinValue: 1,
										MaxValue: 1000,
									},

									Label{Text: "Максимум резервних копій:"},
									NumberEdit{
										AssignTo: &st.logMaxBackups,
										Value:    float64(st.cfg.Logging.MaxBackups),
										Decimals: 0,
										MinValue: 0,
										MaxValue: 100,
									},

									Label{Text: "Максимальний вік (дні):"},
									NumberEdit{
										AssignTo: &st.logMaxAge,
										Value:    float64(st.cfg.Logging.MaxAge),
										Decimals: 0,
										MinValue: 1,
										MaxValue: 365,
									},

									Label{Text: "Стискати старі логи:"},
									CheckBox{AssignTo: &st.logCompress, Checked: st.cfg.Logging.Compress},
								},
							},

							// CID Rules Configuration
							GroupBox{
								Title:  "Правила CID",
								Layout: Grid{Columns: 2},
								Children: []Widget{
									Label{Text: "Обов'язковий префікс:"},
									LineEdit{AssignTo: &st.requiredPrefix, Text: st.cfg.CIDRules.RequiredPrefix},

									Label{Text: "Валідна довжина:"},
									NumberEdit{
										AssignTo: &st.validLength,
										Value:    float64(st.cfg.CIDRules.ValidLength),
										Decimals: 0,
										MinValue: 1,
										MaxValue: 100,
									},

									Label{Text: "Зміщення номеру акаунту:"},
									NumberEdit{
										AssignTo: &st.accNumOffset,
										Value:    float64(st.cfg.CIDRules.AccNumOffset),
										Decimals: 0,
										MinValue: 0,
										MaxValue: 10000,
									},

									Label{Text: "Додавання до номеру акаунту:"},
									NumberEdit{
										AssignTo: &st.accNumAdd,
										Value:    float64(st.cfg.CIDRules.AccNumAdd),
										Decimals: 0,
										MinValue: 0,
										MaxValue: 10000,
									},
								},
							},

							// Save and Reset buttons
							Composite{
								Layout: HBox{},
								Children: []Widget{
									HSpacer{},
									PushButton{
										Text:      "Скинути",
										MinSize:   Size{Width: 100},
										OnClicked: st.resetSettings,
									},
									PushButton{
										Text:      "Зберегти",
										MinSize:   Size{Width: 100},
										OnClicked: st.saveSettings,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// saveSettings saves the current form values to the configuration file
func (st *SettingsTab) saveSettings() {
	// Update Server config
	st.cfg.Server.Host = st.serverHost.Text()
	st.cfg.Server.Port = st.serverPort.Text()

	// Update Client config
	st.cfg.Client.Host = st.clientHost.Text()
	st.cfg.Client.Port = st.clientPort.Text()

	// Parse reconnect durations
	if reconnectInitial, err := time.ParseDuration(st.reconnectInitial.Text()); err == nil {
		st.cfg.Client.ReconnectInitial = reconnectInitial
	} else {
		walk.MsgBox(nil, "Помилка",
			fmt.Sprintf("Невірний формат початкової затримки перепідключення: %v", err),
			walk.MsgBoxIconError)
		return
	}

	if reconnectMax, err := time.ParseDuration(st.reconnectMax.Text()); err == nil {
		st.cfg.Client.ReconnectMax = reconnectMax
	} else {
		walk.MsgBox(nil, "Помилка",
			fmt.Sprintf("Невірний формат максимальної затримки перепідключення: %v", err),
			walk.MsgBoxIconError)
		return
	}

	// Update Queue config
	st.cfg.Queue.BufferSize = int(st.bufferSize.Value())

	// Update Logging config
	st.cfg.Logging.Filename = st.logFilename.Text()
	st.cfg.Logging.MaxSize = int(st.logMaxSize.Value())
	st.cfg.Logging.MaxBackups = int(st.logMaxBackups.Value())
	st.cfg.Logging.MaxAge = int(st.logMaxAge.Value())
	st.cfg.Logging.Compress = st.logCompress.Checked()

	// Update CID Rules config
	st.cfg.CIDRules.RequiredPrefix = st.requiredPrefix.Text()
	st.cfg.CIDRules.ValidLength = int(st.validLength.Value())
	st.cfg.CIDRules.AccNumOffset = int(st.accNumOffset.Value())
	st.cfg.CIDRules.AccNumAdd = int(st.accNumAdd.Value())

	// Save to file
	if err := st.cfg.Save("config.yaml"); err != nil {
		walk.MsgBox(nil, "Помилка",
			fmt.Sprintf("Не вдалося зберегти налаштування: %v", err),
			walk.MsgBoxIconError)
		return
	}

	walk.MsgBox(nil, "Успіх",
		"Налаштування успішно збережено!\n\nПерезапустіть додаток для застосування змін.",
		walk.MsgBoxIconInformation)
}

// resetSettings resets the form to the current configuration values
func (st *SettingsTab) resetSettings() {
	st.serverHost.SetText(st.cfg.Server.Host)
	st.serverPort.SetText(st.cfg.Server.Port)

	st.clientHost.SetText(st.cfg.Client.Host)
	st.clientPort.SetText(st.cfg.Client.Port)
	st.reconnectInitial.SetText(st.cfg.Client.ReconnectInitial.String())
	st.reconnectMax.SetText(st.cfg.Client.ReconnectMax.String())

	st.bufferSize.SetValue(float64(st.cfg.Queue.BufferSize))

	st.logFilename.SetText(st.cfg.Logging.Filename)
	st.logMaxSize.SetValue(float64(st.cfg.Logging.MaxSize))
	st.logMaxBackups.SetValue(float64(st.cfg.Logging.MaxBackups))
	st.logMaxAge.SetValue(float64(st.cfg.Logging.MaxAge))
	st.logCompress.SetChecked(st.cfg.Logging.Compress)

	st.requiredPrefix.SetText(st.cfg.CIDRules.RequiredPrefix)
	st.validLength.SetValue(float64(st.cfg.CIDRules.ValidLength))
	st.accNumOffset.SetValue(float64(st.cfg.CIDRules.AccNumOffset))
	st.accNumAdd.SetValue(float64(st.cfg.CIDRules.AccNumAdd))
}

// CreateSettingsTab is a helper function for backward compatibility
func CreateSettingsTab(cfg *config.Config) TabPage {
	st := NewSettingsTab(cfg)
	return st.CreateSettingsTab()
}
