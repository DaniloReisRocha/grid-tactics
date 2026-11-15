package realtime

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWebSocketNgrokOrigins(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := NewWSConnection(w, r, []string{"https://*.ngrok-free.app"})
		if err == nil {
			conn.Close()
		}
	}))
	defer server.Close()
	for _, value := range []string{"https://first.ngrok-free.app", "https://second.ngrok-free.app", "https://first.ngrok-free.app.attacker.test"} {
		t.Run(value, func(t *testing.T) {
			conn, response, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), http.Header{"Origin": []string{value}})
			if conn != nil {
				defer conn.Close()
			}
			if response != nil && response.Body != nil {
				defer response.Body.Close()
			}
			if strings.HasSuffix(value, ".attacker.test") {
				if err == nil || response == nil || response.StatusCode != http.StatusForbidden {
					t.Fatal("unlisted origin accepted")
				}
			} else if err != nil {
				t.Fatalf("ngrok origin blocked: %v", err)
			}
		})
	}
}
