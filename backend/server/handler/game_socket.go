package handler

import (
	"net/http"

	"github.com/DaniloReisRocha/grid-tactics/backend/internal/utils/endpoint"
)

type GameSocketHandler struct {
	gameWs func(w http.ResponseWriter, r *http.Request)
}

func (handler *GameSocketHandler) GameWs() endpoint.APIEndpoint {
	return endpoint.NewWebSocketRoute("/api/game-connect", handler.gameWs)
}
