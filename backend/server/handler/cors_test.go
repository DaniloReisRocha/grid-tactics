package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func TestCrossOriginTeamCatalogWithNgrokHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	origin := "https://7491-2804-593c-c003-7c98-d842-3b62-52d4-45e3.ngrok-free.app"
	router := gin.New()
	router.Use(cors.New(apiCORSConfig([]string{"https://*.ngrok-free.app"})))
	method, path, handler := (TeamHandler{}).Options().NewEndpoint()
	router.Handle(method, path, handler)
	preflight := httptest.NewRequest(http.MethodOptions, "/api/team-options", nil)
	preflight.Header.Set("Origin", origin)
	preflight.Header.Set("Access-Control-Request-Method", "GET")
	preflight.Header.Set("Access-Control-Request-Headers", "ngrok-skip-browser-warning")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, preflight)
	if response.Code != http.StatusNoContent || response.Header().Get("Access-Control-Allow-Origin") != origin || !strings.Contains(strings.ToLower(response.Header().Get("Access-Control-Allow-Headers")), "ngrok-skip-browser-warning") {
		t.Fatalf("preflight blocked: %d %v", response.Code, response.Header())
	}
	get := httptest.NewRequest(http.MethodGet, "/api/team-options", nil)
	get.Header.Set("Origin", origin)
	get.Header.Set("ngrok-skip-browser-warning", "1")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, get)
	if response.Code != http.StatusOK || response.Header().Get("Access-Control-Allow-Origin") != origin || !strings.Contains(response.Body.String(), `"weapons"`) {
		t.Fatal("catalog request failed")
	}
	get.Header.Set("Origin", "https://not-allowed.example")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, get)
	if response.Code != http.StatusForbidden {
		t.Fatal("unlisted origin accepted")
	}
}
