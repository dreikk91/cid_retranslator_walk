package server

import (
	"bufio"
	"bytes"
	"cid_retranslator_walk/cidparser"
	"cid_retranslator_walk/config"
	"cid_retranslator_walk/queue"
	"container/ring"
	"context"
	"io"
	"log/slog"
	"net"
	"slices"
	//"sort"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	host             string
	port             string
	queue            *queue.Queue
	rules            *config.CIDRules
	cancel           context.CancelFunc
	stopOnce         sync.Once
	listener         net.Listener
	isRunning        bool
	devices          map[int]*Device
	deviceMu         sync.RWMutex
	globalEventsRing *ring.Ring
	globalMu         sync.RWMutex
	maxGlobalEvents  int
	inactiveTimeout  time.Duration
	lastActive       map[int]time.Time

	// Постійні канали для UI
	deviceUpdates    chan Device
	eventUpdates     chan GlobalEvent
	deviceEventChans map[int]chan Event
	closeOnce        sync.Once
}

// Event represents an event for a device
type Event struct {
	Time time.Time `json:"time"`
	Data string    `json:"data"`
}

// Device represents a device with its events
type Device struct {
	ID            int       `json:"id"`
	LastEventTime time.Time `json:"lastEventTime"`
	LastEvent     string    `json:"lastEvent"`
	Events        []Event   `json:"events"`
}

// GlobalEvent represents a global event across all devices
type GlobalEvent struct {
	Time     time.Time `json:"time"`
	DeviceID int       `json:"deviceID"`
	Data     string    `json:"data"`
}

// connection represents a client connection to the server.
type connection struct {
	conn   net.Conn
	queue  *queue.Queue
	rules  *config.CIDRules
	server *Server
}

func New(cfg *config.ServerConfig, q *queue.Queue, rules *config.CIDRules) *Server {
	return &Server{
		host:             cfg.Host,
		port:             cfg.Port,
		queue:            q,
		rules:            rules,
		devices:          make(map[int]*Device),
		globalEventsRing: ring.New(500),
		maxGlobalEvents:  500,
		inactiveTimeout:  time.Hour,
		lastActive:       make(map[int]time.Time),

		// Ініціалізуємо постійні канали з буфером
		deviceUpdates:    make(chan Device, 100),
		eventUpdates:     make(chan GlobalEvent, 100),
		deviceEventChans: make(map[int]chan Event),
	}
}

func (server *Server) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	server.cancel = cancel
	server.queue.UpdateStartTime()

	listener, err := net.Listen("tcp", server.host+":"+server.port)
	if err != nil {
		slog.Error("Failed to start server", "error", err)
		return
	}
	server.listener = listener
	server.isRunning = true

	slog.Info("Server started", "host", server.host, "port", server.port)

	go func() {
		defer server.listener.Close()
		for {
			conn, err := server.listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					slog.Info("Server listener stopped.")
					return
				default:
					slog.Error("Accept error", "error", err)
				}
				continue
			}
			slog.Info("Accepted connection", "from", conn.RemoteAddr())
			connHandler := &connection{conn: conn, queue: server.queue, rules: server.rules, server: server}
			go connHandler.handleRequest(ctx)
		}
	}()

	// Горутина для очищення неактивних пристроїв
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				server.cleanupInactiveDevices()
			}
		}
	}()

	<-ctx.Done()
	slog.Info("Server stopping...")
	server.isRunning = false

	// Закриваємо канали при зупинці
	server.closeChannels()
}

func (server *Server) Stop() {
	server.stopOnce.Do(func() {
		if server.cancel != nil {
			slog.Info("Stopping server...")
			server.cancel()
			if server.listener != nil {
				server.listener.Close()
			}
		}
	})
}

func (server *Server) closeChannels() {
	server.closeOnce.Do(func() {
		close(server.deviceUpdates)
		close(server.eventUpdates)
		slog.Info("Server channels closed")
	})
}

func (server *Server) cleanupInactiveDevices() {
	//server.deviceMu.Lock()
	//defer server.deviceMu.Unlock()
	//now := time.Now()
	//var toDelete []int
	//for id, last := range server.lastActive {
	//	if now.Sub(last) > server.inactiveTimeout {
	//		toDelete = append(toDelete, id)
	//	}
	//}
	//for _, id := range toDelete {
	//	delete(server.devices, id)
	//	delete(server.lastActive, id)
	//}
	//if len(toDelete) > 0 {
	//	slog.Info("Cleaned up inactive devices", "count", len(toDelete))
	//}
}

