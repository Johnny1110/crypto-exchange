package service

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/settings"
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func Register(request dto.RegisterReq) {
	// check if user exist
	var exists int
	err := db.QueryRow("SELECT 1 FROM users WHERE username = ?", req.Username).Scan(&exists)
	if err != sql.ErrNoRows {
		log.Warnf("[User] Register failed %s", err.Error())
		c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
		return
	}

	// gen hash pwd
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// gen userId
	userID := uuid.NewString()
	_, err = db.Exec(
		`INSERT INTO users (id, username, password_hash) VALUES (?, ?, ?)`,
		userID, req.Username, string(hash),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// create balances data for user
	allAssets := settings.GetAllAssets()
	err = createBalanceForUser(db, userID, allAssets)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": userID, "username": req.Username})
}
