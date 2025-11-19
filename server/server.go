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
	"strconv"
	"sync"
	"time"
)

const (
	// Розміри буферів
	maxDeviceEvents  = 100
	maxGlobalEvents  = 500
	deviceChanBuffer = 100
	eventChanBuffer  = 100
	detailChanBuffer = 200

	// Таймаути
	readTimeout      = 60 * time.Second
	writeTimeout     = 10 * time.Second
	replyTimeout     = 10 * time.Second
	inactiveTimeout  = time.Hour
	cleanupInterval  = 5 * time.Minute

	// Протокол
	terminatorByte = 0x14
	ackByte        = 0x06
	nackByte       = 0x15
	maxBufferSize  = 8192
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

	// Захищені даними
	deviceMu         sync.RWMutex
	devices          map[int]*Device
	lastActive       map[int]time.Time
	deviceEventChans map[int]chan Event

	// Глобальні події
	globalMu         sync.RWMutex
	globalEventsRing *ring.Ring

	// Постійні канали для UI
	deviceUpdates chan Device
	eventUpdates  chan GlobalEvent
	closeOnce     sync.Once
	wg            sync.WaitGroup
}

type Event struct {
	Time time.Time `json:"time"`
	Data string    `json:"data"`
}

type Device struct {
	ID            int       `json:"id"`
	LastEventTime time.Time `json:"lastEventTime"`
	LastEvent     string    `json:"lastEvent"`
	Events        []Event   `json:"events"`
}

type GlobalEvent struct {
	Time     time.Time `json:"time"`
	DeviceID int       `json:"deviceID"`
	Data     string    `json:"data"`
}

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
		globalEventsRing: ring.New(maxGlobalEvents),
		lastActive:       make(map[int]time.Time),
		deviceUpdates:    make(chan Device, deviceChanBuffer),
		eventUpdates:     make(chan GlobalEvent, eventChanBuffer),
		deviceEventChans: make(map[int]chan Event),
	}
}

func (s *Server) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.queue.UpdateStartTime()

	listener, err := net.Listen("tcp", s.host+":"+s.port)
	if err != nil {
		slog.Error("Failed to start server", "error", err)
		return
	}
	s.listener = listener
	s.isRunning = true

	slog.Info("Server started", "host", s.host, "port", s.port)

	// Горутина прийому з'єднань
	go s.acceptConnections(ctx)

	// Горутина очищення неактивних пристроїв
	go s.cleanupLoop(ctx)

	<-ctx.Done()
	slog.Info("Server stopping...")
	s.isRunning = false
	s.closeChannels()
}

func (s *Server) acceptConnections(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic in acceptConnections", "panic", r)
		}
		s.listener.Close()
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				slog.Info("Server listener stopped")
				return
			default:
				slog.Error("Accept error", "error", err)
				continue
			}
		}

		slog.Info("Accepted connection", "from", conn.RemoteAddr())
		s.wg.Add(1)
		
		connHandler := &connection{
			conn:   conn,
			queue:  s.queue,
			rules:  s.rules,
			server: s,
		}
		go connHandler.handleRequest(ctx)
	}
}

func (s *Server) cleanupLoop(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic in cleanupLoop", "panic", r)
		}
	}()

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanupInactiveDevices()
		}
	}
}

func (s *Server) Stop() {
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			slog.Info("Stopping server...")
			s.cancel()
			if s.listener != nil {
				s.listener.Close()
			}

			done := make(chan struct{})
			go func() {
				s.wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				slog.Info("All connections closed gracefully")
			case <-time.After(5 * time.Second):
				slog.Warn("Server stop timed out")
			}
		}
	})
}

func (s *Server) closeChannels() {
	s.closeOnce.Do(func() {
		close(s.deviceUpdates)
		close(s.eventUpdates)
		slog.Info("Server channels closed")
	})
}

func (s *Server) cleanupInactiveDevices() {
	s.deviceMu.Lock()
	defer s.deviceMu.Unlock()

	now := time.Now()
	var toDelete []int

	for id, last := range s.lastActive {
		if now.Sub(last) > inactiveTimeout {
			toDelete = append(toDelete, id)
		}
	}

	for _, id := range toDelete {
		delete(s.devices, id)
		delete(s.lastActive, id)
		
		if ch, exists := s.deviceEventChans[id]; exists {
			close(ch)
			delete(s.deviceEventChans, id)
		}
	}

	if len(toDelete) > 0 {
		slog.Info("Cleaned up inactive devices", "count", len(toDelete))
	}
}

