package handlers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ManualAdjustment(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	var req settlementReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Secret != "frizo" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "secret invalid"})
	}

	// validate username exist
	var userID string
	err := db.QueryRow(
		"SELECT id FROM users WHERE username = ?",
		req.Username,
	).Scan(&userID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query user"})
		return
	}

	_, err = db.Exec(`
        INSERT INTO balances(user_id, asset, available, locked)
        VALUES(?, ?, ?, 0)
        ON CONFLICT(user_id, asset) DO UPDATE
          SET available = available + excluded.available
    `, userID, req.Asset, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to adjust balance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}
