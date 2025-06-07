package middleware

import (
	"errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/controller"
	"github.com/johnny1110/crypto-exchange/security"
	"github.com/labstack/gommon/log"
	"net/http"
	"strings"
)

// CORS middleware
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins:  true, // allow all
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
	})
}

// ErrorHandler middleware
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			switch err.Type {
			case gin.ErrorTypePublic:
				c.JSON(c.Writer.Status(), controller.HandleCodeError(controller.BAD_REQUEST, err))
			case gin.ErrorTypeBind:
				c.JSON(http.StatusBadRequest, controller.HandleCodeError(controller.BAD_REQUEST, err))
			default:
				c.JSON(http.StatusInternalServerError, controller.HandleCodeError(controller.SYSTEM_ERROR, err))
			}
		}
	}
}

// AuthMiddleware middleware
func AuthMiddleware(credentialCache *security.CredentialCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			//Authorization header required
			c.JSON(http.StatusInternalServerError, controller.HandleCodeError(controller.ACCESS_DENIED, errors.New("authorization header required")))
			c.Abort()
			return
		}

		// support Bearer token
		token := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		user, err := credentialCache.Get(token)
		if err != nil || user == nil {
			// Invalid or expired token
			c.JSON(http.StatusUnauthorized, controller.HandleCodeError(controller.ACCESS_DENIED, errors.New("invalid or expired token")))
			c.Abort()
			return
		}

		// store user info and token into context
		c.Set("user", user)
		c.Set("userId", user.ID)
		c.Set("username", user.Username)
		c.Set("token", token)

		log.Infof("Logged in user: %s", user.Username)

		c.Next()
	}
}

// AdminMiddleware middleware
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		adminToken := c.GetHeader("Admin-Token")
		if adminToken == "" {
			//Authorization header required
			c.JSON(http.StatusInternalServerError, controller.HandleCodeError(controller.ACCESS_DENIED, errors.New("authorization header required")))
			c.Abort()
			return
		}

		// TODO: revamp this, make it better.
		if adminToken != "frizo" {
			// Invalid or expired token
			c.JSON(http.StatusUnauthorized, controller.HandleCodeError(controller.ACCESS_DENIED, errors.New("invalid or expired token")))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware middleware
func RateLimitMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// TODO: IP check
		c.Next()
	})
}
