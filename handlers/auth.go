package handlers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// AuthMiddleware check header token, and also inject userId into context
// Need register in main.go
func AuthMiddleware(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	token := c.GetHeader("Authorization")

	var userID string
	err := db.QueryRow("SELECT id FROM users WHERE token = ?", token).Scan(&userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	// Inject userID，continue handler
	c.Set("userID", userID)

	log.Println("Logged in userID:", userID)

	c.Next()
}
