package utils

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestWSHub(t *testing.T) {
	hub := NewWSHub()
	go hub.Run()

	serverConnChan := make(chan *websocket.Conn, 10)

	// Setup a test server to upgrade connections
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		familyID := r.URL.Query().Get("familyID")
		hub.Register(familyID, conn)
		serverConnChan <- conn
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	t.Run("Broadcast to Family", func(t *testing.T) {
		// Client 1 in Family A
		c1, _, err := websocket.DefaultDialer.Dial(wsURL+"?familyID=A", nil)
		assert.NoError(t, err)
		defer c1.Close()

		// Client 2 in Family A
		c2, _, err := websocket.DefaultDialer.Dial(wsURL+"?familyID=A", nil)
		assert.NoError(t, err)
		defer c2.Close()

		// Client 3 in Family B
		c3, _, err := websocket.DefaultDialer.Dial(wsURL+"?familyID=B", nil)
		assert.NoError(t, err)
		defer c3.Close()

		// Give some time to register
		time.Sleep(50 * time.Millisecond)

		msg := map[string]string{"event": "test"}
		hub.BroadcastToFamily("A", msg)

		// C1 and C2 should receive, C3 should not
		var res1, res2 map[string]string
		
		err = c1.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		assert.NoError(t, err)
		err = c1.ReadJSON(&res1)
		assert.NoError(t, err)
		assert.Equal(t, "test", res1["event"])

		err = c2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		assert.NoError(t, err)
		err = c2.ReadJSON(&res2)
		assert.NoError(t, err)
		assert.Equal(t, "test", res2["event"])

		// C3 should timeout
		var res3 map[string]string
		err = c3.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
		assert.NoError(t, err)
		err = c3.ReadJSON(&res3)
		assert.Error(t, err)
	})

	t.Run("Unregister", func(t *testing.T) {
		// Clear channel
		for len(serverConnChan) > 0 { <-serverConnChan }

		c, _, err := websocket.DefaultDialer.Dial(wsURL+"?familyID=C", nil)
		assert.NoError(t, err)
		defer c.Close()
		
		var srvConn *websocket.Conn
		select {
		case srvConn = <-serverConnChan:
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for server conn")
		}

		time.Sleep(100 * time.Millisecond)
		hub.mu.Lock()
		assert.NotNil(t, hub.families["C"])
		hub.mu.Unlock()

		hub.Unregister("C", srvConn)
		time.Sleep(100 * time.Millisecond)
		
		hub.mu.Lock()
		assert.Nil(t, hub.families["C"])
		hub.mu.Unlock()
	})
}
