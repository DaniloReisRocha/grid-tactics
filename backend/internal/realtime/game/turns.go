package game

import (
	"errors"
	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime"
	"github.com/gofrs/uuid"
)

func (s *GameSession) EndTurn(id uuid.UUID) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, err := s.actingPlayer(id); err != nil {
		return err
	}
	current := 0
	for i, playerID := range s.TurnOrder {
		if playerID == id {
			current = i
			break
		}
	}
	for offset := 1; offset <= len(s.TurnOrder); offset++ {
		next := s.TurnOrder[(current+offset)%len(s.TurnOrder)]
		p := s.Players[next]
		alive := false
		for i := range p.Units {
			if p.Units[i].CurrentLife > 0 {
				p.Units[i].CurrentStamina = p.Units[i].Unit.Stamina
				p.Units[i].HasAttacked = false
				alive = true
			}
		}
		if !alive {
			continue
		}
		s.CurrentTurn = &next
		s.TurnCounter++
		s.changed()
		return nil
	}
	return errors.New("nenhum jogador ativo")
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
			}
		}
		s.changed()
		break
	}
	s.lock.Unlock()
	_ = s.BroadcastState()
}
