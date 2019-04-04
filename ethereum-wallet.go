package main

import (
	"fmt"
	"log"
	"context"
	"crypto/ecdsa"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
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
	if err := c.BindJSON(&required); err != nil || required.Coin != "ETH" || required.Message != "chow_getaddress" || required.Key != "getaccess" {
		c.JSON(200, gin.H{
			"success": false,
			"message": "Access Denied",
		})
	} else {
		db, err := sql.Open("mysql", "sqluser@tcp(127.0.0.1:3306)/ethereum")
		if err != nil {
			c.JSON(200,gin.H{
				"success":false,
				"message":"DB Connection Failed",
			})
		} else {
			privateKey, err := crypto.GenerateKey()
			if err != nil {
				c.JSON(200,gin.H{
					"success":false,
					"message":"Address Generation Failed",
				})
			}
		}
	}
}