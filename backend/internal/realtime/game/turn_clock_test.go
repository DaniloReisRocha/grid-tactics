package game

import (
	"testing"
	"time"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
)

func TestTurnClockAutomaticallyAdvancesAndReplenishesNextUnit(t *testing.T) {
	s, _, b := startedSession(t)
	s.lock.Lock()
	if remaining := time.Until(s.TurnDeadline); remaining < 44*time.Second || remaining > TurnDuration {
		t.Fatalf("incorrect initial budget %v", remaining)
	}
	s.Players[b].Units[0].CurrentStamina = 0
	s.Players[b].Units[0].HasAttacked = true
	before := s.TurnCounter
	s.startTurnClock(30 * time.Millisecond)
	s.lock.Unlock()
	deadline := time.Now().Add(2 * time.Second)
	for {
		s.lock.RLock()
		advanced := s.TurnCounter > before
		s.lock.RUnlock()
		if advanced {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("server timer did not advance without client action")
		}
		time.Sleep(5 * time.Millisecond)
	}
	s.lock.RLock()
	defer s.lock.RUnlock()
	unit := s.Players[b].Units[0]
	if s.TurnCounter != before+1 || *s.CurrentTurn != b || unit.CurrentStamina != unit.Unit.EffectiveStamina() || unit.HasAttacked {
		t.Fatal("timeout did not start exactly one next unit turn")
	}
	if remaining := time.Until(s.TurnDeadline); remaining < 44*time.Second || remaining > TurnDuration {
		t.Fatal("next unit did not receive 45 seconds")
	}
}

func TestTurnClockDoesNotResetOnActionsOrPlayingRejoin(t *testing.T) {
	s, a, b := startedSession(t)
	unit := &s.Players[a].Units[0]
	unit.Position = models.Position{}
	s.Players[b].Units[0].Position = models.Position{X: 9, Y: 9}
	deadline := s.TurnDeadline
	if err := s.MovePiece(a, "a", models.Position{X: 1}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddPlayer(nil, a, nil); err != nil {
		t.Fatal(err)
	}
	if !s.TurnDeadline.Equal(deadline) {
		t.Fatal("movement or reconnect granted extra time")
	}
}

func TestPausedClockResumesRemainingBudget(t *testing.T) {
	s, _, _ := startedSession(t)
	s.lock.Lock()
	s.startTurnClock(time.Second)
	oldGeneration := s.clockGeneration
	before := s.TurnCounter
	s.lock.Unlock()
	s.Disconnect(nil, false)
	s.lock.RLock()
	remaining := s.turnRemaining
	var disconnected uuid.UUID
	for id, player := range s.Players {
		if !player.Connected {
			disconnected = id
		}
	}
	if s.Status != models.GameStatusPaused || !s.TurnDeadline.IsZero() || remaining <= 0 || remaining > time.Second {
		t.Fatal("disconnect did not pause remaining time")
	}
	s.lock.RUnlock()
	s.expireTurn(oldGeneration)
	time.Sleep(20 * time.Millisecond)
	if err := s.AddPlayer(nil, disconnected, nil); err != nil {
		t.Fatal(err)
	}
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.Status != models.GameStatusPlaying || s.TurnCounter != before || time.Until(s.TurnDeadline) > remaining || time.Until(s.TurnDeadline) < remaining-100*time.Millisecond {
		t.Fatal("resume reset or spent paused turn time")
	}
}

func TestExpiredAndStaleActionsCannotAdvanceTheNextTurn(t *testing.T) {
	s, a, b := startedSession(t)
	s.lock.Lock()
	oldGeneration := s.clockGeneration
	turn := s.TurnCounter
	s.TurnDeadline = time.Now().Add(-time.Millisecond)
	s.lock.Unlock()
	if s.EndTurn(a, turn) == nil || s.MovePiece(a, "a", models.Position{}) == nil || s.Attack(a, "a", models.Position{}) == nil || s.UseAbility(a, "a", models.AbilityMedkitID, models.Position{}) == nil {
		t.Fatal("accepted action after expiry")
	}
	s.expireTurn(oldGeneration)
	if s.EndTurn(b, turn) == nil {
		t.Fatal("stale end-turn request ended the new turn")
	}
	s.lock.RLock()
	next := s.TurnCounter
	s.lock.RUnlock()
	s.expireTurn(oldGeneration)
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.TurnCounter != next || next != turn+1 {
		t.Fatal("stale callback advanced twice")
	}
}

func TestFinishedOrRemovedSessionStopsTurnTimer(t *testing.T) {
	s, _, _ := startedSession(t)
	s.lock.Lock()
	generation := s.clockGeneration
	s.Status = models.GameStatusFinished
	s.changed()
	before := s.TurnCounter
	s.lock.Unlock()
	s.expireTurn(generation)
	s.lock.RLock()
	if s.turnTimer != nil || !s.TurnDeadline.IsZero() || s.TurnCounter != before {
		t.Fatal("finished game retained timer")
	}
	s.lock.RUnlock()
	manager := NewSessionManager()
	removed := manager.CreateSession("remove-clock")
	removed.lock.Lock()
	removed.Status = models.GameStatusPlaying
	removed.startTurnClock(time.Hour)
	removed.lock.Unlock()
	manager.RemoveSession("remove-clock")
	removed.lock.RLock()
	defer removed.lock.RUnlock()
	if !removed.closed || removed.turnTimer != nil || !removed.TurnDeadline.IsZero() {
		t.Fatal("removed session retained timer")
	}
}
