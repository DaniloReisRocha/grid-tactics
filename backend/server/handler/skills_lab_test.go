package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gin-gonic/gin"
)

func TestSkillsLabHTTPFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := SkillsLabHandler{}
	for _, ep := range []func() (string, string, gin.HandlerFunc){h.Config().NewEndpoint, h.Attack().NewEndpoint} {
		method, path, handler := ep()
		router.Handle(method, path, handler)
	}
	get := httptest.NewRecorder()
	router.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/skills-lab?layout=arena", nil))
	var config models.SkillsLabConfig
	if get.Code != http.StatusOK || json.Unmarshal(get.Body.Bytes(), &config) != nil || len(config.Weapons) != 5 || len(config.Board) != 225 {
		t.Fatalf("failed to load lab: %d %s", get.Code, get.Body.String())
	}
	input := models.SkillsLabAttack{Board: config.Board, AttackerID: "a", TargetID: "b", Tokens: []models.SkillToken{
		{ID: "a", Team: 0, Position: models.Position{X: 5, Y: 5}, Life: 6, WeaponID: models.WeaponShotgunID},
		{ID: "b", Team: 1, Position: models.Position{X: 6, Y: 5}, Life: 6, Shield: 3, WeaponID: models.WeaponPistolID},
	}}
	body, _ := json.Marshal(input)
	for i := 0; i < 2; i++ {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/skills-lab/attack", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, request)
		var result models.SkillsLabResult
		if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &result) != nil || len(result.Tokens) != 2 || result.Tokens[1].Shield != 0 {
			t.Fatalf("attack failed: %d %s", response.Code, response.Body.String())
		}
	}
	for _, url := range []string{"/api/skills-lab?layout=invalid", "/api/skills-lab?seed=" + string(bytes.Repeat([]byte("a"), 101))} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, url, nil))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("invalid query accepted: %s", url)
		}
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/skills-lab/attack", bytes.NewBufferString(`{"tokens":[]}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("empty scene accepted: %d", response.Code)
	}
}
