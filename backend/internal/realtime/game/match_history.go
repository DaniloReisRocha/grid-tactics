package game

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
)

type matchRecorder struct {
	path     string
	history  models.MatchHistory
	sequence uint64
}

// Server sessions persist locally; the ordinary constructor keeps tests and
// isolated simulations in memory.
func NewSessionManagerWithHistory(directory string) (*SessionManager, error) {
	if directory == "" {
		return nil, fmt.Errorf("history directory must not be empty")
	}
	directory, err := filepath.Abs(directory)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, fmt.Errorf("create history directory: %w", err)
	}
	probe, err := os.CreateTemp(directory, ".history-probe-*")
	if err != nil {
		return nil, fmt.Errorf("write history directory: %w", err)
	}
	probe.Close()
	os.Remove(probe.Name())
	manager := NewSessionManager()
	manager.historyDirectory = directory
	return manager, nil
}

func (s *GameSession) enableHistory(directory string) {
	id := uuid.Must(uuid.NewV4()).String()
	now := s.StartedAt.UTC()
	s.history = &matchRecorder{
		// Room codes are untrusted and never participate in the filesystem path.
		path: filepath.Join(directory, now.Format("20060102T150405.000000000Z")+"_"+id+".json"),
		history: models.MatchHistory{
			SchemaVersion: 1, ID: id, SessionID: s.ID, Seed: s.ID,
			MaximumPlayers: s.maximumPlayers, TurnDurationMS: TurnDuration.Milliseconds(),
			CreatedAt: now, LobbyEvents: []models.HistoryEvent{}, Rounds: []models.HistoryRound{},
		},
	}
	s.recordHistory(models.HistoryAction{Type: models.HistoryCreated})
}

// Caller holds the session lock. Serializing and saving synchronously ensures
// neither timer callbacks nor concurrent actions can overtake a revision.
func (s *GameSession) recordHistory(action models.HistoryAction) {
	r := s.history
	payload := s.snapshot()
	state, err := json.Marshal(payload)
	if err != nil {
		log.Printf("history session=%q: serialize state: %v", s.ID, err)
		return
	}
	now := time.Now().UTC()
	h := &r.history
	r.sequence++
	event := models.HistoryEvent{Sequence: r.sequence, At: now, Revision: s.Revision, HistoryAction: action, State: state}
	var previous *models.HistoryTurn
	if len(h.Rounds) > 0 {
		round := &h.Rounds[len(h.Rounds)-1]
		previous = &round.Turns[len(round.Turns)-1]
	}
	if previous == nil {
		h.LobbyEvents = append(h.LobbyEvents, event)
	} else {
		previous.Events = append(previous.Events, event)
	}
	finished := s.Status == models.GameStatusFinished
	newTurn := s.CurrentTurn != nil && s.CurrentUnitID != "" &&
		(previous == nil || previous.Number != s.TurnCounter)
	if previous != nil && previous.EndedAt == nil && (newTurn || finished || s.closed) {
		previous.EndedAt, previous.EndReason = &now, action.Type
	}
	if newTurn {
		if h.StartedAt == nil {
			h.StartedAt = &now
		}
		if len(h.Rounds) == 0 || h.Rounds[len(h.Rounds)-1].Number != s.RoundCounter {
			if len(h.Rounds) > 0 {
				h.Rounds[len(h.Rounds)-1].EndedAt = &now
			}
			h.Rounds = append(h.Rounds, models.HistoryRound{
				Number: s.RoundCounter, StartedAt: now, Turns: []models.HistoryTurn{},
			})
		}
		round := &h.Rounds[len(h.Rounds)-1]
		round.Turns = append(round.Turns, models.HistoryTurn{
			Number: s.TurnCounter, RoundSlot: s.RoundTurn, PlayerID: *s.CurrentTurn,
			UnitID: s.CurrentUnitID, StartedAt: now, InitialState: state, Events: []models.HistoryEvent{},
		})
	}
	if finished && h.FinishedAt == nil {
		h.FinishedAt = &now
		if len(h.Rounds) > 0 {
			h.Rounds[len(h.Rounds)-1].EndedAt = &now
		}
	}
	if s.closed {
		h.ClosedAt = &now
		if len(h.Rounds) > 0 && h.Rounds[len(h.Rounds)-1].EndedAt == nil {
			h.Rounds[len(h.Rounds)-1].EndedAt = &now
		}
	}
	if h.InitialState == nil {
		h.InitialState = state
	}
	h.LatestState, h.Status, h.UpdatedAt, h.Winner = state, s.Status, now, payload.Winner
	if err := r.save(); err != nil {
		// Retain all events in memory; the next change retries the complete file.
		log.Printf("history session=%q path=%q: %v", s.ID, r.path, err)
	}
}

func (r *matchRecorder) save() error {
	file, err := os.CreateTemp(filepath.Dir(r.path), ".history-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(r.history); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	// Same-directory replacement preserves the previous valid JSON until the
	// complete new document is synced, including on Windows.
	return os.Rename(file.Name(), r.path)
}
