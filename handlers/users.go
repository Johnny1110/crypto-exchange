package handlers

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/settings"
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func Register(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

// CreateBalanceForUser creates initial zero balances for a new user across given assets.
func createBalanceForUser(db *sql.DB, userID string, assets []string) error {
	// Use a transaction to batch inserts
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Prepare statement once
	stmt, err := tx.Prepare(`
        INSERT INTO balances(user_id, asset, available, locked)
        VALUES (?, ?, 0, 0)
        ON CONFLICT(user_id, asset) DO NOTHING
    `)
	if err != nil {
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer stmt.Close()

	// Execute for each asset
	for _, asset := range assets {
		if _, err = stmt.Exec(userID, asset); err != nil {
			return fmt.Errorf("insert balance %s: %w", asset, err)
		}
	}
	return nil
}

func Login(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	var req dto.LoginReq
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