// UpdateDevice updates or adds an event for the device
func (server *Server) UpdateDevice(id int, eventData string) {
	now := time.Now()
	event := Event{Time: now, Data: eventData}

	// Оновлюємо device
	server.deviceMu.Lock()
	dev, exists := server.devices[id]
	if !exists {
		dev = &Device{
			ID:            id,
			LastEventTime: now,
			LastEvent:     eventData,
			Events:        make([]Event, 0, 100),
		}
		server.devices[id] = dev
	}
	dev.LastEventTime = now
	dev.LastEvent = eventData
	dev.Events = append(dev.Events, event)
	if len(dev.Events) > 100 {
		dev.Events = dev.Events[len(dev.Events)-100:]
	}
	server.lastActive[id] = now

	// Створюємо копію для відправки в канал
	deviceCopy := Device{
		ID:            dev.ID,
		LastEventTime: dev.LastEventTime,
		LastEvent:     dev.LastEvent,
		Events:        nil,
	}
	server.deviceMu.Unlock()

	// Оновлюємо global events
	server.globalMu.Lock()
	server.globalEventsRing = server.globalEventsRing.Next()
	globalEvent := GlobalEvent{Time: now, DeviceID: id, Data: eventData}
	server.globalEventsRing.Value = globalEvent
	server.globalMu.Unlock()

	// Non-blocking відправка в UI канали
	select {
	case server.deviceUpdates <- deviceCopy:
		slog.Info("✅ Device sent successfully", "deviceID", id)
	default:
		slog.Error("❌ Device channel FULL!", "deviceID", id)
	}

	select {
	case server.eventUpdates <- globalEvent:
		slog.Info("✅ Event sent successfully", "deviceID", id)
	default:
		slog.Error("❌ Event channel FULL!", "deviceID", id)
	}

	// ВИПРАВЛЕНО: Відправка в device-specific канал
	server.deviceMu.RLock() // Використовуємо RLock для читання map
	if ch, ok := server.deviceEventChans[id]; ok {
		select {
		case ch <- event:
		default:
			// канал переповнений — пропускаємо
		}
	}
	server.deviceMu.RUnlock()
}

// GetDeviceUpdatesChannel повертає read-only канал для оновлень пристроїв
func (server *Server) GetDeviceUpdatesChannel() <-chan Device {
	return server.deviceUpdates
}

// GetEventUpdatesChannel повертає read-only канал для глобальних подій
func (server *Server) GetEventUpdatesChannel() <-chan GlobalEvent {
	return server.eventUpdates
}

// GetDevices returns a snapshot of all devices (для початкового завантаження UI)
func (server *Server) GetDevices() []Device {
	server.deviceMu.RLock()
	defer server.deviceMu.RUnlock()

	devs := make([]Device, 0, len(server.devices))
	for _, d := range server.devices {
		devs = append(devs, Device{
			ID:            d.ID,
			LastEventTime: d.LastEventTime,
			LastEvent:     d.LastEvent,
			Events:        nil, // Без історії для швидкості
		})
	}

	// Сортуємо за ID
	slices.SortFunc(devs, func(a, b Device) int { return a.ID - b.ID })
	return devs
}

// GetGlobalEvents returns a snapshot of global events (для початкового завантаження UI)
func (server *Server) GetGlobalEvents() []GlobalEvent {
	server.globalMu.RLock()
	defer server.globalMu.RUnlock()

	events := make([]GlobalEvent, 0, server.maxGlobalEvents)
	r := server.globalEventsRing
	i := 0
	for r != nil && i < server.maxGlobalEvents {
		if val, ok := r.Value.(GlobalEvent); ok && !val.Time.IsZero() {
			events = append(events, val)
		}
		r = r.Next()
		i++
		if r == server.globalEventsRing {
			break
		}
	}

	// Обмежуємо до 500
	if len(events) > 500 {
		events = events[len(events)-500:]
	}

	// Сортуємо за часом (новіші спочатку)
	//sort.Slice(events, func(p, q int) bool {
	//	return events[p].Time.After(events[q].Time)
	//})

	return events
}

// GetDeviceEvents returns the events for a specific device
func (server *Server) GetDeviceEvents(id int) []Event {
	server.deviceMu.RLock()
	defer server.deviceMu.RUnlock()
	if dev, ok := server.devices[id]; ok {
		return append([]Event{}, dev.Events...)
	}
	return []Event{}
}

// GetDeviceEventChannel повертає канал для подій конкретного пристрою
func (server *Server) GetDeviceEventChannel(deviceID int) <-chan Event {
	server.deviceMu.Lock()
	defer server.deviceMu.Unlock()

	// Створюємо канал якщо його ще немає
	if server.deviceEventChans == nil {
		server.deviceEventChans = make(map[int]chan Event)
	}

	if _, exists := server.deviceEventChans[deviceID]; !exists {
		server.deviceEventChans[deviceID] = make(chan Event, 200)
	}

	return server.deviceEventChans[deviceID]
}

