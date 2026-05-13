package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	Rooms     map[int]map[*websocket.Conn]bool
	Broadcast chan VoteUpdate
	Mu        sync.Mutex
}

type VoteUpdate struct {
	PollID int
	Data   interface{}
}

func NewHub() *Hub {
	return &Hub{
		Rooms:     make(map[int]map[*websocket.Conn]bool),
		Broadcast: make(chan VoteUpdate),
	}
}

func (h *Hub) Run() {
	for {
		update := <-h.Broadcast
		h.Mu.Lock()
		if clients, ok := h.Rooms[update.PollID]; ok {
			for client := range clients {
				err := client.WriteJSON(update.Data)
				if err != nil {
					client.Close()
					delete(h.Rooms[update.PollID], client)
				}
			}
		}
		h.Mu.Unlock()
	}
}
