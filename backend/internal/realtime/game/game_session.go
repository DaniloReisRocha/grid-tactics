package game

import (
	"errors"
	"fmt"
	"maps"
	"math/rand"
	"sync"
	"time"

	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
)

const gameStateUpdate = "gameStateUpdate"

type PlayerState struct {
	models.Player
	Connection *realtime.WSConnection
}

type GameSession struct {
	history          *matchRecorder
	GrenadeEvent     *models.GrenadeEvent
	TurnDeadline     time.Time
	turnRemaining    time.Duration
	turnTimer        *time.Timer
	clockGeneration  uint64
	closed           bool
	Orbs             []models.Orb
	RoundOrder       []models.UnitTurn
	RoundCounter     int
	RoundTurn        int
	CurrentUnitID    string
	ID               string
	maximumPlayers   int
	spawnSlots       []models.Position
	Players          map[uuid.UUID]*PlayerState
	Board            []models.Cell
	TurnOrder        []uuid.UUID
	CurrentTurn      *uuid.UUID
	TurnCounter      int
	Status           models.GameStatus
	Winner           *uuid.UUID
	Revision         uint64
	MatchAbilityUses map[string]int
	StartedAt        time.Time
	LastAction       time.Time
	lock             *sync.RWMutex
}

func NewGameSession(id string, maximumPlayers int) *GameSession {
	return &GameSession{
		ID: id, maximumPlayers: maximumPlayers, Players: make(map[uuid.UUID]*PlayerState),
		Board: generateBoard(id, 15, 15), TurnOrder: []uuid.UUID{}, Status: models.GameStatusWaiting,
		MatchAbilityUses: map[string]int{},
		StartedAt:        time.Now(), LastAction: time.Now(), lock: &sync.RWMutex{},
	}
}

func (s *GameSession) HasPlayer(id uuid.UUID) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, exists := s.Players[id]
	return exists
}

func (s *GameSession) changed(actions ...models.HistoryAction) {
	if s.Status == models.GameStatusPlaying && s.CurrentTurn != nil && s.RoundTurn < len(s.RoundOrder) {
		s.recalculateTurnOrder()
	}
	if s.Status == models.GameStatusFinished || s.Status == models.GameStatusWaiting {
		s.clearTurnClock()
	}
	s.Revision++
	s.LastAction = time.Now()
	if s.history != nil {
		action := models.HistoryAction{Type: models.HistoryStateChange}
		if len(actions) > 0 {
			action = actions[0]
		}
		s.recordHistory(action)
	}
}

// Snapshot under the session lock; serialize/send outside it so clients cannot block gameplay.
func (s *GameSession) BroadcastState() error {
	s.lock.RLock()
	payload := s.snapshot()
	connections := make([]*realtime.WSConnection, 0, len(s.Players))
	for _, p := range s.Players {
		if p.Connection != nil {
			connections = append(connections, p.Connection)
		}
	}
	s.lock.RUnlock()
	var failures []error
	for _, conn := range connections {
		if err := conn.Send(realtime.ActionSent{Type: gameStateUpdate, Payload: payload}); err != nil {
			failures = append(failures, err)
			conn.Close()
		}
	}
	return errors.Join(failures...)
}

// Caller holds the session lock. History serializes this snapshot before unlock.
func (s *GameSession) snapshot() models.GameStatePayload {
	players := make(map[uuid.UUID]*models.Player, len(s.Players))
	for id, p := range s.Players {
		copyPlayer := p.Player
		copyPlayer.Units = append([]models.UnitInstance{}, p.Units...)
		for i := range copyPlayer.Units {
			copyPlayer.Units[i].AbilityUses = maps.Clone(copyPlayer.Units[i].AbilityUses)
		}
		players[id] = &copyPlayer
	}
	var turn, winner *uuid.UUID
	if s.CurrentTurn != nil {
		id := *s.CurrentTurn
		turn = &id
	}
	if s.Winner != nil {
		id := *s.Winner
		winner = &id
	}
	var grenadeEvent *models.GrenadeEvent
	if s.GrenadeEvent != nil {
		copyEvent := *s.GrenadeEvent
		grenadeEvent = &copyEvent
	}
	payload := models.GameStatePayload{
		GrenadeEvent: grenadeEvent,
		ServerTimeMS: time.Now().UnixMilli(), TurnRemainingMS: s.turnRemaining.Milliseconds(),
		Orbs:       append([]models.Orb{}, s.Orbs...),
		RoundOrder: append([]models.UnitTurn{}, s.RoundOrder...), RoundCounter: s.RoundCounter,
		RoundTurn: s.RoundTurn, CurrentUnitID: s.CurrentUnitID,
		ID: s.ID, Players: players, Board: append([]models.Cell{}, s.Board...),
		TurnOrder: append([]uuid.UUID{}, s.TurnOrder...), CurrentTurn: turn,
		TurnCounter: s.TurnCounter, Status: s.Status, Winner: winner, Revision: s.Revision,
		MatchAbilityUses: maps.Clone(s.MatchAbilityUses),
	}
	if !s.TurnDeadline.IsZero() {
		payload.TurnDeadlineMS = s.TurnDeadline.UnixMilli()
		payload.TurnRemainingMS = max(0, time.Until(s.TurnDeadline).Milliseconds())
	}
	return payload
}

