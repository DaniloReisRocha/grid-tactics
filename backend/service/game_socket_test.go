package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime"
	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime/game"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

func readState(t *testing.T, conn *websocket.Conn, status string) models.GameStatePayload {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		var message struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}
		if err := conn.ReadJSON(&message); err != nil {
			t.Fatal(err)
		}
		if message.Type == actionFailed {
			t.Fatalf("unexpected action error: %s", message.Payload)
		}
		if message.Type != "gameStateUpdate" {
			continue
		}
		var state models.GameStatePayload
		if err := json.Unmarshal(message.Payload, &state); err != nil {
			t.Fatal(err)
		}
		if state.Status == status {
			return state
		}
	}
}

func sendAction(t *testing.T, conn *websocket.Conn, kind string, payload any) {
	t.Helper()
	if err := conn.WriteJSON(realtime.ActionSent{Type: kind, Payload: payload}); err != nil {
		t.Fatal(err)
	}
}

func TestTwoClientsMoveTurnReconnectReplaceAndLeave(t *testing.T) {
	SessionManager = game.NewSessionManager()
	roster := &GameUnitsGetService{}
	service := GameWsService{GetUnits: roster.Run, AllowedOrigins: []string{"http://allowed.test"}}
	server := httptest.NewServer(http.HandlerFunc(service.Receive))
	defer server.Close()
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	dial := func() *websocket.Conn {
		t.Helper()
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			t.Fatal(err)
		}
		return conn
	}
	join := func(conn *websocket.Conn, id uuid.UUID) {
		sendAction(t, conn, joinGame, joinGameAction{SessionID: "integration", AccoID: id.String()})
	}
	a, b := uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4())
	one, two := dial(), dial()
	defer one.Close()
	defer two.Close()
	join(one, a)
	readState(t, one, "waiting")
	join(two, b)
	state := readState(t, one, "playing")
	readState(t, two, "playing")
	if state.CurrentTurn == nil || *state.CurrentTurn != a {
		t.Fatal("wrong initial turn")
	}
	unit := state.Players[a].Units[0]
	used := map[models.Position]bool{}
	for _, p := range state.Players {
		for _, u := range p.Units {
			used[u.Position] = true
		}
	}
	var target models.Position
	found := false
	for _, delta := range []models.Position{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} {
		next := models.Position{X: unit.X + delta.X, Y: unit.Y + delta.Y}
		if next.X >= 0 && next.X < 10 && next.Y >= 0 && next.Y < 10 && !used[next] {
			target = next
			found = true
			break
		}
	}
	if !found {
		t.Fatal("no neighbor for test movement")
	}
	sendAction(t, one, movePiece, movePieceAction{UnitID: unit.Unit.ID, To: target})
	state = readState(t, one, "playing")
	other := readState(t, two, "playing")
	if state.Players[a].Units[0].Position != target || other.Revision != state.Revision {
		t.Fatal("movement not synchronized")
	}
	sendAction(t, one, endTurn, struct{}{})
	state = readState(t, one, "playing")
	readState(t, two, "playing")
	if *state.CurrentTurn != b || state.TurnCounter != 2 {
		t.Fatal("turn not synchronized")
	}
	sendAction(t, one, movePiece, movePieceAction{UnitID: unit.Unit.ID, To: unit.Position})
	_ = one.SetReadDeadline(time.Now().Add(3 * time.Second))
	var failure struct {
		Type    string              `json:"type"`
		Payload actionFailedPayload `json:"payload"`
	}
	if err := one.ReadJSON(&failure); err != nil {
		t.Fatal(err)
	}
	if failure.Type != actionFailed || failure.Payload.Action != movePiece {
		t.Fatal("incorrect error action")
	}
	two.Close()
	readState(t, one, "paused")
	two = dial()
	defer two.Close()
	join(two, b)
	state = readState(t, one, "playing")
	readState(t, two, "playing")
	if state.TurnCounter != 2 || state.Players[a].Units[0].Position != target {
		t.Fatal("reconnect reset match")
	}
	replacement := dial()
	defer replacement.Close()
	join(replacement, b)
	readState(t, one, "playing")
	readState(t, replacement, "playing")
	_ = two.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, _, err := two.ReadMessage()
	if !websocket.IsCloseError(err, 4001) {
		t.Fatalf("expected replacement close code, got %v", err)
	}
	sendAction(t, one, leaveGame, struct{}{})
	state = readState(t, replacement, "finished")
	if state.Winner == nil || *state.Winner != b {
		t.Fatal("forfeit did not end match")
	}
}

func TestSocketRejectsUnlistedOrigin(t *testing.T) {
	roster := &GameUnitsGetService{}
	service := GameWsService{GetUnits: roster.Run, AllowedOrigins: []string{"http://allowed.test"}}
	server := httptest.NewServer(http.HandlerFunc(service.Receive))
	defer server.Close()
	headers := http.Header{"Origin": []string{"http://other.test"}}
	conn, response, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), headers)
	if conn != nil {
		conn.Close()
	}
	if response != nil {
		defer response.Body.Close()
	}
	if err == nil || response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatal("unlisted origin accepted")
	}
}
