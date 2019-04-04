package main

import (
	"fmt"
	"context"
	"github.com/gin-gonic/gin"
)

type Required struct {
	Coin   string   `json:"coin" binding:"required"`
	Message   string `json:"message" binding:"required"`
	Key   string `json:"Key" binding:"required"`
}


func main() {

}

func CreateAddress(c *gin.Context) {

}