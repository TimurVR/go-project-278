package main

import (
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
)
func main(){
 	r := gin.Default()
	r.Use(gin.Recovery())
 	r.GET("/ping", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
    })
  	if err := r.Run(":8080"); err != nil {
    	log.Fatalf("failed to run server: %v", err)
		panic("failed to run server")
  	}
}