func (s *GameSession) AddPlayer(conn *realtime.WSConnection, id uuid.UUID, units []models.Unit) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if p, exists := s.Players[id]; exists {
		previous := p.Connection
		p.Connection = conn
		p.Connected = true
		if previous != nil && previous != conn {
			previous.Replace()
		}
		if s.Status == models.GameStatusPaused && s.allConnected() {
			s.Status = models.GameStatusPlaying
			s.startTurnClock(s.turnRemaining)
		}
		s.changed(models.HistoryAction{Type: models.HistoryReconnect, PlayerID: &id})
		return nil
	}
	if s.Status != models.GameStatusWaiting || len(s.Players) >= s.maximumPlayers {
		return errors.New("partida fechada")
	}
	if len(units) == 0 {
		return errors.New("conta sem unidades")
	}
	if err := s.prepareSpawns(len(units)); err != nil {
		return err
	}
	occupied := s.occupied()
	roster := make([]models.UnitInstance, 0, len(units))
	r := rand.New(rand.NewSource(int64(seedHash(s.ID + id.String()))))
	for _, unit := range units {
		found := false
		for _, index := range r.Perm(len(s.spawnSlots)) {
			position := s.spawnSlots[index]
			if occupied[position] {
				continue
			}
			occupied[position] = true
			roster = append(roster, models.UnitInstance{
				Position: position, Player: len(s.TurnOrder) + 1, OwnerID: id, Unit: unit,
				CurrentLife: unit.LP, CurrentShield: unit.InitialShield(), CurrentStamina: unit.EffectiveStamina(),
			})
			found = true
			break
		}
		if !found {
			return errors.New("tabuleiro sem espaço para unidades")
		}
	}
	s.Players[id] = &PlayerState{Player: models.Player{
		AccoID: id, Name: fmt.Sprintf("Jogador %d", len(s.TurnOrder)+1), Connected: true, Units: roster,
	}, Connection: conn}
	s.TurnOrder = append(s.TurnOrder, id)
	for index, playerID := range s.TurnOrder {
		player := s.Players[playerID]
		player.Name = fmt.Sprintf("Jogador %d", index+1)
		for i := range player.Units {
			player.Units[i].Player = index + 1
		}
	}
	if len(s.Players) == s.maximumPlayers {
		s.Status = models.GameStatusPlaying
		s.startRound()
		s.beginUnitTurn()
	}
	s.changed(models.HistoryAction{Type: models.HistoryJoin, PlayerID: &id})
	return nil
}

func (s *GameSession) allConnected() bool {
	if len(s.Players) < s.maximumPlayers {
		return false
	}
	for _, p := range s.Players {
		if !p.Connected {
			return false
		}
	}
	return true
}

func (s *GameSession) occupied() map[models.Position]bool {
	result := make(map[models.Position]bool)
	for _, p := range s.Players {
		for _, u := range p.Units {
			if u.CurrentLife > 0 {
				result[u.Position] = true
			}
		}
	}
	return result
}

func (s *GameSession) actingPlayer(id uuid.UUID) (*PlayerState, error) {
	if s.closed || s.Status != models.GameStatusPlaying {
		return nil, errors.New("partida não está em andamento")
	}
	if !s.TurnDeadline.IsZero() && !time.Now().Before(s.TurnDeadline) {
		return nil, errors.New("o tempo deste turno acabou")
	}
	if s.CurrentTurn == nil || *s.CurrentTurn != id {
		return nil, errors.New("não é seu turno")
	}
	p, ok := s.Players[id]
	if !ok || !p.Connected {
		return nil, errors.New("jogador indisponível")
	}
	return p, nil
}

func (s *GameSession) actingUnit(id uuid.UUID, unitID string) (*PlayerState, error) {
	p, err := s.actingPlayer(id)
	if err != nil {
		return nil, err
	}
	if unitID != s.CurrentUnitID {
		return nil, errors.New("não é o turno desta unidade")
	}
	return p, nil
}
