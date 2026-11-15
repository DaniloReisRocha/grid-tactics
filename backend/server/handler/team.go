package handler

import (
	"github.com/DaniloReisRocha/grid-tactics/backend/internal/utils/endpoint"
	"github.com/DaniloReisRocha/grid-tactics/backend/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TeamHandler struct{}

func (TeamHandler) Options() endpoint.APIEndpoint {
	return endpoint.APIEndpoint{NewEndpoint: func() (string, string, gin.HandlerFunc) {
		return http.MethodGet, "/api/team-options", func(c *gin.Context) { c.JSON(http.StatusOK, service.TeamOptions()) }
	}}
}
