package game

import (
	"errors"
	"fmt"
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
	ID             string
	maximumPlayers int
	Players        map[uuid.UUID]*PlayerState
	Board          []models.Cell
	TurnOrder      []uuid.UUID
	CurrentTurn    *uuid.UUID
	TurnCounter    int
	Status         string
	Winner         *uuid.UUID
	Revision       uint64
	StartedAt      time.Time
	LastAction     time.Time
	lock           *sync.RWMutex
}

func NewGameSession(id string, maximumPlayers int) *GameSession {
	return &GameSession{
		ID: id, maximumPlayers: maximumPlayers, Players: make(map[uuid.UUID]*PlayerState),
		Board: generateBoard(id, 0.3), TurnOrder: []uuid.UUID{}, Status: "waiting",
		StartedAt: time.Now(), LastAction: time.Now(), lock: &sync.RWMutex{},
	}
}

func (s *GameSession) changed() { s.Revision++; s.LastAction = time.Now() }

// Snapshot under the session lock; serialize/send outside it so clients cannot block gameplay.
func (s *GameSession) BroadcastState() error {
	s.lock.RLock()
	players := make(map[uuid.UUID]*models.Player, len(s.Players))
	connections := make([]*realtime.WSConnection, 0, len(s.Players))
	for id, p := range s.Players {
		copyPlayer := p.Player
		copyPlayer.Units = append([]models.UnitInstance{}, p.Units...)
		players[id] = &copyPlayer
		if p.Connection != nil {
			connections = append(connections, p.Connection)
		}
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
	payload := models.GameStatePayload{
		ID: s.ID, Players: players, Board: append([]models.Cell{}, s.Board...),
		TurnOrder: append([]uuid.UUID{}, s.TurnOrder...), CurrentTurn: turn,
		TurnCounter: s.TurnCounter, Status: s.Status, Winner: winner, Revision: s.Revision,
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
		if s.Status == "paused" && s.allConnected() {
			s.Status = "playing"
		}
		s.changed()
		return nil
	}
	if s.Status != "waiting" || len(s.Players) >= s.maximumPlayers {
		return errors.New("partida fechada")
	}
	if len(units) == 0 {
		return errors.New("conta sem unidades")
	}
	occupied := s.occupied()
	roster := make([]models.UnitInstance, 0, len(units))
	r := rand.New(rand.NewSource(hashSid(s.ID + id.String())))
	for _, unit := range units {
		found := false
		for _, index := range r.Perm(len(s.Board)) {
			cell := s.Board[index]
			if occupied[models.Position{X: cell.X, Y: cell.Y}] {
				continue
			}
			position := models.Position{X: cell.X, Y: cell.Y}
			occupied[position] = true
			roster = append(roster, models.UnitInstance{
				Position: position, Player: len(s.TurnOrder) + 1, OwnerID: id, Unit: unit,
				CurrentLife: unit.LP, CurrentShield: unit.Shield, CurrentStamina: unit.Stamina,
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
		first := s.TurnOrder[0]
		s.CurrentTurn = &first
		s.TurnCounter = 1
		s.Status = "playing"
	}
	s.changed()
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
	if s.Status != "playing" {
		return nil, errors.New("partida não está em andamento")
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