// CloseDeviceEventChannel закриває канал подій для пристрою (викликається при закритті діалогу)
func (server *Server) CloseDeviceEventChannel(deviceID int) {
	server.deviceMu.Lock()
	defer server.deviceMu.Unlock()

	if ch, exists := server.deviceEventChans[deviceID]; exists {
		close(ch)
		delete(server.deviceEventChans, deviceID)
	}
}

func (c *connection) handleRequest(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic in handler", "panic", r, "from", c.conn.RemoteAddr())
		}
	}()
	remoteAddr := c.conn.RemoteAddr()
	slog.Debug("Handling request", "from", remoteAddr)
	defer c.conn.Close()

	reader := bufio.NewReader(c.conn)
	var buffer []byte
	readTimeout := 60 * time.Second

	for {
		select {
		case <-ctx.Done():
			slog.Info("Closing connection due to server shutdown.", "client", remoteAddr)
			return
		default:
		}

		if err := c.conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
			slog.Error("Failed to set read deadline", "from", remoteAddr, "error", err)
			return
		}

		chunk := make([]byte, 1024)
		n, err := reader.Read(chunk)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				slog.Warn("Read timeout", "from", remoteAddr)
				if _, err := c.conn.Write([]byte{0x00}); err != nil {
					slog.Error("Error sending NACK on timeout", "error", err)
				}
				continue
			}
			if err != io.EOF {
				slog.Error("Read error", "from", remoteAddr, "error", err)
			} else {
				slog.Info("Connection closed by client", "client", remoteAddr)
			}
			return
		}
		chunk = chunk[:n]
		buffer = append(buffer, chunk...)

		// Split по 0x14
		for {
			idx := bytes.IndexByte(buffer, 0x14)
			if idx == -1 {
				break
			}

			msg := buffer[:idx]
			buffer = buffer[idx+1:]

			if len(msg) == 0 {
				slog.Warn("Empty message", "from", remoteAddr)
				if _, err := c.conn.Write([]byte{0x15}); err != nil {
					slog.Error("Error sending NACK for empty msg", "error", err)
				}
				continue
			}

			slog.Debug("Received message", "from", remoteAddr, "data", string(msg))

			if cidparser.IsHeartBeat(string(msg)) {
				if _, err := c.conn.Write([]byte{0x06}); err != nil {
					slog.Error("Error sending ACK for heartbeat", "error", err)
				}
				continue
			}

			if !cidparser.IsMessageValid(string(msg), c.rules) {
				slog.Warn("Invalid message format", "from", remoteAddr, "data", string(msg))
				if _, err := c.conn.Write([]byte{0x15}); err != nil {
					slog.Error("Error sending NACK for invalid format", "error", err)
				}
				continue
			}

			newMessage, err := cidparser.ChangeAccountNumber(msg, c.rules)
			if err != nil {
				slog.Error("Error processing message", "from", remoteAddr, "error", err)
				if _, err := c.conn.Write([]byte{0x15}); err != nil {
					slog.Error("Error sending NACK for processing error", "error", err)
				}
				continue
			}

			replyCh := make(chan queue.DeliveryData, 1)
			sharedData := queue.SharedData{
				Payload: newMessage,
				ReplyCh: replyCh,
			}

			select {
			case c.queue.DataChannel <- sharedData:
				deviceID := extractDeviceID(newMessage)
				c.server.UpdateDevice(deviceID, string(newMessage))

				select {
				case clientReply, ok := <-replyCh:
					if !ok {
						slog.Warn("Reply channel closed unexpectedly", "from", remoteAddr)
						return
					}
					response, responseType := []byte{0x15}, "NACK"
					if clientReply.Status {
						response, responseType = []byte{0x06}, "ACK"
					}
					if _, err := c.conn.Write(response); err != nil {
						slog.Error("Error sending response", "type", responseType, "error", err)
						return
					}
					slog.Info("Message relayed", "from", remoteAddr, "status", responseType, "data", string(msg))

				case <-time.After(10 * time.Second):
					slog.Error("Timeout waiting for client reply", "from", remoteAddr)
					if _, err := c.conn.Write([]byte{0x15}); err != nil {
						slog.Error("Error sending NACK after timeout", "error", err)
					}
				}
			default:
				slog.Warn("Queue buffer full, rejecting message", "from", remoteAddr)
				if _, err := c.conn.Write([]byte{0x15}); err != nil {
					slog.Error("Error sending NACK for buffer full", "error", err)
				}
			}
		}

		if len(buffer) > 8192 {
			slog.Warn("Large buffer accumulation", "from", remoteAddr, "size", len(buffer))
			buffer = nil
		}
	}
}

func extractDeviceID(message []byte) int {
	if len(message) < 11 {
		slog.Error("Message too short to extract device ID", "length", len(message))
		return 0
	}
	accountNumber, err := strconv.Atoi(string(message[7:11]))
	if err != nil {
		slog.Error("Failed to extract device ID", "error", err)
		return 0
	}
	return accountNumber
}
