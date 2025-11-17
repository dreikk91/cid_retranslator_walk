package cidparser

import (
	"cid_retranslator_gio/config"
	"strings"
	"testing"
)

func TestIsMessageValid(t *testing.T) {
	// NOTE: The IsMessageValid function in its current form does not correctly handle heartbeats.
	// A valid heartbeat would be considered invalid by this function.
	rules := &config.CIDRules{
		RequiredPrefix: "5",
		ValidLength:    21, // Assuming rules expect a 21-byte message
	}

	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{"Valid Message (21 bytes)", "5040 182109E60300000\x14", true},
		{"Invalid Prefix", "1040 182109E60300000\x14", false},
		{"Invalid Length (Short)", "5040 182109\x14", false},
		{"Invalid Length (Long)", "5040 182109E60300000\x14EXTRA", false},
		{"Empty Message", "", false},
		{"Heartbeat Message (Fails)", "1500           @   ", false}, // This should ideally be true if handled
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMessageValid(tt.message, rules); got != tt.want {
				t.Errorf("IsMessageValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeTestCode(t *testing.T) {
	rules := &config.CIDRules{
		TestCodeMap: map[string]string{"E603": "E602"},
	}

	tests := []struct {
		name string
		code string
		want string
	}{
		{"Known Code", "E603", "E602"},
		{"Unknown Code", "E111", "E111"},
		{"Empty Code", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := changeTestCode(tt.code, rules); got != tt.want {
				t.Errorf("changeTestCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeAccountNumber(t *testing.T) {
	tests := []struct {
		name          string
		message       []byte
		rules         *config.CIDRules
		expected      []byte
		expectError   bool
		expectedError string
	}{
		{
			name:        "valid message, account number in range 2000-2200",
			message:     []byte("5010 182100R41601018"), // 20 bytes, no delimiter
			rules:       &config.CIDRules{AccNumAdd: 1000},
			expected:    []byte("5010 183100R41601018\x14"), // Expect delimiter to be added
			expectError: false,
		},
		{
			name:        "valid message, account number below range",
			message:     []byte("5010 181002R41601018"),
			rules:       &config.CIDRules{AccNumAdd: 1000},
			expected:    []byte("5010 181002R41601018\x14"),
			expectError: false,
		},
		{
			name:        "valid message with test code substitution",
			message:     []byte("5010 186569E60301018"),
			rules:       &config.CIDRules{AccNumAdd: 1000, TestCodeMap: map[string]string{"E603": "E602"}},
			expected:    []byte("5010 186569E60201018\x14"),
			expectError: false,
		},
		{
			name:          "non-numeric account number",
			message:       []byte("5010 18XXXXR41601018"),
			rules:         &config.CIDRules{},
			expectError:   true,
			expectedError: "error converting account number 'XXXX'",
		},
		{
			name:          "invalid message length (short)",
			message:       []byte("short"),
			rules:         &config.CIDRules{},
			expectError:   true,
			expectedError: "invalid message length: got 5, want 20",
		},
		{
			name:          "invalid message length (long)",
			message:       []byte("a very long message that is not 20 bytes"),
			rules:         &config.CIDRules{},
			expectError:   true,
			expectedError: "invalid message length: got 39, want 20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The function now expects message without delimiter
			result, err := ChangeAccountNumber(tt.message, tt.rules)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected an error but got none")
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing '%s' but got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error but got: %v", err)
				}
				if string(result) != string(tt.expected) {
					t.Errorf("expected '%s' but got '%s'", string(tt.expected), string(result))
				}
			}
		})
	}
}

// Updated test for the new IsHeartBeat implementation
func TestIsHeartBeat(t *testing.T) {
	const validBody = "           @    " // 16 chars
	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{"Valid Heartbeat (lower bound)", "1000" + validBody, true},
		{"Valid Heartbeat (upper bound)", "1999" + validBody, true},
		{"Valid Heartbeat (middle)", "1500" + validBody, true},
		{"Invalid Heartbeat (below lower bound)", "0999" + validBody, false},
		{"Invalid Heartbeat (above upper bound)", "2000" + validBody, false},
		{"Invalid Body", "1010" + "invalid body   ", false},
		{"Invalid Code (not a number)", "XXXX" + validBody, false},
		{"Invalid Length (short)", "1000" + "           @  ", false},
		{"Invalid Length (long)", "1000" + "            @   ", false},
		{"Empty Message", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsHeartBeat(tt.message); got != tt.want {
				t.Errorf("IsHeartBeat() for message '%s' = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}