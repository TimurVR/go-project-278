package main

import (
	"context"
	"database/sql"
	"go-project-278/Internal/app"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main(){
 	r := gin.Default()
	ctx := context.Background()
	database, err := sql.Open("postgres", "postgresql://go_project_user:5crLwQD0QYVCjkppXQ5Dtjn2IPWvoBz5@dpg-d5svobu3jp1c738v4g40-a.oregon-postgres.render.com:5432/go_project_db?sslmode=require")
	a := app.NewApp(ctx, database)
	a.Routes(r)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		log.Fatalf("Не удалось подключиться к БД %w:", err)
	}
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
