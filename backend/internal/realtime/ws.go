package realtime

import (
	"encoding/json"
	"errors"
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
	done     chan struct{}
	closeMux sync.Once
}

func NewWSConnection(w http.ResponseWriter, r *http.Request, allowedOrigins []string) (*WSConnection, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize: 1024, WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true
			}
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	ws := &WSConnection{conn: conn, send: make(chan []byte, 256), done: make(chan struct{})}
	go ws.writePump()
	return ws, nil
}

func (ws *WSConnection) Send(v ActionSent) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	select {
	case <-ws.done:
		return errors.New("conexão encerrada")
	default:
	}
	// The queue is never closed; done coordinates shutdown without send/close races.
	select {
	case <-ws.done:
		return errors.New("conexão encerrada")
	case ws.send <- data:
		return nil
	default:
		return errors.New("fila de envio cheia")
	}
}
func (ws *WSConnection) Close() {
	ws.closeMux.Do(func() { close(ws.done); _ = ws.conn.Close() })
}

// Tell clients not to reconnect automatically when another tab takes this account.
func (ws *WSConnection) Replace() {
	_ = ws.conn.WriteControl(websocket.CloseMessage,
		websocket.FormatCloseMessage(4001, "account connected elsewhere"), time.Now().Add(time.Second))
	ws.Close()
}
func (ws *WSConnection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()
	defer ws.Close()
	for {
		select {
		case <-ws.done:
			return
		case message := <-ws.send:
			_ = ws.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := ws.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = ws.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := ws.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
func (ws *WSConnection) ReadPump(onMessage func([]byte)) {
	defer ws.Close()
	ws.conn.SetReadLimit(64 * 1024)
	_ = ws.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.conn.SetPongHandler(func(string) error { return ws.conn.SetReadDeadline(time.Now().Add(60 * time.Second)) })
	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			return
		}
		onMessage(message)
	}
}
