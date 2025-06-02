package security

import (
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/handlers"
	"log"
	"net/http"
)

// AuthMiddleware check header token, and also inject userId into context
// Need register in main.go
func AuthMiddleware(c *gin.Context) {
	cc := c.MustGet("credentialCache").(*CredentialCache)
	token := c.GetHeader("Authorization")

	user, err := cc.Get(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, handlers.HandleCodeError(handlers.ACCESS_DENIED, err))
		return
	}
	// Inject user，continue handler
	c.Set("userID", user.ID)
	log.Println("Logged in userID:", user.ID)

	c.Next()
}
