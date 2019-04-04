package main

import (
	"fmt"
	"log"
	"context"
	"github.com/gin-gonic/gin"
)

type Required struct {
	Coin   string   `json:"coin" binding:"required"`
	Message   string `json:"message" binding:"required"`
	Key   string `json:"Key" binding:"required"`
}


func main() {
	r := gin.Default()
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"ethAPI": "password",
	}))
	authorized.POST("/getAddress", CreateAddress)

	r.Run(":8080")
}

func CreateAddress(c *gin.Context) {
	var required Required
	if err := c.BindJSON(&required); err != nil || required.Coin != "ETH" || required.Message != "pixi_get_address" || required.Key != "seg_pixiu" {
		c.JSON(200, gin.H{
			"success": false,
			"message": "Access Denied",
		})
	}
}