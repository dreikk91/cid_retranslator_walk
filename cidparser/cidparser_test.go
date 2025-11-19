package cidparser

import (
	"cid_retranslator_walk/config"
	"testing"
)

func TestIsMessageValid(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		rules    *config.CIDRules
		expected bool
	}{
		{
			name:    "Valid message",
			message: "5010 188823R57516331",
			rules: &config.CIDRules{
				RequiredPrefix: "5",
				ValidLength:    20,
			},
			expected: true,
		},
		{
			name:    "Empty message",
			message: "",
			rules: &config.CIDRules{
				RequiredPrefix: "5",
				ValidLength:    20,
			},
			expected: false,
		},
		{
			name:    "Wrong length",
			message: "5010 188823R575",
			rules: &config.CIDRules{
				RequiredPrefix: "5",
				ValidLength:    20,
			},
			expected: false,
		},
		{
			name:    "Wrong prefix",
			message: "4010 188823R57516331",
			rules: &config.CIDRules{
				RequiredPrefix: "5",
				ValidLength:    20,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMessageValid(tt.message, tt.rules)
			if result != tt.expected {
				t.Errorf("IsMessageValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestChangeAccountNumber(t *testing.T) {
	tests := []struct {
		name        string
		message     []byte
		rules       *config.CIDRules
		wantErr     bool
		checkResult func([]byte) bool
	}{
		{
			name:    "Valid account number in range",
			message: []byte("5010 182001R57516331"),
			rules: &config.CIDRules{
				AccNumAdd:   2100,
				TestCodeMap: map[string]string{},
			},
			wantErr: false,
			checkResult: func(result []byte) bool {
				// Очікуємо 2001 + 2100 = 4101
				return string(result[7:11]) == "4101"
			},
		},
		{
			name:    "Account number out of range",
			message: []byte("5010 188823R57516331"),
			rules: &config.CIDRules{
				AccNumAdd:   2100,
				TestCodeMap: map[string]string{},
			},
			wantErr: false,
			checkResult: func(result []byte) bool {
				// Не змінюється, бо не в діапазоні 2000-2200
				return string(result[7:11]) == "8823"
			},
		},
		{
			name:    "Message too short",
			message: []byte("5010 188821"),
			rules: &config.CIDRules{
				AccNumAdd:   2100,
				TestCodeMap: map[string]string{},
			},
			wantErr: true,
		},
		{
			name:    "Invalid account number format",
			message: []byte("5010 18A823R57516331"),
			rules: &config.CIDRules{
				AccNumAdd:   2100,
				TestCodeMap: map[string]string{},
			},
			wantErr: true,
		},
		{
			name:    "Test code replacement",
			message: []byte("5010 188823E603516331"),
			rules: &config.CIDRules{
				AccNumAdd: 2100,
				TestCodeMap: map[string]string{
					"E603": "E602",
				},
			},
			wantErr: false,
			checkResult: func(result []byte) bool {
				// Перевіряємо заміну коду
				return string(result[11:15]) == "E602"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ChangeAccountNumber(tt.message, tt.rules)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ChangeAccountNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && tt.checkResult != nil {
				if !tt.checkResult(result) {
					t.Errorf("ChangeAccountNumber() result check failed, got %s", string(result))
				}
			}
		})
	}
}

func TestIsHeartBeat(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "Valid heartbeat",
			message:  "1010           @    ",
			expected: true,
		},
		{
			name:     "Invalid heartbeat - wrong code",
			message:  "2234           @   ",
			expected: false,
		},
		{
			name:     "Invalid heartbeat - wrong body",
			message:  "1234XXXXXXXXXXX@   ",
			expected: false,
		},
		{
			name:     "Too short",
			message:  "1234",
			expected: false,
		},
		{
			name:     "Too long",
			message:  "1234           @    X",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHeartBeat(tt.message)
			if result != tt.expected {
				t.Errorf("IsHeartBeat() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestChangeTestCode(t *testing.T) {
	rules := &config.CIDRules{
		TestCodeMap: map[string]string{
			"E603": "E602",
			"E100": "E101",
		},
	}

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "Code exists in map",
			code:     "E603",
			expected: "E602",
		},
		{
			name:     "Code does not exist in map",
			code:     "E999",
			expected: "E999",
		},
		{
			name:     "Another mapped code",
			code:     "E100",
			expected: "E101",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := changeTestCode(tt.code, rules)
			if result != tt.expected {
				t.Errorf("changeTestCode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Benchmark для ChangeAccountNumber
func BenchmarkChangeAccountNumber(b *testing.B) {
	message := []byte("5010 188823E603516331")
	rules := &config.CIDRules{
		AccNumAdd: 2100,
		TestCodeMap: map[string]string{
			"E603": "E602",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ChangeAccountNumber(message, rules)
	}
}

// Benchmark для IsMessageValid
func BenchmarkIsMessageValid(b *testing.B) {
	message := "5010 188823E603516331"
	rules := &config.CIDRules{
		RequiredPrefix: "5",
		ValidLength:    20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsMessageValid(message, rules)
	}
}

// Benchmark для IsHeartBeat
func BenchmarkIsHeartBeat(b *testing.B) {
	message := "1010           @    "

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsHeartBeat(message)
	}
}