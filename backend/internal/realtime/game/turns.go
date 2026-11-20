package game

import (
	"errors"
	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
	"math/rand"
)

func (s *GameSession) EndTurn(id uuid.UUID, expectedTurn ...int) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(expectedTurn) > 0 && expectedTurn[0] != s.TurnCounter {
		return errors.New("este turno já foi encerrado")
	}
	if _, err := s.actingPlayer(id); err != nil {
		return err
	}
	action := models.HistoryAction{Type: models.HistoryEndTurn, PlayerID: &id, UnitID: s.CurrentUnitID}
	s.advanceUnitTurn()
	s.spawnMilestoneOrbs()
	s.changed(action)
	return nil
}

// Preserve randomized unit priority, but always alternate players, including
// across rounds. A smaller team repeats units until every living unit has acted.
func (s *GameSession) startRound() {
	if len(s.RoundOrder) == 0 {
		for _, id := range s.TurnOrder {
			queue := []models.UnitTurn{}
			for _, unit := range s.Players[id].Units {
				queue = append(queue, models.UnitTurn{PlayerID: id, UnitID: unit.Unit.ID})
			}
			rand.Shuffle(len(queue), func(a, b int) { queue[a], queue[b] = queue[b], queue[a] })
			s.RoundOrder = append(s.RoundOrder, queue...)
		}
	}
	previous := s.CurrentTurn
	if s.RoundCounter == 0 {
		previous = nil
	}
	s.RoundOrder = s.scheduleTurns(nil, previous)
	s.RoundCounter++
	s.RoundTurn = 0
}

// Caller holds the session lock. Executed slots are immutable; only future
// slots are rebuilt after eliminations, with priority for units not yet played.
func (s *GameSession) scheduleTurns(played []models.UnitTurn, previous *uuid.UUID) []models.UnitTurn {
	queues := make(map[uuid.UUID][]models.UnitTurn)
	seen := make(map[models.UnitTurn]bool)
	for _, turn := range s.RoundOrder {
		if !seen[turn] && s.turnUnit(turn) != nil {
			queues[turn.PlayerID] = append(queues[turn.PlayerID], turn)
			seen[turn] = true
		}
	}
	for _, id := range s.TurnOrder {
		for _, unit := range s.Players[id].Units {
			turn := models.UnitTurn{PlayerID: id, UnitID: unit.Unit.ID}
			if !seen[turn] && unit.CurrentLife > 0 {
				queues[id] = append(queues[id], turn)
				seen[turn] = true
			}
		}
	}
	owners := []uuid.UUID{}
	for _, id := range s.TurnOrder {
		if len(queues[id]) > 0 {
			owners = append(owners, id)
		}
	}
	if len(owners) < 2 {
		return nil
	}
	done := make(map[models.UnitTurn]bool)
	cursor := make(map[uuid.UUID]int)
	for _, turn := range played {
		done[turn] = true
		for i, candidate := range queues[turn.PlayerID] {
			if candidate == turn {
				cursor[turn.PlayerID] = (i + 1) % len(queues[turn.PlayerID])
			}
		}
	}
	remaining := 0
	for turn := range seen {
		if !done[turn] {
			remaining++
		}
	}
	next := 0
	if previous == nil {
		next = rand.Intn(len(owners))
	} else {
		for i, owner := range owners {
			if owner == *previous {
				next = (i + 1) % len(owners)
				break
			}
		}
	}
	result := []models.UnitTurn{}
	for remaining > 0 {
		owner := owners[next]
		queue := queues[owner]
		index := cursor[owner]
		for offset := 0; offset < len(queue); offset++ {
			candidate := (cursor[owner] + offset) % len(queue)
			if !done[queue[candidate]] {
				index = candidate
				break
			}
		}
		turn := queue[index]
		result = append(result, turn)
		if !done[turn] {
			done[turn] = true
			remaining--
		}
		cursor[owner] = (index + 1) % len(queue)
		next = (next + 1) % len(owners)
	}
	return result
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
	s.startTurnClock(TurnDuration)
}

func (s *GameSession) recalculateTurnOrder() {
	played := append([]models.UnitTurn{}, s.RoundOrder[:s.RoundTurn+1]...)
	remaining := s.scheduleTurns(played, s.CurrentTurn)
	s.RoundOrder = append(played, remaining...)
}

func (s *GameSession) advanceUnitTurn() {
	s.recalculateTurnOrder()
	s.RoundTurn++
	if s.RoundTurn >= len(s.RoundOrder) {
		s.startRound()
	}
	if s.RoundTurn >= len(s.RoundOrder) {
		s.CurrentTurn = nil
		s.CurrentUnitID = ""
		s.clearTurnClock()
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
		if s.Status == models.GameStatusWaiting {
			delete(s.Players, id)
			for i, turn := range s.TurnOrder {
				if turn == id {
					s.TurnOrder = append(s.TurnOrder[:i], s.TurnOrder[i+1:]...)
					break
				}
			}
		} else if s.Status != models.GameStatusFinished {
			s.pauseTurnClock()
			s.Status = models.GameStatusPaused
			if leave {
				for other := range s.Players {
					if other != id {
						winner := other
						s.Winner = &winner
						break
					}
				}
				s.Status = models.GameStatusFinished
				s.CurrentTurn = nil
				s.CurrentUnitID = ""
			}
		}
		actionType := models.HistoryDisconnect
		if leave {
			actionType = models.HistoryLeave
		}
		s.changed(models.HistoryAction{Type: actionType, PlayerID: &id})
		break
	}
	s.lock.Unlock()
	_ = s.BroadcastState()
}
