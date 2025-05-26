package handlers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/entity"
	"log"
	"net/http"
)

func GetBalance(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	userID := c.MustGet("userID").(string)

	log.Println("GetBalance", userID)

	rows, err := db.Query(`
        SELECT asset, available, locked
        FROM balances
        WHERE user_id = ?
    `, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query balances"})
		return
	}
	defer rows.Close()

	var result []entity.Balance
	for rows.Next() {
		var b entity.Balance
		if err := rows.Scan(&b.Asset, &b.Available, &b.Locked); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "scan error"})
			return
		}
		result = append(result, b)
	}
	c.JSON(http.StatusOK, result)

}
