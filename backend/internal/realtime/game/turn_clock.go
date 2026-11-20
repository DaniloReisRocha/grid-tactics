package game

import (
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"time"
)

const TurnDuration = 45 * time.Second

// Clock helpers run under the session lock. Generations invalidate callbacks
// already queued when a player ends the turn, disconnects, or ends the match.
func (s *GameSession) stopTurnTimer() {
	s.clockGeneration++
	if s.turnTimer != nil {
		s.turnTimer.Stop()
		s.turnTimer = nil
	}
}

func (s *GameSession) startTurnClock(remaining time.Duration) {
	s.stopTurnTimer()
	s.turnRemaining = max(0, remaining)
	s.TurnDeadline = time.Now().Add(s.turnRemaining)
	generation := s.clockGeneration
	s.turnTimer = time.AfterFunc(s.turnRemaining, func() { s.expireTurn(generation) })
}

func (s *GameSession) pauseTurnClock() {
	if !s.TurnDeadline.IsZero() {
		s.turnRemaining = max(0, time.Until(s.TurnDeadline))
	}
	s.stopTurnTimer()
	s.TurnDeadline = time.Time{}
}

func (s *GameSession) clearTurnClock() {
	s.stopTurnTimer()
	s.TurnDeadline = time.Time{}
	s.turnRemaining = 0
}

func (s *GameSession) expireTurn(generation uint64) {
	s.lock.Lock()
	if s.closed || s.Status != models.GameStatusPlaying || generation != s.clockGeneration || s.CurrentTurn == nil {
		s.lock.Unlock()
		return
	}
	s.advanceUnitTurn()
	s.spawnMilestoneOrbs()
	s.changed()
	s.lock.Unlock()
	_ = s.BroadcastState()
}

// Close releases timers when a session is removed.
func (s *GameSession) Close() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.closed = true
	s.clearTurnClock()
}
