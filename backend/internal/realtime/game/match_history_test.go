package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
)

func readHistory(t *testing.T, s *GameSession) models.MatchHistory {
	t.Helper()
	data, err := os.ReadFile(s.history.path)
	if err != nil {
		t.Fatal(err)
	}
	var history models.MatchHistory
	if err := json.Unmarshal(data, &history); err != nil {
		t.Fatal(err)
	}
	return history
}

func historyState(t *testing.T, data json.RawMessage) models.GameStatePayload {
	t.Helper()
	var state models.GameStatePayload
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	return state
}

func TestHistoryPersistsLobbyRoundsTimeoutReconnectAndForfeit(t *testing.T) {
	directory := t.TempDir()
	manager, err := NewSessionManagerWithHistory(directory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(manager.Close)
	s := manager.CreateSession("../../room:unsafe")
	if filepath.Dir(s.history.path) != directory {
		t.Fatal("room code escaped history directory")
	}
	a, b := uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4())
	initial := readHistory(t, s)
	if initial.SchemaVersion != 1 || initial.Seed != s.ID || len(initial.Rounds) != 0 || len(initial.LobbyEvents) != 1 {
		t.Fatal("missing lobby metadata at creation")
	}
	if err := s.AddPlayer(nil, a, []models.Unit{testUnit("a")}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddPlayer(nil, b, []models.Unit{testUnit("b")}); err != nil {
		t.Fatal(err)
	}
	if err := s.EndTurn(*s.CurrentTurn); err != nil {
		t.Fatal(err)
	}
	// Exercise the real timer callback and its persistence, without client input.
	s.lock.Lock()
	s.startTurnClock(10 * time.Millisecond)
	s.lock.Unlock()
	deadline := time.Now().Add(3 * time.Second)
	for {
		s.lock.RLock()
		done := s.RoundCounter == 2
		s.lock.RUnlock()
		if done {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timeout did not persist next round")
		}
		time.Sleep(5 * time.Millisecond)
	}
	s.Disconnect(nil, false)
	for id, p := range s.Players {
		if !p.Connected {
			if err := s.AddPlayer(nil, id, nil); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	s.Disconnect(nil, true)
	h := readHistory(t, s)
	if len(h.LobbyEvents) != 3 || len(h.Rounds) != 2 || len(h.Rounds[0].Turns) != 2 || len(h.Rounds[1].Turns) != 1 {
		t.Fatalf("wrong lobby/round/turn grouping: %+v", h.Rounds)
	}
	if h.Rounds[0].Turns[0].EndReason != models.HistoryEndTurn || h.Rounds[0].Turns[1].EndReason != models.HistoryTimeout {
		t.Fatal("end-turn events attributed to the wrong turn")
	}
	if h.Status != models.GameStatusFinished || h.FinishedAt == nil || h.Winner == nil || h.Rounds[1].EndedAt == nil {
		t.Fatal("missing final result")
	}
	last := h.Rounds[1].Turns[0]
	if len(last.Events) != 3 || last.Events[0].Type != models.HistoryDisconnect || last.Events[1].Type != models.HistoryReconnect || last.Events[2].Type != models.HistoryLeave {
		t.Fatal("missing pause/resume/forfeit lifecycle")
	}
	// Playback is a chronological fold over authoritative states, independent
	// of RNG and today's combat rules. Every revision must appear exactly once.
	events := append([]models.HistoryEvent{}, h.LobbyEvents...)
	for _, round := range h.Rounds {
		for _, turn := range round.Turns {
			events = append(events, turn.Events...)
		}
	}
	var replay models.GameStatePayload
	for i, event := range events {
		if event.Sequence != uint64(i+1) || event.Revision != uint64(i) {
			t.Fatal("missing or reordered revision")
		}
		replay = historyState(t, event.State)
		if replay.Revision != event.Revision {
			t.Fatal("snapshot revision mismatch")
		}
	}
	if !reflect.DeepEqual(replay, historyState(t, h.LatestState)) {
		t.Fatal("replay did not reproduce latest state")
	}
	manager.RemoveSession(s.ID)
	closed := readHistory(t, s)
	if closed.ClosedAt == nil || closed.FinishedAt == nil {
		t.Fatal("archive lost after session removal")
	}
	newSession := manager.CreateSession(s.ID)
	if newSession.history.path == s.history.path {
		t.Fatal("reused room overwrote previous match")
	}
	entries, err := filepath.Glob(filepath.Join(directory, "*.json"))
	if err != nil || len(entries) != 2 {
		t.Fatalf("expected two separate archives: %v %v", entries, err)
	}
}

func TestHistoryPreservesMovementSkillsDamageAndWinningAttack(t *testing.T) {
	s, a, b := startedSession(t)
	actor, enemy := &s.Players[a].Units[0], &s.Players[b].Units[0]
	actor.Position, enemy.Position = models.Position{X: 1, Y: 1}, models.Position{X: 3, Y: 1}
	ability, _ := models.AbilityByID(models.AbilityBarricadeID)
	actor.Unit.Abilities = []models.Ability{ability}
	s.enableHistory(t.TempDir())
	if err := s.MovePiece(a, "a", models.Position{X: 2, Y: 1}); err != nil {
		t.Fatal(err)
	}
	if err := s.UseAbility(a, "a", ability.ID, models.Position{X: 2, Y: 2}); err != nil {
		t.Fatal(err)
	}
	if err := s.Attack(a, "a", models.Position{X: 3, Y: 1}); err != nil {
		t.Fatal(err)
	}
	h := readHistory(t, s)
	turn := h.Rounds[0].Turns[0]
	if len(turn.Events) != 3 || turn.Events[0].Type != models.HistoryMove || turn.Events[1].AbilityID != ability.ID || turn.Events[2].Type != models.HistoryAttack {
		t.Fatal("accepted commands missing")
	}
	moved := historyState(t, turn.Events[0].State).Players[a].Units[0]
	if len(moved.Path) != 2 || moved.Position.X != 2 || moved.CurrentStamina != actor.Unit.EffectiveStamina()-1 {
		t.Fatal("movement path or stamina not archived")
	}
	used := historyState(t, turn.Events[1].State)
	if used.Players[a].Units[0].AbilityUses[ability.ID] != 1 {
		t.Fatal("ability consumption missing")
	}
	found := false
	for _, cell := range used.Board {
		if cell.X == 2 && cell.Y == 2 && cell.Object == models.ObjectBarricade {
			found = true
		}
	}
	if !found {
		t.Fatal("terrain mutation missing")
	}
	final := historyState(t, h.LatestState)
	if final.Players[b].Units[0].CurrentLife > 0 || h.Winner == nil || *h.Winner != a || turn.EndReason != models.HistoryAttack {
		t.Fatal("winning damage/result not archived")
	}
	// Previously serialized states must not alias later mutable board/unit data.
	first := historyState(t, turn.InitialState)
	if first.Players[b].Units[0].CurrentLife != 3 || first.Players[a].Units[0].Position.X != 1 {
		t.Fatal("initial state mutated")
	}
	before, err := os.ReadFile(s.history.path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Attack(a, "a", models.Position{}); err == nil {
		t.Fatal("expected rejected attack")
	}
	after, _ := os.ReadFile(s.history.path)
	if string(before) != string(after) {
		t.Fatal("rejected command polluted history")
	}
}

func TestHistoryRetriesFailedSaveWithoutLosingEvents(t *testing.T) {
	s, _, _ := startedSession(t)
	s.enableHistory(t.TempDir())
	path := s.history.path
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s.history.path = filepath.Join(filepath.Dir(path), "missing", "game.json")
	if err := s.EndTurn(*s.CurrentTurn); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("failed write damaged last valid archive")
	}
	s.history.path = path
	if err := s.EndTurn(*s.CurrentTurn); err != nil {
		t.Fatal(err)
	}
	h := readHistory(t, s)
	if len(h.Rounds) != 2 || len(h.Rounds[0].Turns[0].Events) != 1 || len(h.Rounds[0].Turns[1].Events) != 1 {
		t.Fatal("failed persistence discarded an accepted action")
	}
	entries, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".history-*"))
	if len(entries) != 0 {
		t.Fatal("temporary files leaked")
	}
}

func TestHistoryDirectoryMustBeWritableDirectory(t *testing.T) {
	if _, err := NewSessionManagerWithHistory(""); err == nil {
		t.Fatal("empty history directory accepted")
	}
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("existing file"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewSessionManagerWithHistory(file); err == nil {
		t.Fatal("regular file accepted as history directory")
	}
}
