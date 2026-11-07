package service

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime"
	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime/game"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
)

const (
	echo         = "echo"
	joinGame     = "joinGame"
	movePiece    = "movePiece"
	attackPiece  = "attackPiece"
	endTurn      = "endTurn"
	leaveGame    = "leaveGame"
	actionFailed = "actionFailed"
)

var SessionManager = game.NewSessionManager()

type joinGameAction struct {
	SessionID string `json:"sessionID"`
	AccoID    string `json:"accoID"`
}
type movePieceAction struct {
	UnitID string          `json:"unitID"`
	To     models.Position `json:"to"`
}
type GameWsService struct {
	Logger         *log.Logger
	GetUnits       models.GameUnitsGet
	AllowedOrigins []string
}
type functionResolver func(*realtime.WSConnection, realtime.ActionReceived) error

func (gs *GameWsService) Receive(w http.ResponseWriter, r *http.Request) {
	conn, err := realtime.NewWSConnection(w, r, gs.AllowedOrigins)
	if err != nil {
		return
	}
	defer conn.Close()
	defer func() {
		if session := SessionManager.FindSessionByConn(conn); session != nil {
			session.Disconnect(conn, false)
		}
	}()
	actions := map[string]functionResolver{
		echo: receiveAction(func(conn *realtime.WSConnection, payload json.RawMessage) error {
			return conn.Send(realtime.ActionSent{Type: echo, Payload: payload})
		}),
		joinGame:    receiveAction(gs.joinGame),
		movePiece:   receiveAction(gs.movePiece),
		attackPiece: receiveAction(gs.attack),
		endTurn: receiveAction(func(conn *realtime.WSConnection, _ struct{}) error {
			return gs.inSession(conn, func(session *game.GameSession) error { return session.EndTurn(conn.AccoID) })
		}),
		leaveGame: receiveAction(func(conn *realtime.WSConnection, _ struct{}) error {
			if session := SessionManager.FindSessionByConn(conn); session != nil {
				session.Disconnect(conn, true)
			}
			return nil
		}),
	}
	conn.ReadPump(func(message []byte) {
		var action realtime.ActionReceived
		if err := json.Unmarshal(message, &action); err != nil {
			_ = conn.Send(realtime.ActionSent{Type: actionFailed, Payload: actionFailedPayload{Message: "mensagem JSON inválida"}})
			return
		}
		resolver, ok := actions[action.Type]
		if !ok {
			_ = conn.Send(realtime.ActionSent{Type: actionFailed, Payload: actionFailedPayload{Action: action.Type, Message: "ação desconhecida"}})
			return
		}
		if err := resolver(conn, action); err != nil && gs.Logger != nil {
			gs.Logger.Printf("action %s: %v", action.Type, err)
		}
	})
}
func (gs *GameWsService) joinGame(conn *realtime.WSConnection, payload joinGameAction) error {
	payload.SessionID = strings.TrimSpace(payload.SessionID)
	if payload.SessionID == "" || len(payload.SessionID) > 100 {
		return errors.New("sessão deve ter entre 1 e 100 caracteres")
	}
	id, err := uuid.FromString(payload.AccoID)
	if err != nil || id == uuid.Nil {
		return errors.New("conta deve ser um UUID válido")
	}
	if previous := SessionManager.FindSessionByConn(conn); previous != nil {
		if previous.ID != payload.SessionID || conn.AccoID != id {
			return errors.New("saia da partida atual antes de entrar em outra")
		}
	}
	units, err := gs.GetUnits(id)
	if err != nil {
		return err
	}
	session := SessionManager.CreateSession(payload.SessionID)
	if err := session.AddPlayer(conn, id, units); err != nil {
		return err
	}
	conn.AccoID = id
	if err := session.BroadcastState(); err != nil && gs.Logger != nil {
		gs.Logger.Print(err)
	}
	return nil
}
func (gs *GameWsService) inSession(conn *realtime.WSConnection, action func(*game.GameSession) error) error {
	session := SessionManager.FindSessionByConn(conn)
	if session == nil {
		return errors.New("sessão não encontrada")
	}
	if err := action(session); err != nil {
		return err
	}
	if err := session.BroadcastState(); err != nil && gs.Logger != nil {
		gs.Logger.Print(err)
	}
	return nil
}
func (gs *GameWsService) movePiece(conn *realtime.WSConnection, payload movePieceAction) error {
	return gs.inSession(conn, func(session *game.GameSession) error {
		return session.MovePiece(conn.AccoID, payload.UnitID, payload.To)
	})
}
func (gs *GameWsService) attack(conn *realtime.WSConnection, payload movePieceAction) error {
	return gs.inSession(conn, func(session *game.GameSession) error { return session.Attack(conn.AccoID, payload.UnitID, payload.To) })
}

type actionFailedPayload struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

func receiveAction[A any](resolver func(*realtime.WSConnection, A) error) functionResolver {
	return func(conn *realtime.WSConnection, action realtime.ActionReceived) (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("WebSocket action panic: %v", recovered)
				err = errors.New("falha interna ao processar ação")
			}
			if err != nil {
				_ = conn.Send(realtime.ActionSent{Type: actionFailed, Payload: actionFailedPayload{Action: action.Type, Message: err.Error()}})
			}
		}()
		var payload A
		raw := action.Payload
		if len(raw) == 0 {
			raw = json.RawMessage("{}")
		}
		if err = json.Unmarshal(raw, &payload); err != nil {
			return errors.New("payload inválido")
		}
		return resolver(conn, payload)
	}
}
