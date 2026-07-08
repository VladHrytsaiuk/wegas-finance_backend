package controllers

import (
	"log"
	"net/http"
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSController struct {
	Hub      *utils.WSHub
	upgrader websocket.Upgrader
}

func NewWSController(hub *utils.WSHub, allowedOrigins []string) *WSController {
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		originSet[origin] = struct{}{}
	}

	return &WSController{
		Hub: hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := strings.TrimSpace(r.Header.Get("Origin"))
				if origin == "" {
					return true
				}

				_, allowed := originSet[origin]
				return allowed
			},
		},
	}
}

func (ctrl *WSController) HandleWS(c *gin.Context) {
	familyID := c.GetString("familyID")
	if familyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := ctrl.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ Failed to upgrade to websocket: %v", err)
		return
	}

	ctrl.Hub.Register(familyID, conn)

	// Чекаємо на закриття з'єднання
	go func() {
		defer ctrl.Hub.Unregister(familyID, conn)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
