package main

import (
	"github.com/ComputerSocietyVITC/recruitment-backend/types"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	router := gin.Default()

	// will update during deployment
	router.SetTrustedProxies([]string{"127.0.0.1"})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	router.POST("/hello", func(c *gin.Context) {
		var json types.ExampleJSONBody
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid json",
			})
			return
		}
		name := json.Name
		c.JSON(http.StatusOK, gin.H{
			"message": "hello " + name,
		})
	})

	router.Run()
}
