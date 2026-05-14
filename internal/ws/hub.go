package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

// maintains the set of active websocket connections
type Hub struct {
	Rooms     map[int]map[*websocket.Conn]bool
	Broadcast chan VoteUpdate
	Mu        sync.Mutex
}

// payload containing a new poll results to be broadcasted
type VoteUpdate struct {
	PollID int
	Data   interface{}
}

// initializes and returns a new hub instance for websocket management
func NewHub() *Hub {
	return &Hub{
		Rooms:     make(map[int]map[*websocket.Conn]bool),
		Broadcast: make(chan VoteUpdate),
	}
}

// starts an infinite loop that listens for updates (delivers information to the user about votes that have been cast)
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