// UpdateDevice - ВИПРАВЛЕНО: без race conditions
func (s *Server) UpdateDevice(id int, eventData string) {
	now := time.Now()
	event := Event{Time: now, Data: eventData}

	// 1. Оновлюємо device під write lock
	s.deviceMu.Lock()
	dev, exists := s.devices[id]
	if !exists {
		dev = &Device{
			ID:            id,
			LastEventTime: now,
			LastEvent:     eventData,
			Events:        make([]Event, 0, maxDeviceEvents),
		}
		s.devices[id] = dev
	}
	
	dev.LastEventTime = now
	dev.LastEvent = eventData
	dev.Events = append(dev.Events, event)
	
	if len(dev.Events) > maxDeviceEvents {
		dev.Events = dev.Events[len(dev.Events)-maxDeviceEvents:]
	}
	
	s.lastActive[id] = now

	// Копіюємо device для UI
	deviceCopy := Device{
		ID:            dev.ID,
		LastEventTime: dev.LastEventTime,
		LastEvent:     dev.LastEvent,
	}

	// Копіюємо канал під тим самим локом
	var deviceEventCh chan Event
	if ch, ok := s.deviceEventChans[id]; ok {
		deviceEventCh = ch
	}
	s.deviceMu.Unlock()

	// 2. Оновлюємо global events
	s.globalMu.Lock()
	s.globalEventsRing = s.globalEventsRing.Next()
	globalEvent := GlobalEvent{Time: now, DeviceID: id, Data: eventData}
	s.globalEventsRing.Value = globalEvent
	s.globalMu.Unlock()

	// 3. Відправляємо в UI канали (non-blocking)
	select {
	case s.deviceUpdates <- deviceCopy:
		slog.Debug("Device update sent", "deviceID", id)
	default:
		slog.Warn("Device channel full, dropping update", "deviceID", id)
	}

	select {
	case s.eventUpdates <- globalEvent:
		slog.Debug("Event update sent", "deviceID", id)
	default:
		slog.Warn("Event channel full, dropping update", "deviceID", id)
	}

	// 4. Відправляємо в device-specific канал
	if deviceEventCh != nil {
		select {
		case deviceEventCh <- event:
			slog.Debug("Device event sent", "deviceID", id)
		default:
			slog.Debug("Device event channel full", "deviceID", id)
		}
	}
}

func (s *Server) GetDeviceUpdatesChannel() <-chan Device {
	return s.deviceUpdates
}

func (s *Server) GetEventUpdatesChannel() <-chan GlobalEvent {
	return s.eventUpdates
}

func (s *Server) GetDevices() []Device {
	s.deviceMu.RLock()
	defer s.deviceMu.RUnlock()

	devs := make([]Device, 0, len(s.devices))
	for _, d := range s.devices {
		devs = append(devs, Device{
			ID:            d.ID,
			LastEventTime: d.LastEventTime,
			LastEvent:     d.LastEvent,
		})
	}

	slices.SortFunc(devs, func(a, b Device) int {
		return a.ID - b.ID
	})
	
	return devs
}

func (s *Server) GetGlobalEvents() []GlobalEvent {
	s.globalMu.RLock()
	defer s.globalMu.RUnlock()

	events := make([]GlobalEvent, 0, maxGlobalEvents)
	r := s.globalEventsRing
	
	for i := 0; i < maxGlobalEvents; i++ {
		if val, ok := r.Value.(GlobalEvent); ok && !val.Time.IsZero() {
			events = append(events, val)
		}
		r = r.Next()
		if r == s.globalEventsRing {
			break
		}
	}

	if len(events) > maxGlobalEvents {
		events = events[len(events)-maxGlobalEvents:]
	}

	return events
}

func (s *Server) GetDeviceEvents(id int) []Event {
	s.deviceMu.RLock()
	defer s.deviceMu.RUnlock()
	
	if dev, ok := s.devices[id]; ok {
		eventsCopy := make([]Event, len(dev.Events))
		copy(eventsCopy, dev.Events)
		return eventsCopy
	}
	
	return []Event{}
}

func (s *Server) GetDeviceEventChannel(deviceID int) <-chan Event {
	s.deviceMu.Lock()
	defer s.deviceMu.Unlock()

	if s.deviceEventChans == nil {
		s.deviceEventChans = make(map[int]chan Event)
	}

	if _, exists := s.deviceEventChans[deviceID]; !exists {
		s.deviceEventChans[deviceID] = make(chan Event, detailChanBuffer)
	}

	return s.deviceEventChans[deviceID]
}

