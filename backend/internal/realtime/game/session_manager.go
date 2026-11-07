package game

import (
	"sync"

	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime"
)

type SessionManager struct {
	sessions map[string]*GameSession
	lock     *sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*GameSession),
		lock:     &sync.RWMutex{},
	}
}

func (sm *SessionManager) CreateSession(id string) *GameSession {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	if session, ok := sm.sessions[id]; ok {
		return session
	}

	session := NewGameSession(id, 2)
	sm.sessions[id] = session
	return session
}

func (sm *SessionManager) GetSession(id string) (*GameSession, bool) {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	session, ok := sm.sessions[id]
	return session, ok
}

func (sm *SessionManager) FindSessionByConn(conn *realtime.WSConnection) *GameSession {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	for _, session := range sm.sessions {
		session.lock.RLock()
		for _, player := range session.Players {
			if player.Connection == conn {
				session.lock.RUnlock()
				return session
			}
		}
		session.lock.RUnlock()
	}

	return nil
}

func (sm *SessionManager) RemoveSession(id string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	delete(sm.sessions, id)
}
