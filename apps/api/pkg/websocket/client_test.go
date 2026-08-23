package websocket_test

import (
	"net/http/httptest"
	"testing"

	"github.com/kisssonik/hearts/pkg/websocket"
	"github.com/stretchr/testify/assert"
)

func TestNewUpgrader_CheckOrigin(t *testing.T) {
	upgrader := websocket.NewUpgrader([]string{"http://localhost:5173"})

	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{"allowed origin", "http://localhost:5173", true},
		{"unknown origin is rejected", "http://evil.example.com", false},
		{"scheme must match exactly", "https://localhost:5173", false},
		{"port must match exactly", "http://localhost:3000", false},
		{"no Origin header means a non-browser client", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/ws", nil)
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}
			assert.Equal(t, tt.want, upgrader.CheckOrigin(r))
		})
	}
}
