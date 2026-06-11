package utils

import (
	"sync"

	"github.com/gorilla/websocket"
)

type WSHub struct {
	// Клієнти, згруповані по FamilyID
	// map[familyID]map[*websocket.Conn]bool
	families   map[string]map[*websocket.Conn]bool
	mu         sync.Mutex
	register   chan *WSClient
	unregister chan *WSClient
}

type WSClient struct {
	FamilyID string
	Conn     *websocket.Conn
}

func NewWSHub() *WSHub {
	return &WSHub{
		families:   make(map[string]map[*websocket.Conn]bool),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
	}
}

func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.families[client.FamilyID] == nil {
				h.families[client.FamilyID] = make(map[*websocket.Conn]bool)
			}
			h.families[client.FamilyID][client.Conn] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.families[client.FamilyID]; ok {
				if _, ok := clients[client.Conn]; ok {
					delete(clients, client.Conn)
					_ = client.Conn.Close()
					if len(clients) == 0 {
						delete(h.families, client.FamilyID)
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *WSHub) BroadcastToFamily(familyID string, message interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients, ok := h.families[familyID]
	if !ok {
		return
	}

	for conn := range clients {
		err := conn.WriteJSON(message)
		if err != nil {
			_ = conn.Close()
			delete(clients, conn)
		}
	}
}

func (h *WSHub) Register(familyID string, conn *websocket.Conn) {
	h.register <- &WSClient{FamilyID: familyID, Conn: conn}
}

func (h *WSHub) Unregister(familyID string, conn *websocket.Conn) {
	h.unregister <- &WSClient{FamilyID: familyID, Conn: conn}
}
