package websocket

import (
	"context"
	"log"
	"net/http"
	"sync"

	"live-polling-backend/services"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for WebSocket connections
	},
}

type Hub struct {
	mu           sync.RWMutex
	rooms        map[string]map[*Client]bool
	subCancels   map[string]func()
	Register     chan *Client
	Unregister   chan *Client
	redisService *services.RedisService
}

func NewHub(redisService *services.RedisService) *Hub {
	return &Hub{
		rooms:        make(map[string]map[*Client]bool),
		subCancels:   make(map[string]func()),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		redisService: redisService,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			roomClients, exists := h.rooms[client.PollID]
			if !exists {
				roomClients = make(map[*Client]bool)
				h.rooms[client.PollID] = roomClients
				// Start Redis Pub/Sub listener for this poll room
				h.startRedisSubscription(client.PollID)
			}
			roomClients[client] = true
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if roomClients, ok := h.rooms[client.PollID]; ok {
				if _, exists := roomClients[client]; exists {
					delete(roomClients, client)
					close(client.Send)
					if len(roomClients) == 0 {
						delete(h.rooms, client.PollID)
						if cancel, hasCancel := h.subCancels[client.PollID]; hasCancel {
							cancel()
							delete(h.subCancels, client.PollID)
						}
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) startRedisSubscription(pollID string) {
	ctx := context.Background()
	ch, cleanup, err := h.redisService.Subscribe(ctx, pollID)
	if err != nil {
		log.Printf("Failed to subscribe to Redis events for poll %s: %v", pollID, err)
		return
	}

	h.subCancels[pollID] = cleanup

	go func() {
		for msg := range ch {
			h.broadcastToRoom(pollID, []byte(msg))
		}
	}()
}

func (h *Hub) broadcastToRoom(pollID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	roomClients, exists := h.rooms[pollID]
	if !exists {
		return
	}

	for client := range roomClients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(roomClients, client)
		}
	}
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, pollID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		Hub:    h,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		PollID: pollID,
	}

	h.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
