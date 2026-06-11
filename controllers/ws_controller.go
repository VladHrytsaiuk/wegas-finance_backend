package controllers

import (
	"log"
	"net/http"

	"github.com/VladHrytsaiuk/wegas-finance/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // У реальному проекті обмежте Origin
	},
}

type WSController struct {
	Hub *utils.WSHub
}

func NewWSController(hub *utils.WSHub) *WSController {
	return &WSController{Hub: hub}
}

func (ctrl *WSController) HandleWS(c *gin.Context) {
	familyID := c.GetString("familyID")
	if familyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
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
