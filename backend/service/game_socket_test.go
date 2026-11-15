package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
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
	roster := &GameUnitsGetService{}
	service := GameWsService{Sessions: game.NewSessionManager(), GetUnits: roster.Run, AllowedOrigins: []string{"http://allowed.test"}}
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
	if state.CurrentTurn == nil || state.CurrentUnitID == "" || state.RoundCounter != 1 || len(state.RoundOrder) != 6 {
		t.Fatal("missing initial round")
	}
	if *state.CurrentTurn == b {
		a, b = b, a
		one, two = two, one
	}
	var unit models.UnitInstance
	for _, candidate := range state.Players[a].Units {
		if candidate.Unit.ID == state.CurrentUnitID {
			unit = candidate
			break
		}
	}
	used := map[models.Position]bool{}
	for _, p := range state.Players {
		for _, u := range p.Units {
			used[u.Position] = true
		}
	}
	var target models.Position
	found := false
	cells := map[models.Position]models.Cell{}
	for _, cell := range state.Board {
		cells[models.Position{X: cell.X, Y: cell.Y}] = cell
	}
	for _, candidate := range []models.UnitInstance{unit} {
		unit = candidate
		for _, delta := range []models.Position{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} {
			next := models.Position{X: unit.X + delta.X, Y: unit.Y + delta.Y}
			cell, exists := cells[next]
			if exists && game.CanTraverse(cells[unit.Position], cell) && !used[next] {
				target = next
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Fatal("no neighbor for test movement")
	}
	sendAction(t, one, movePiece, movePieceAction{UnitID: unit.Unit.ID, To: target})
	state = readState(t, one, "playing")
	other := readState(t, two, "playing")
	findUnit := func(state models.GameStatePayload) models.UnitInstance {
		for _, candidate := range state.Players[a].Units {
			if candidate.Unit.ID == unit.Unit.ID {
				return candidate
			}
		}
		t.Fatal("unit missing")
		return models.UnitInstance{}
	}
	if findUnit(state).Position != target || other.Revision != state.Revision || len(findUnit(state).Path) < 2 {
		t.Fatal("movement not synchronized")
	}
	sendAction(t, one, endTurn, struct{}{})
	state = readState(t, one, "playing")
	readState(t, two, "playing")
	if *state.CurrentTurn != b || state.TurnCounter != 2 {
		t.Fatal("turn not synchronized")
	}
	roundOrder := append([]models.UnitTurn{}, state.RoundOrder...)
	roundTurn, roundCounter, activeUnit := state.RoundTurn, state.RoundCounter, state.CurrentUnitID
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
	if state.TurnCounter != 2 || findUnit(state).Position != target {
		t.Fatal("reconnect reset match")
	}
	if !reflect.DeepEqual(roundOrder, state.RoundOrder) || state.RoundTurn != roundTurn || state.RoundCounter != roundCounter || state.CurrentUnitID != activeUnit {
		t.Fatal("reconnect reshuffled or advanced the round")
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

func TestSelectedTeamAndPassivePersistOnWebSocketReconnect(t *testing.T) {
	roster := &GameUnitsGetService{}
	svc := GameWsService{Sessions: game.NewSessionManager(), GetUnits: roster.Run}
	server := httptest.NewServer(http.HandlerFunc(svc.Receive))
	defer server.Close()
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	dial := func() *websocket.Conn {
		c, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	a, b := uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4())
	one, two := dial(), dial()
	defer one.Close()
	defer two.Close()
	selection := testTeam()
	sendAction(t, one, joinGame, joinGameAction{SessionID: "selected-team", AccoID: a.String(), Team: selection})
	readState(t, one, "waiting")
	sendAction(t, two, joinGame, joinGameAction{SessionID: "selected-team", AccoID: b.String()})
	state := readState(t, one, "playing")
	readState(t, two, "playing")
	for i, unit := range state.Players[a].Units {
		if unit.Unit.Weapons[0].ID != selection[i].WeaponID || len(unit.Unit.Abilities) != len(selection[i].AbilityIDs) || unit.CurrentStamina != unit.Unit.EffectiveStamina() {
			t.Fatal("selected equipment did not reach the game")
		}
	}
	unit := state.Players[a].Units[0]
	for tries := 0; state.CurrentUnitID != unit.Unit.ID && tries < 6; tries++ {
		conn := one
		if *state.CurrentTurn == b {
			conn = two
		}
		sendAction(t, conn, endTurn, struct{}{})
		state = readState(t, one, "playing")
		readState(t, two, "playing")
	}
	if state.CurrentUnitID != unit.Unit.ID {
		t.Fatal("equipped unit never received its turn")
	}
	sendAction(t, one, useAbility, useAbilityAction{UnitID: unit.Unit.ID, AbilityID: models.AbilityShieldID, To: unit.Position})
	var failure struct {
		Type string `json:"type"`
	}
	if err := one.ReadJSON(&failure); err != nil || failure.Type != actionFailed {
		t.Fatal("passive accepted as active skill", err)
	}
	if state.Players[a].Units[0].CurrentShield != 3 {
		t.Fatal("equipped passive was not applied at spawn")
	}
	one.Close()
	readState(t, two, "paused")
	reconnected := dial()
	defer reconnected.Close()
	// An attempted equipment change during rejoin must preserve the confirmed roster.
	selection[0].WeaponID = models.WeaponKatanaID
	sendAction(t, reconnected, joinGame, joinGameAction{SessionID: "selected-team", AccoID: a.String(), Team: selection})
	state = readState(t, reconnected, "playing")
	readState(t, two, "playing")
	unit = state.Players[a].Units[0]
	if unit.Unit.Weapons[0].ID != models.WeaponPistolID || unit.CurrentShield != 3 || unit.AbilityUses[models.AbilityShieldID] != 0 {
		t.Fatal("reconnection changed loadout or restored consumed skill")
	}
}
