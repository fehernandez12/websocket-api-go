package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	clients    []*Client
	register   chan *Client
	unregister chan *Client
	mutex      *sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make([]*Client, 0),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		mutex:      &sync.Mutex{},
	}
}

func (h *Hub) Handle(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.WithError(err).Error("failed to upgrade connection")
		http.Error(w, "failed to upgrade connection", http.StatusInternalServerError)
	}
	client := NewClient(h, socket)
	h.register <- client
	go client.Write()
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.onConnect(client)
		case client := <-h.unregister:
			h.onDisconnect(client)
		}
	}
}

func (h *Hub) onConnect(c *Client) {
	logrus.Infof("client connected: %s", c.socket.RemoteAddr().String())
	h.mutex.Lock()
	defer h.mutex.Unlock()
	c.id = uuid.New().String()
	c.ip = c.socket.RemoteAddr().String()
	h.clients = append(h.clients, c)
}

func (h *Hub) onDisconnect(client *Client) {
	logrus.Infof("client disconnected: %s", client.socket.RemoteAddr().String())
	h.mutex.Lock()
	defer h.mutex.Unlock()
	i := -1
	for j, c := range h.clients {
		if c.id == client.id {
			i = j
			break
		}
	}
	copy(h.clients[i:], h.clients[i+1:])
	h.clients[len(h.clients)-1] = nil
	h.clients = h.clients[:len(h.clients)-1]
}

func (h *Hub) Broadcast(message any, ignore *Client) {
	data, _ := json.Marshal(message)
	for _, client := range h.clients {
		if client != ignore {
			client.outbound <- data
		}
	}
}
