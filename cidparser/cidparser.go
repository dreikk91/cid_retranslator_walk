package cidparser

import (
	"cid_retranslator_walk/config"
	"fmt"
	"log/slog"
	"strconv"
)

const (
	// Константи для CID протоколу
	requiredMessageLength = 20
	heartbeatLength       = 20
	heartbeatBody         = "           @   "
	
	// Індекси для парсингу
	accountStart  = 7
	accountEnd    = 11
	codeStart     = 11
	codeEnd       = 15
	secondStart   = 15
	secondEnd     = 20
	bodyStart     = 4
	bodyEnd       = 19
	
	// Діапазон номерів акаунтів для обробки
	accountMin = 2000
	accountMax = 2200
	
	// Діапазон кодів для heartbeat
	heartbeatCodeMin = 1000
	heartbeatCodeMax = 1999
	
	// Термінатор
	terminator = "\x14"
)

// IsMessageValid перевіряє валідність повідомлення згідно правил
func IsMessageValid(message string, rules *config.CIDRules) bool {
	if len(message) == 0 {
		return false
	}
	
	if len(message) != rules.ValidLength {
		slog.Debug("Invalid message length", 
			"got", len(message), 
			"want", rules.ValidLength)
		return false
	}
	
	if len(message) > 0 && string(message[0]) != rules.RequiredPrefix {
		slog.Debug("Invalid message prefix", 
			"got", string(message[0]), 
			"want", rules.RequiredPrefix)
		return false
	}
	
	return true
}

// ChangeAccountNumber змінює номер акаунту в повідомленні згідно правил
func ChangeAccountNumber(message []byte, rules *config.CIDRules) ([]byte, error) {
	messageString := string(message)

	if len(messageString) < requiredMessageLength {
		return nil, fmt.Errorf("invalid message length: got %d, want at least %d", 
			len(messageString), requiredMessageLength)
	}

	// Розбираємо повідомлення на частини
	firstPart := messageString[:accountStart]
	accountNumber := messageString[accountStart:accountEnd]
	messageCode := messageString[codeStart:codeEnd]
	secondPart := messageString[secondStart:secondEnd]

	// Перетворюємо номер акаунту
	num, err := strconv.Atoi(accountNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid account number %q: %w", accountNumber, err)
	}

	// Змінюємо номер тільки якщо він в діапазоні
	if num >= accountMin && num <= accountMax {
		num += rules.AccNumAdd
		slog.Debug("Account number changed", 
			"original", accountNumber, 
			"new", num)
	}

	// Форматуємо з zero-padding
	resultStr := fmt.Sprintf("%04d", num)
	
	// Застосовуємо заміну коду (якщо є)
	newMessageCode := changeTestCode(messageCode, rules)
	
	// Збираємо нове повідомлення
	newMessage := []byte(firstPart + resultStr + newMessageCode + secondPart + terminator)

	return newMessage, nil
}

// changeTestCode замінює код тесту згідно мапи правил
func changeTestCode(code string, rules *config.CIDRules) string {
	if newCode, ok := rules.TestCodeMap[code]; ok {
		slog.Debug("Test code replaced", "old", code, "new", newCode)
		return newCode
	}
	return code
}

// IsHeartBeat перевіряє чи є повідомлення heartbeat
func IsHeartBeat(message string) bool {
	if len(message) != heartbeatLength {
		return false
	}

	// Перевіряємо body повідомлення
	body := message[bodyStart:bodyEnd]
	if body != heartbeatBody {
		return false
	}

	// Перевіряємо код (перші 4 символи)
	codeStr := message[:4]
	code, err := strconv.Atoi(codeStr)
	if err != nil {
		return false
	}

	// Код має бути в діапазоні 1000-1999
	return code >= heartbeatCodeMin && code <= heartbeatCodeMax
}