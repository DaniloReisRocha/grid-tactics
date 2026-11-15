package game

import (
	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
	"math/rand"
)

func (s *GameSession) EndTurn(id uuid.UUID) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, err := s.actingPlayer(id); err != nil {
		return err
	}
	s.advanceUnitTurn()
	s.changed()
	return nil
}

// Shuffle and interleave once per match. Later rounds reuse the same order,
// keeping eliminated units in their slots and skipping them when advancing.
func (s *GameSession) startRound() {
	if len(s.RoundOrder) > 0 {
		s.RoundCounter++
		s.RoundTurn = 0
		for s.RoundTurn < len(s.RoundOrder) && s.turnUnit(s.RoundOrder[s.RoundTurn]) == nil {
			s.RoundTurn++
		}
		return
	}
	queues := make([][]models.UnitTurn, len(s.TurnOrder))
	remaining := 0
	for i, id := range s.TurnOrder {
		for _, unit := range s.Players[id].Units {
			if unit.CurrentLife > 0 {
				queues[i] = append(queues[i], models.UnitTurn{PlayerID: id, UnitID: unit.Unit.ID})
				remaining++
			}
		}
		rand.Shuffle(len(queues[i]), func(a, b int) { queues[i][a], queues[i][b] = queues[i][b], queues[i][a] })
	}
	s.RoundOrder = []models.UnitTurn{}
	if remaining == 0 {
		return
	}
	next := rand.Intn(len(queues))
	for remaining > 0 {
		if len(queues[next]) > 0 {
			s.RoundOrder = append(s.RoundOrder, queues[next][0])
			queues[next] = queues[next][1:]
			remaining--
		}
		next = (next + 1) % len(queues)
	}
	s.RoundCounter++
	s.RoundTurn = 0
}

func (s *GameSession) turnUnit(turn models.UnitTurn) *models.UnitInstance {
	if p := s.Players[turn.PlayerID]; p != nil {
		for i := range p.Units {
			if p.Units[i].Unit.ID == turn.UnitID && p.Units[i].CurrentLife > 0 {
				return &p.Units[i]
			}
		}
	}
	return nil
}

func (s *GameSession) beginUnitTurn() {
	turn := s.RoundOrder[s.RoundTurn]
	unit := s.turnUnit(turn)
	s.CurrentTurn = &turn.PlayerID
	s.CurrentUnitID = turn.UnitID
	s.TurnCounter++
	unit.CurrentStamina = unit.Unit.EffectiveStamina()
	unit.HasAttacked = false
	unit.TurnCount++
}

func (s *GameSession) advanceUnitTurn() {
	s.RoundTurn++
	for s.RoundTurn < len(s.RoundOrder) && s.turnUnit(s.RoundOrder[s.RoundTurn]) == nil {
		s.RoundTurn++
	}
	if s.RoundTurn >= len(s.RoundOrder) {
		s.startRound()
	}
	if s.RoundTurn >= len(s.RoundOrder) {
		s.CurrentTurn = nil
		s.CurrentUnitID = ""
		return
	}
	s.beginUnitTurn()
}

// Disconnect pauses a started match, allowing the same account to reconnect.
// An explicit leave forfeits a started match; waiting-room players release their slot.
func (s *GameSession) Disconnect(conn *realtime.WSConnection, leave bool) {
	s.lock.Lock()
	for id, p := range s.Players {
		if p.Connection != conn {
			continue
		}
		p.Connection = nil
		p.Connected = false
		if s.Status == "waiting" {
			delete(s.Players, id)
			for i, turn := range s.TurnOrder {
				if turn == id {
					s.TurnOrder = append(s.TurnOrder[:i], s.TurnOrder[i+1:]...)
					break
				}
			}
		} else if s.Status != "finished" {
			s.Status = "paused"
			if leave {
				for other := range s.Players {
					if other != id {
						winner := other
						s.Winner = &winner
						break
					}
				}
				s.Status = "finished"
				s.CurrentTurn = nil
				s.CurrentUnitID = ""
			}
		}
		s.changed()
		break
	}
	s.lock.Unlock()
	_ = s.BroadcastState()
}
