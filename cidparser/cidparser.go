package cidparser

import (
	"cid_retranslator_walk/config"
	"fmt"
	"log/slog"
	"strconv"
)

// IsMessageValid checks if a message conforms to the configured rules.
func IsMessageValid(message string, rules *config.CIDRules) bool {
	if len(message) == 0 {
		return false
	}
	if len(message) != rules.ValidLength {
		return false
	}
	if string(message[0]) != rules.RequiredPrefix {
		return false
	}
	return true
}

// ChangeAccountNumber modifies the account number in a message according to the rules.
func ChangeAccountNumber(message []byte, rules *config.CIDRules) ([]byte, error) {
	messageString := string(message)

	if len(messageString) != 20 {
		return nil, fmt.Errorf("invalid message length: got %d, want at least 21", len(messageString))
	}

	firstPart := messageString[:7]
	accountNumber := messageString[7:11]
	messageCode := messageString[11:15]
	secondPart := messageString[15:]

	num, err := strconv.Atoi(accountNumber)
	if err != nil {
		return nil, fmt.Errorf("error converting account number '%s': %w", accountNumber, err)
	}

	if num >= 2000 && num <= 2200 {
		num = num + rules.AccNumAdd
	}

	// Use Sprintf with %04d to ensure the account number is zero-padded to 4 digits.
	resultStr := fmt.Sprintf("%04d", num)
	newMessageCode := changeTestCode(messageCode, rules)
	newMessage := []byte(firstPart + resultStr + newMessageCode + secondPart + "\x14")

	slog.Debug("Changed account number", "original", accountNumber, "new", resultStr)
	return newMessage, nil
}

func changeTestCode(code string, rules *config.CIDRules) string {
	if newCode, ok := rules.TestCodeMap[code]; ok {
		return newCode
	}
	return code
}

func IsHeartBeat(message string) bool {

	if len(message) != 20 {
		return false
	}

	body := message[4:19]

	if body == "           @   " {
		code, err := strconv.Atoi(message[:4])
		if err != nil {
			return false
		}
		return code >= 1000 && code <= 1999
	}

	return false
}