func (s *Server) CloseDeviceEventChannel(deviceID int) {
	s.deviceMu.Lock()
	defer s.deviceMu.Unlock()

	if ch, exists := s.deviceEventChans[deviceID]; exists {
		close(ch)
		delete(s.deviceEventChans, deviceID)
		slog.Debug("Device event channel closed", "deviceID", deviceID)
	}
}

func (c *connection) handleRequest(ctx context.Context) {
	defer c.server.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic in handleRequest", "panic", r, "from", c.conn.RemoteAddr())
		}
		c.conn.Close()
	}()

	remoteAddr := c.conn.RemoteAddr()
	slog.Debug("Handling request", "from", remoteAddr)

	reader := bufio.NewReader(c.conn)
	var buffer []byte

	for {
		select {
		case <-ctx.Done():
			slog.Info("Closing connection due to shutdown", "client", remoteAddr)
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
				slog.Debug("Read timeout", "from", remoteAddr)
				if _, err := c.conn.Write([]byte{nackByte}); err != nil {
					slog.Error("Error sending NACK on timeout", "error", err)
				}
				continue
			}
			if err != io.EOF {
				slog.Error("Read error", "from", remoteAddr, "error", err)
			} else {
				slog.Debug("Connection closed by client", "client", remoteAddr)
			}
			return
		}
		
		chunk = chunk[:n]
		buffer = append(buffer, chunk...)

		// Split по terminator byte
		for {
			idx := bytes.IndexByte(buffer, terminatorByte)
			if idx == -1 {
				break
			}

			msg := buffer[:idx]
			buffer = buffer[idx+1:]

			if len(msg) == 0 {
				slog.Debug("Empty message", "from", remoteAddr)
				if _, err := c.conn.Write([]byte{nackByte}); err != nil {
					slog.Error("Error sending NACK", "error", err)
				}
				continue
			}

			slog.Debug("Received message", "from", remoteAddr, "length", len(msg))

			if cidparser.IsHeartBeat(string(msg)) {
				if _, err := c.conn.Write([]byte{ackByte}); err != nil {
					slog.Error("Error sending ACK for heartbeat", "error", err)
				}
				continue
			}

			if !cidparser.IsMessageValid(string(msg), c.rules) {
				slog.Debug("Invalid message format", "from", remoteAddr)
				if _, err := c.conn.Write([]byte{nackByte}); err != nil {
					slog.Error("Error sending NACK", "error", err)
				}
				continue
			}

			newMessage, err := cidparser.ChangeAccountNumber(msg, c.rules)
			if err != nil {
				slog.Error("Error processing message", "from", remoteAddr, "error", err)
				if _, err := c.conn.Write([]byte{nackByte}); err != nil {
					slog.Error("Error sending NACK", "error", err)
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
					
					response := []byte{nackByte}
					if clientReply.Status {
						response = []byte{ackByte}
					}
					
					if err := c.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
						slog.Error("Failed to set write deadline", "error", err)
						return
					}
					
					if _, err := c.conn.Write(response); err != nil {
						slog.Error("Error sending response", "error", err)
						return
					}
					
					slog.Debug("Message relayed", "from", remoteAddr, "ack", clientReply.Status)

				case <-time.After(replyTimeout):
					slog.Error("Timeout waiting for client reply", "from", remoteAddr)
					if _, err := c.conn.Write([]byte{nackByte}); err != nil {
						slog.Error("Error sending NACK after timeout", "error", err)
					}
				}
			default:
				slog.Warn("Queue buffer full, rejecting message", "from", remoteAddr)
				if _, err := c.conn.Write([]byte{nackByte}); err != nil {
					slog.Error("Error sending NACK", "error", err)
				}
			}
		}

		if len(buffer) > maxBufferSize {
			slog.Warn("Buffer overflow, resetting", "from", remoteAddr, "size", len(buffer))
			buffer = nil
		}
	}
}

func extractDeviceID(message []byte) int {
	const (
		minMessageLength = 11
		accountStart     = 7
		accountEnd       = 11
	)

	if len(message) < minMessageLength {
		slog.Error("Message too short to extract device ID", "length", len(message))
		return 0
	}

	accountNumberStr := string(message[accountStart:accountEnd])
	accountNumber, err := strconv.Atoi(accountNumberStr)
	if err != nil {
		slog.Error("Failed to parse device ID", "error", err, "value", accountNumberStr)
		return 0
	}

	return accountNumber
}