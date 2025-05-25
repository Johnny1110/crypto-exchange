package handlers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func Register(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if user exist
	var exists int
	err := db.QueryRow("SELECT 1 FROM users WHERE username = ?", req.Username).Scan(&exists)
	if err != sql.ErrNoRows {
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

	c.JSON(http.StatusCreated, gin.H{"id": userID, "username": req.Username})
}

func Login(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// select user by username
	var userID, hash string
	err := db.QueryRow(
		"SELECT id, password_hash FROM users WHERE username = ?",
		req.Username,
	).Scan(&userID, &hash)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// validate pwd
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// 生成并更新 token
	token := uuid.NewString()
	_, err = db.Exec(
		"UPDATE users SET token = ? WHERE id = ?",
		token, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func Logout(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	// extract token from header ("Authorization")
	token := c.GetHeader("Authorization")

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorization header missing"})
		return
	}

	res, err := db.Exec(
		`UPDATE users SET token = NULL WHERE token = ?`,
		token,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	c.Status(http.StatusNoContent)
}
