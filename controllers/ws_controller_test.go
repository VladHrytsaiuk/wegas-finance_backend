package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestWSController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := utils.NewWSHub()
	go hub.Run()
	controller := NewWSController(hub, []string{"http://127.0.0.1"})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("familyID", "test-family")
		c.Next()
	})
	r.GET("/ws", controller.HandleWS)

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	t.Run("Connect to WS", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("Origin", "http://127.0.0.1")

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
		assert.NoError(t, err)
		defer conn.Close()

		// Allow some time for registration
		time.Sleep(50 * time.Millisecond)

		// Broadcast something and check if it arrives
		msg := map[string]string{"type": "ping"}
		hub.BroadcastToFamily("test-family", msg)

		var received map[string]string
		err = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		assert.NoError(t, err)
		err = conn.ReadJSON(&received)
		assert.NoError(t, err)
		assert.Equal(t, "ping", received["type"])
	})
}
