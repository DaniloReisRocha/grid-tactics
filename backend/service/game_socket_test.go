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

func readState(t *testing.T, conn *websocket.Conn, status models.GameStatus) models.GameStatePayload {
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
	readState(t, one, models.GameStatusWaiting)
	join(two, b)
	state := readState(t, one, models.GameStatusPlaying)
	readState(t, two, models.GameStatusPlaying)
	if state.CurrentTurn == nil || state.CurrentUnitID == "" || state.RoundCounter != 1 || len(state.RoundOrder) != 6 {
		t.Fatal("missing initial round")
	}
	if state.TurnDeadlineMS <= state.ServerTimeMS || state.TurnDeadlineMS-state.ServerTimeMS > 45000 || state.TurnRemainingMS <= 0 {
		t.Fatal("initial turn deadline was not sent over websocket")
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
	state = readState(t, one, models.GameStatusPlaying)
	other := readState(t, two, models.GameStatusPlaying)
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
	if state.TurnDeadlineMS != other.TurnDeadlineMS {
		t.Fatal("clients received different turn deadlines")
	}
	sendAction(t, one, endTurn, struct{}{})
	state = readState(t, one, models.GameStatusPlaying)
	readState(t, two, models.GameStatusPlaying)
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
	readState(t, one, models.GameStatusPaused)
	two = dial()
	defer two.Close()
	join(two, b)
	state = readState(t, one, models.GameStatusPlaying)
	readState(t, two, models.GameStatusPlaying)
	if state.TurnCounter != 2 || findUnit(state).Position != target {
		t.Fatal("reconnect reset match")
	}
	if !reflect.DeepEqual(roundOrder, state.RoundOrder) || state.RoundTurn != roundTurn || state.RoundCounter != roundCounter || state.CurrentUnitID != activeUnit {
		t.Fatal("reconnect reshuffled or advanced the round")
	}
	replacement := dial()
	defer replacement.Close()
	join(replacement, b)
	readState(t, one, models.GameStatusPlaying)
	readState(t, replacement, models.GameStatusPlaying)
	_ = two.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, _, err := two.ReadMessage()
	if !websocket.IsCloseError(err, 4001) {
		t.Fatalf("expected replacement close code, got %v", err)
	}
	sendAction(t, one, leaveGame, struct{}{})
	state = readState(t, replacement, models.GameStatusFinished)
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

func TestOrbSpawnCollectionAndReconnectAreSynchronized(t *testing.T) {
	manager := game.NewSessionManager()
	session := manager.CreateSession("orb-socket")
	roster := &GameUnitsGetService{}
	a, b := uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4())
	// Prepare the fixture before opening sockets; no goroutine owns it yet.
	for _, owner := range []uuid.UUID{a, b} {
		units, err := roster.Run(owner)
		if err != nil {
			t.Fatal(err)
		}
		if err := session.AddPlayer(nil, owner, units); err != nil {
			t.Fatal(err)
		}
	}
	for i := range session.Board {
		session.Board[i].Height = 0
		session.Board[i].Terrain = models.TerrainGrass
		session.Board[i].Object = models.ObjectNone
	}
	for _, owner := range []uuid.UUID{a, b} {
		for i := range session.Players[owner].Units {
			if session.Players[owner].Units[i].Unit.ID != session.CurrentUnitID {
				session.Players[owner].Units[i].CurrentLife = 0
				break
			}
		}
	}
	if err := session.EndTurn(*session.CurrentTurn); err != nil {
		t.Fatal(err)
	}
	if len(session.Orbs) != 1 {
		t.Fatal("missing milestone fixture")
	}
	orb := session.Orbs[0]
	owner := *session.CurrentTurn
	var actor *models.UnitInstance
	occupied := map[models.Position]bool{}
	for _, player := range session.Players {
		for i := range player.Units {
			unit := &player.Units[i]
			occupied[unit.Position] = true
			if unit.Unit.ID == session.CurrentUnitID {
				actor = unit
			}
		}
	}
	found := false
	for _, delta := range []models.Position{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} {
		position := models.Position{X: orb.X + delta.X, Y: orb.Y + delta.Y}
		if position.X >= 0 && position.X < 15 && position.Y >= 0 && position.Y < 15 && !occupied[position] {
			actor.Position = position
			found = true
			break
		}
	}
	if !found {
		t.Fatal("no adjacent fixture slot")
	}
	shield, _ := models.AbilityByID(models.AbilityShieldID)
	grenade, _ := models.AbilityByID(models.AbilityGrenadeID)
	actor.Unit.Abilities = []models.Ability{shield, grenade}
	actor.CurrentShield = 0
	actor.AbilityUses = map[string]int{grenade.ID: 1}
	actorID := actor.Unit.ID
	svc := GameWsService{Sessions: manager, GetUnits: roster.Run}
	server := httptest.NewServer(http.HandlerFunc(svc.Receive))
	defer server.Close()
	dial := func() *websocket.Conn {
		t.Helper()
		conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
		if err != nil {
			t.Fatal(err)
		}
		return conn
	}
	join := func(conn *websocket.Conn, id uuid.UUID) {
		sendAction(t, conn, joinGame, joinGameAction{SessionID: "orb-socket", AccoID: id.String()})
	}
	clients := map[uuid.UUID]*websocket.Conn{a: dial(), b: dial()}
	defer clients[a].Close()
	defer clients[b].Close()
	join(clients[a], a)
	readState(t, clients[a], models.GameStatusPlaying)
	join(clients[b], b)
	one, two := readState(t, clients[a], models.GameStatusPlaying), readState(t, clients[b], models.GameStatusPlaying)
	if len(one.Orbs) != 1 || !reflect.DeepEqual(one.Orbs, two.Orbs) || one.Orbs[0].Triggered {
		t.Fatal("spawn not synchronized")
	}
	sendAction(t, clients[owner], movePiece, movePieceAction{UnitID: actorID, To: orb.Position})
	one, two = readState(t, clients[a], models.GameStatusPlaying), readState(t, clients[b], models.GameStatusPlaying)
	if !reflect.DeepEqual(one.Orbs, two.Orbs) || !one.Orbs[0].Triggered || one.Orbs[0].CollectedBy != actorID {
		t.Fatal("collection not synchronized")
	}
	for _, state := range []models.GameStatePayload{one, two} {
		for _, unit := range state.Players[owner].Units {
			if unit.Unit.ID == actorID && (unit.CurrentShield != 3 || unit.AbilityUses[grenade.ID] != 0) {
				t.Fatal("ability restoration not synchronized")
			}
		}
	}
	sendAction(t, clients[owner], useAbility, useAbilityAction{UnitID: actorID, AbilityID: grenade.ID, To: orb.Position})
	one, two = readState(t, clients[a], models.GameStatusPlaying), readState(t, clients[b], models.GameStatusPlaying)
	if one.GrenadeEvent == nil || !reflect.DeepEqual(one.GrenadeEvent, two.GrenadeEvent) || one.GrenadeEvent.Position != orb.Position {
		t.Fatal("grenade animation event not synchronized")
	}
	other := a
	if owner == a {
		other = b
	}
	clients[owner].Close()
	readState(t, clients[other], models.GameStatusPaused)
	reconnected := dial()
	defer reconnected.Close()
	join(reconnected, owner)
	state := readState(t, reconnected, models.GameStatusPlaying)
	readState(t, clients[other], models.GameStatusPlaying)
	if !reflect.DeepEqual(state.Orbs, one.Orbs) {
		t.Fatal("reconnect reset orb state")
	}
	if !reflect.DeepEqual(state.GrenadeEvent, one.GrenadeEvent) {
		t.Fatal("reconnect changed detonation event identity")
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
	readState(t, one, models.GameStatusWaiting)
	sendAction(t, two, joinGame, joinGameAction{SessionID: "selected-team", AccoID: b.String()})
	state := readState(t, one, models.GameStatusPlaying)
	readState(t, two, models.GameStatusPlaying)
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
		state = readState(t, one, models.GameStatusPlaying)
		readState(t, two, models.GameStatusPlaying)
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
	readState(t, two, models.GameStatusPaused)
	reconnected := dial()
	defer reconnected.Close()
	// An attempted equipment change during rejoin must preserve the confirmed roster.
	selection[0].WeaponID = models.WeaponKatanaID
	sendAction(t, reconnected, joinGame, joinGameAction{SessionID: "selected-team", AccoID: a.String(), Team: selection})
	state = readState(t, reconnected, models.GameStatusPlaying)
	readState(t, two, models.GameStatusPlaying)
	unit = state.Players[a].Units[0]
	if unit.Unit.Weapons[0].ID != models.WeaponPistolID || unit.CurrentShield != 3 || unit.AbilityUses[models.AbilityShieldID] != 0 {
		t.Fatal("reconnection changed loadout or restored consumed skill")
	}
}
