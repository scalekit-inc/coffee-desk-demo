package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ScalekitCallback(c *gin.Context) {
	var requestBody map[string]interface{}
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "SSO callback received successfully",
	})
}
