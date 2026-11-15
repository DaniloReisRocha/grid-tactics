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
	if get.Code != http.StatusOK || json.Unmarshal(get.Body.Bytes(), &config) != nil || len(config.Weapons) != 5 || len(config.Abilities) != 4 || len(config.Board) != 225 {
		t.Fatalf("failed to load lab: %d %s", get.Code, get.Body.String())
	}
	input := models.SkillsLabAttack{Board: config.Board, AttackerID: "a", TargetID: "b", Tokens: []models.SkillToken{
		{ID: "a", Team: 0, Position: models.Position{X: 5, Y: 5}, Life: 6, WeaponID: models.WeaponShotgunID},
		{ID: "b", Team: 1, Position: models.Position{X: 6, Y: 5}, Life: 6, Shield: 3, WeaponID: models.WeaponPistolID},
	}}
	input.Tokens[0].AbilityIDs = []string{models.AbilityShieldID, models.AbilityBarricadeID}
	input.Tokens[0].AbilityUses = map[string]int{models.AbilityShieldID: 1}
	input.Tokens[0].StaminaSpent = 4
	input.Tokens[1].AbilityIDs = []string{"", models.AbilityTrampolineID}
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
		if len(result.Tokens[0].AbilityIDs) != 2 || result.Tokens[0].AbilityIDs[0] != models.AbilityShieldID || result.Tokens[0].AbilityIDs[1] != models.AbilityBarricadeID || len(result.Tokens[1].AbilityIDs) != 2 || result.Tokens[1].AbilityIDs[0] != "" || result.Tokens[1].AbilityIDs[1] != models.AbilityTrampolineID {
			t.Fatal("attack response did not preserve selected ability slots")
		}
		if result.Tokens[0].StaminaSpent != 4 {
			t.Fatal("attack response lost spent stamina")
		}
		if result.Tokens[0].AbilityUses[models.AbilityShieldID] != 1 {
			t.Fatal("weapon attack renewed a consumed ability")
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

func TestSkillsLabAbilityEndpointPreservesUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := SkillsLabHandler{}
	method, path, handler := h.Ability().NewEndpoint()
	router.Handle(method, path, handler)
	input := models.SkillsLabAbility{ActorID: "a", AbilityID: models.AbilityTrampolineID, Target: models.Position{X: 2, Y: 1}, Board: []models.Cell{{X: 1, Y: 1, Terrain: models.TerrainGrass}, {X: 2, Y: 1, Height: 1, Terrain: models.TerrainGrass}}, Tokens: []models.SkillToken{
		{ID: "a", Position: models.Position{X: 1, Y: 1}, Life: 3, MaxShield: 3, WeaponID: models.WeaponPistolID, AbilityIDs: []string{models.AbilityTrampolineID}, StaminaSpent: 4},
	}}
	for attempt := 0; attempt < 2; attempt++ {
		body, _ := json.Marshal(input)
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/skills-lab/ability", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, request)
		if attempt == 1 {
			if response.Code != http.StatusBadRequest {
				t.Fatal("repeated single-use ability accepted")
			}
			break
		}
		var result models.SkillsLabAbilityResult
		if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &result) != nil || len(result.Tokens) != 1 || result.Tokens[0].AbilityUses[models.AbilityTrampolineID] != 1 || result.Tokens[0].StaminaSpent != 4 {
			t.Fatalf("ability request failed: %s", response.Body.String())
		}
		input.Board, input.Tokens, input.MatchAbilityUses = result.Board, result.Tokens, result.MatchAbilityUses
	}
}

func TestSkillsLabAttackCanTargetAndRemoveBarricade(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := SkillsLabHandler{}
	method, path, handler := h.Attack().NewEndpoint()
	router.Handle(method, path, handler)
	point := models.Position{X: 2, Y: 1}
	input := models.SkillsLabAttack{AttackerID: "a", Target: &point, Board: []models.Cell{
		{X: 1, Y: 1, Terrain: models.TerrainGrass}, {X: 2, Y: 1, Terrain: models.TerrainGrass, Object: models.ObjectBarricade},
	}, Tokens: []models.SkillToken{{ID: "a", Position: models.Position{X: 1, Y: 1}, Life: 3, WeaponID: models.WeaponPistolID}}}
	body, _ := json.Marshal(input)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/skills-lab/attack", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)
	var result models.SkillsLabResult
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &result) != nil || len(result.Board) != 2 || result.Board[1].Object != models.ObjectNone || result.Tokens[0].Life != 3 {
		t.Fatalf("invalid barricade response: %s", response.Body.String())
	}
}
