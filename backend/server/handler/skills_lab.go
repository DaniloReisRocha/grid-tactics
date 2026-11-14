package handler

import (
	"net/http"

	"github.com/DaniloReisRocha/grid-tactics/backend/internal/utils/endpoint"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/DaniloReisRocha/grid-tactics/backend/service"
	"github.com/gin-gonic/gin"
)

// These endpoints simulate only the supplied scenario. They never access a game
// session or persist changes, and can be used without joining a multiplayer game.
type SkillsLabHandler struct{}

func (h SkillsLabHandler) Config() endpoint.APIEndpoint {
	return endpoint.APIEndpoint{NewEndpoint: func() (string, string, gin.HandlerFunc) {
		return http.MethodGet, "/api/skills-lab", func(c *gin.Context) {
			seed := c.DefaultQuery("seed", "skills-lab")
			layout := c.DefaultQuery("layout", "arena")
			if len(seed) > 100 || (layout != "arena" && layout != "procedural") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "seed ou tipo de cenário inválido"})
				return
			}
			c.JSON(http.StatusOK, service.SkillsLabConfig(seed, layout))
		}
	}}
}

func (h SkillsLabHandler) Attack() endpoint.APIEndpoint {
	return endpoint.APIEndpoint{NewEndpoint: func() (string, string, gin.HandlerFunc) {
		return http.MethodPost, "/api/skills-lab/attack", func(c *gin.Context) {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
			var input models.SkillsLabAttack
			if err := c.ShouldBindJSON(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "cenário inválido"})
				return
			}
			result, err := service.ResolveSkillsLabAttack(input)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, result)
		}
	}}
}
