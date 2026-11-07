package realtime

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

type ActionSent struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type ActionReceived struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type WSConnection struct {
	AccoID   uuid.UUID
	conn     *websocket.Conn
	send     chan []byte
	closeMux sync.Once
	closed   bool
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWSConnection(w http.ResponseWriter, r *http.Request) (*WSConnection, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	ws := &WSConnection{
		conn: conn,
		send: make(chan []byte, 256),
	}

	go ws.writePump()
	return ws, nil
}

func (ws *WSConnection) Send(v ActionSent) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	ws.send <- data
	return nil
}

func (ws *WSConnection) Close() {
	ws.closeMux.Do(func() {
		ws.closed = true
		close(ws.send)
		ws.conn.Close()
	})
}

func (ws *WSConnection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		ws.Close()
	}()

	for {
		select {
		case message, ok := <-ws.send:
			if !ok {
				ws.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := ws.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("failed to write message:", err)
				return
			}
		case <-ticker.C:
			if err := ws.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("failed to ping websocket:", err)
				return
			}
		}
	}
}

func (ws *WSConnection) ReadPump(onMessage func([]byte)) {
	defer ws.Close()

	ws.conn.SetReadLimit(512)
	ws.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.conn.SetPongHandler(func(string) error {
		ws.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			log.Println("filed to read message:", err)
			break
		}
		onMessage(message)
	}
}
