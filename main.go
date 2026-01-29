package main

import (
	"context"
	"database/sql"
	"go-project-278/Internal/app"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	database, err := sql.Open("postgres", "postgresql://go_project_user:5crLwQD0QYVCjkppXQ5Dtjn2IPWvoBz5@dpg-d5svobu3jp1c738v4g40-a.oregon-postgres.render.com:5432/go_project_db?sslmode=require")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(gin.Recovery())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	appCtx := context.Background()
	a := app.NewApp(appCtx, database)
	a.Routes(r)
	port := ":3000"
	log.Printf("Сервер запущен на порту %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}