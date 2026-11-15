package origin

import "testing"

func TestAllowed(t *testing.T) {
	patterns := []string{"http://localhost:5173", " https://*.ngrok-free.app "}
	for _, tc := range []struct {
		value string
		want  bool
	}{
		{"http://localhost:5173", true},
		{"https://old-tunnel.ngrok-free.app", true},
		{"https://new-tunnel.ngrok-free.app", true},
		{"https://ngrok-free.app", false},
		{"https://fake-ngrok-free.app", false},
		{"https://tunnel.ngrok-free.app.attacker.test", false},
		{"https://tunnel.ngrok-free.app@attacker.test", false},
		{"https://attacker.test/path.ngrok-free.app", false},
		{"http://tunnel.ngrok-free.app", false},
		{"https://tunnel.ngrok-free.app:9999", false},
		{"https://unrelated.example", false},
		{"null", false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			if got := Allowed(tc.value, patterns); got != tc.want {
				t.Fatalf("Allowed(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}
