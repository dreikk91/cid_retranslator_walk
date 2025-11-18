package core

import (
	"cid_retranslator_walk/client"
	"cid_retranslator_walk/config"
	"cid_retranslator_walk/queue"
	"cid_retranslator_walk/server"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	// "github.com/wailsapp/wails/v2/pkg/runtime"
	"gopkg.in/natefinch/lumberjack.v2"
)

// App struct
type App struct {
	ctx context.Context // Signal context for shutdown
	// wailsCtx   context.Context // Wails context for runtime calls
	cfg        *config.Config
	appQueue   *queue.Queue
	tcpServer  *server.Server
	tcpClient  *client.Client
	logger     *slog.Logger
	fileLogger *lumberjack.Logger // Store fileLogger for closing
	cancelfunc context.CancelFunc
	wg         sync.WaitGroup
	logBuffer  []string
	logMu      sync.RWMutex
	startTime  time.Time
	statsChan <-chan client.Stats
	statsMu   sync.RWMutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	cfg := config.New()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sharedQueue := queue.New(cfg.Queue.BufferSize)

	app := &App{
		ctx:        ctx,
		cfg:        cfg,
		tcpServer:  server.New(&cfg.Server, sharedQueue, &cfg.CIDRules),
		tcpClient:  client.New(&cfg.Client, sharedQueue),
		cancelfunc: cancel,
		logBuffer:  make([]string, 0, 100),
		startTime:  time.Now(),
	}

	// Validate log file path and create directory if needed
	if cfg.Logging.Filename == "" {
		cfg.Logging.Filename = "cid_retranslator.log" // Default filename
	}

	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine executable path: %v\n", err)
	}

	exeDir := filepath.Dir(exePath)

	if !filepath.IsAbs(cfg.Logging.Filename) {
		cfg.Logging.Filename = filepath.Join(exeDir, cfg.Logging.Filename)
	}

	logDir := filepath.Dir(cfg.Logging.Filename)
	if logDir != "." && logDir != "" {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create log directory %s: %v\n", logDir, err)
		}
	}

	fileLogger := &lumberjack.Logger{
		Filename:   cfg.Logging.Filename,
		MaxSize:    cfg.Logging.MaxSize,
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAge:     cfg.Logging.MaxAge,
		Compress:   cfg.Logging.Compress,
	}
	app.fileLogger = fileLogger // Store for closing later
	var multiWriter io.Writer
	if isWindowsGUI() {
		// Тільки файл, без stdout
		multiWriter = fileLogger
	} else {
		// Звичайний режим: і файл, і консоль
		multiWriter = io.MultiWriter(os.Stdout, fileLogger)
	}
	// multiWriter := io.MultiWriter(os.Stdout, fileLogger)

	// Create custom handler for collecting full messages
	handler := &logHandler{
		app:     app,
		handler: slog.NewTextHandler(multiWriter, &slog.HandlerOptions{Level: slog.LevelDebug}),
	}

	app.logger = slog.New(handler)
	slog.SetDefault(app.logger)

	// Log initialization to verify logging setup
	app.logger.Info("Logger initialized", "filename", cfg.Logging.Filename)

	return app
}

// logHandler and related methods
type logHandler struct {
	app     *App
	handler slog.Handler
}

func (h *logHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *logHandler) Handle(ctx context.Context, r slog.Record) error {
	var msgBuilder strings.Builder
	msgBuilder.WriteString(r.Time.Format("15:04:05.000 2006-01-02"))
	msgBuilder.WriteString("\t")
	msgBuilder.WriteString(r.Level.String())
	msgBuilder.WriteString("\t")
	msgBuilder.WriteString(r.Message)
	r.Attrs(func(a slog.Attr) bool {
		if a.Key != slog.TimeKey && a.Key != slog.LevelKey && a.Key != slog.MessageKey {
			msgBuilder.WriteString("\t")
			msgBuilder.WriteString(a.Key)
			msgBuilder.WriteString("=")
			msgBuilder.WriteString(a.Value.String())
		}
		return true
	})

	h.app.logMu.Lock()
	h.app.logBuffer = append(h.app.logBuffer, msgBuilder.String())
	if len(h.app.logBuffer) > 100 {
		h.app.logBuffer = h.app.logBuffer[1:]
	}
	h.app.logMu.Unlock()

	err := h.handler.Handle(ctx, r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write log: %v\n", err)
	}

	return err
}

func (h *logHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &logHandler{app: h.app, handler: h.handler.WithAttrs(attrs)}
}

func (h *logHandler) WithGroup(name string) slog.Handler {
	return &logHandler{app: h.app, handler: h.handler.WithGroup(name)}
}

func (a *App) Ctx() context.Context {
	return a.ctx
}

func isWindowsGUI() bool {
	// Якщо stdout/stderr відсутні — це GUI-збірка
	return runtime.GOOS == "windows" && !isTerminal(os.Stdout)
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// Startup is called when the app starts
func (a *App) Startup() {

	// Start TCP server and client
	a.wg.Add(2)
	go func() {
		defer a.wg.Done()
		a.tcpServer.Run(a.ctx)
	}()
	go func() {
		defer a.wg.Done()
		a.tcpClient.Run(a.ctx)
	}()

}

// Shutdown is called when the app is closing
func (a *App) Shutdown(ctx context.Context) {
	a.logger.Info("Received shutdown signal, initiating graceful shutdown...")
	a.cancelfunc()
	a.tcpServer.Stop()
	a.tcpClient.Stop()
	a.wg.Wait()
	a.appQueue.Close()
	if a.fileLogger != nil {
		if err := a.fileLogger.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close file logger: %v\n", err)
		}
	}
	systray.Quit() // Ensure system tray is closed
	a.logger.Info("Program exited gracefully")
}

// GetServer повертає TCP сервер (для доступу до його методів)
func (a *App) GetServer() *server.Server {
	return a.tcpServer
}

func (a *App) GetClient() *client.Client {
	return a.tcpClient
}


// GetDeviceUpdates повертає канал оновлень пристроїв
func (a *App) GetDeviceUpdates() <-chan server.Device {
	return a.tcpServer.GetDeviceUpdatesChannel()
}

// GetEventUpdates повертає канал глобальних подій
func (a *App) GetEventUpdates() <-chan server.GlobalEvent {
	return a.tcpServer.GetEventUpdatesChannel()
}

// GetInitialDevices повертаєSnapshot всіх пристроїв (для початкового завантаження)
func (a *App) GetInitialDevices() []server.Device {
	return a.tcpServer.GetDevices()
}

// GetInitialEvents повертає Snapshot всіх подій (для початкового завантаження)
func (a *App) GetInitialEvents() []server.GlobalEvent {
	return a.tcpServer.GetGlobalEvents()
}

// GetDeviceEvents повертає події конкретного пристрою
func (a *App) GetDeviceEvents(deviceID int) []server.Event {
	return a.tcpServer.GetDeviceEvents(deviceID)
}

// GetDeviceEventChannel повертає канал для нових подій пристрою
func (a *App) GetDeviceEventChannel(deviceID int) <-chan server.Event {
	return a.tcpServer.GetDeviceEventChannel(deviceID)
}

// CloseDeviceEventChannel закриває канал подій пристрою
func (a *App) CloseDeviceEventChannel(deviceID int) {
	a.tcpServer.CloseDeviceEventChannel(deviceID)
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
