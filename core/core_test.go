package core

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGreet tests the Greet method.
func TestGreet(t *testing.T) {
	// Since NewApp has side effects (logging, file creation), we can't easily use it.
	// We create an App instance manually for this simple test.
	a := &App{}
	name := "Tester"
	expected := "Hello Tester, It's show time!"
	if got := a.Greet(name); got != expected {
		t.Errorf("Greet() = %q, want %q", got, expected)
	}
}

// TestNewAppSanity checks if NewApp returns a struct with non-nil essential fields.
func TestNewAppSanity(t *testing.T) {
	// This is an integration-style test as NewApp performs file I/O.
	// We'll create a temporary directory for logs.
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}
	defer os.Chdir(origDir)

	app := NewApp()

	if app == nil {
		t.Fatal("NewApp() returned nil")
	}
	if app.cfg == nil {
		t.Error("NewApp().cfg is nil")
	}
	if app.appQueue == nil {
		t.Error("NewApp().appQueue is nil")
	}
	if app.tcpServer == nil {
		t.Error("NewApp().tcpServer is nil")
	}
	if app.tcpClient == nil {
		t.Error("NewApp().tcpClient is nil")
	}
	if app.logger == nil {
		t.Error("NewApp().logger is nil")
	}
	if app.ctx == nil {
		t.Error("NewApp().ctx is nil")
	}
	// Check if the log file was created
	logPath := filepath.Join(tempDir, "cid_retranslator.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file was not created at %s", logPath)
	}
}

// TestLogHandlerAndGetLogs tests the custom log handler and the GetLogs buffer.
func TestLogHandlerAndGetLogs(t *testing.T) {
	// Setup a minimal App struct for testing the logger without full initialization.
	app := &App{
		logBuffer: make([]string, 0, 100),
	}
	// Create a handler that writes to a buffer in memory instead of stdout/file.
	var buf strings.Builder
	handler := &logHandler{
		app:     app,
		handler: slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}),
	}
	app.logger = slog.New(handler)
	slog.SetDefault(app.logger) // Temporarily set default logger for the test

	// --- Test 1: Basic logging ---
	app.logger.Info("first message", "key1", "value1")
	logs := app.GetLogs()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log message, got %d", len(logs))
	}
	if !strings.Contains(logs[0], "first message") || !strings.Contains(logs[0], "key1=value1") {
		t.Errorf("Log message content is incorrect. Got: %s", logs[0])
	}

	// --- Test 2: Log buffer capping ---
	// Log 110 times to test the capping at 100

	logs = app.GetLogs()
	if len(logs) != 100 {
		t.Fatalf("Expected log buffer to be capped at 100, but got %d", len(logs))
	}
	// The first message should be "Log entry 10" because we logged 1 + 110 times,
	// and the first 11 are pushed out.
	if !strings.Contains(logs[0], "Log entry 10") {
		t.Errorf("Expected the first log entry to be 'Log entry 10', but got: %s", logs[0])
	}
	// The last message should be "Log entry 109"
	if !strings.Contains(logs[99], "Log entry 109") {
		t.Errorf("Expected the last log entry to be 'Log entry 109', but got: %s", logs[99])
	}
}

// Helper to use Debugf with slog
func (a *App) Debugf(format string, args ...interface{}) {
	a.logger.Debug(fmt.Sprintf(format, args...))
}
