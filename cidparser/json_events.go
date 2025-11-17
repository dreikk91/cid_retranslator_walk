package cidparser


import (
	"encoding/json"
	"fmt"
)

// Event представляє структуру події з JSON
type Event struct {
	ContactIdCode  string `json:"contactId_code"`
	TypeCodeMesUK  string `json:"TypeCodeMes_UK"`
	CodeMesUK      string `json:"CodeMes_UK"`
	Zoneno         int    `json:"Zoneno"`
	AccessCode     int    `json:"AccessCode"`
	GroupSent      int    `json:"GroupSent"`
	AutoReset      int    `json:"AutoReset"`
}

// EventMap — мапа для швидкого пошуку за contactId_code
type EventMap map[string]*Event

// LoadEvents завантажує JSON-масив подій і повертає мапу для пошуку
func LoadEvents(jsonData []byte) (EventMap, error) {
	var events []Event
	if err := json.Unmarshal(jsonData, &events); err != nil {
		return nil, fmt.Errorf("помилка парсингу JSON: %w", err)
	}

	eventMap := make(EventMap, len(events))
	for _, ev := range events {
		eventMap[ev.ContactIdCode] = &ev
	}
	return eventMap, nil
}

// GetEventDescriptions повертає TypeCodeMes_UK та CodeMes_UK за contactId_code
func (em EventMap) GetEventDescriptions(code string) (typeDesc, codeDesc string, found bool) {
	ev, ok := em[code]
	if !ok {
		return "", "", false
	}
	return ev.TypeCodeMesUK, ev.CodeMesUK, true
}

