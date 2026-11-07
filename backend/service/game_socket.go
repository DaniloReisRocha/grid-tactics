package service

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime"
	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime/game"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

const (
	echo         = "echo"
	joinGame     = "joinGame"
	movePiece    = "movePiece"
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
	Logger   *log.Logger
	GetUnits models.GameUnitsGet
}

type functionResolver func(conn *realtime.WSConnection, action realtime.ActionReceived) error

func (gs *GameWsService) Receive(w http.ResponseWriter, r *http.Request) {

	conn, err := realtime.NewWSConnection(w, r)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	actionResolver := map[string]func(conn *realtime.WSConnection, action realtime.ActionReceived) error{
		echo:      func(conn *realtime.WSConnection, action realtime.ActionReceived) error { return nil },
		joinGame:  receiveAction(gs.joinGame),
		movePiece: receiveAction(gs.movePiece),
	}

	conn.ReadPump(func(msg []byte) {
		action := realtime.ActionReceived{}

		err = json.Unmarshal(msg, &action)
		if err != nil {
			log.Println("Invalid message:", err)
			return
		}
		conn.Send(realtime.ActionSent{
			Type:    echo,
			Payload: action,
		})
		if resolv, ok := actionResolver[action.Type]; ok {
			err := resolv(conn, action)
			if err != nil && gs.Logger != nil {
				gs.Logger.Printf("Failed to resolve action %s: %v", action.Type, err)
			}
		} else {
			if gs.Logger != nil {
				gs.Logger.Printf("Action type not found: %s", action.Type)
			}
		}
	})
}

func (gs *GameWsService) joinGame(conn *realtime.WSConnection, payload joinGameAction) (err error) {

	accoID, err := uuid.FromString(payload.AccoID)
	if err != nil {
		return errors.Wrap(err, "invalid account id")
	}

	units, err := gs.GetUnits(accoID)
	if err != nil {
		return err
	}

	conn.AccoID = accoID
	session := SessionManager.CreateSession(payload.SessionID)
	err = session.AddPlayer(conn, accoID, units)
	if err != nil {
		return errors.Wrap(err, "add player")
	}

	e := session.BroadcastState()
	if e != nil && gs.Logger != nil {
		gs.Logger.Printf("failed to resolve joinGame: %v", e)
	}

	return nil
}

func (gs *GameWsService) movePiece(conn *realtime.WSConnection, payload movePieceAction) (err error) {

	session := SessionManager.FindSessionByConn(conn)
	if session == nil {
		return errors.New("session not found")
	}

	err = session.MovePiece(conn.AccoID, payload.UnitID, payload.To)
	if err != nil {
		return errors.Wrap(err, "Failed to move unit")
	}

	e := session.BroadcastState()
	if e != nil && gs.Logger != nil {
		gs.Logger.Printf("failed to resolve movePiece: %v", e)
	}

	return nil
}

type actionFailedPayload struct {
	Action  string `json:"action"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func receiveAction[A any](resolver func(conn *realtime.WSConnection, action A) error) functionResolver {
	return func(conn *realtime.WSConnection, action realtime.ActionReceived) (err error) {
		defer func() {
			if r := recover(); r != nil {
				switch x := r.(type) {
				case string:
					err = errors.New(x)
				case error:
					err = x
				default:
					err = errors.New("unknown panic")
				}
			}
			if err != nil {
				e := conn.Send(realtime.ActionSent{
					Type: actionFailed,
					Payload: actionFailedPayload{
						Action:  joinGame,
						Message: err.Error(),
						Data:    action.Payload,
					},
				})
				if e != nil {
					log.Println("filed to send action error:", e)
				}
			}
		}()

		payload := new(A)
		err = json.Unmarshal(action.Payload, payload)
		if err != nil {
			return errors.Wrap(err, "parse payload")
		}

		return resolver(conn, *payload)
	}
}
