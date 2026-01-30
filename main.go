package main

import (
	"context"
	"database/sql"
	"go-project-278/Internal/app"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://go_project_user:Fj2SbLdlar3a4l48bXHObp5r6ewZEzpO@dpg-d5u8jbh4tr6s739dbca0-a/go_project_db_h0do"
	}
	database, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	if os.Getenv("ENVIRONMENT") == "development" {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	} else {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	}
	r.Use(gin.Recovery())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	appCtx := context.Background()
	a := app.NewApp(appCtx, database)
	a.Routes(r)
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}	
	if os.Getenv("IN_DOCKER") == "true" {
		port = "8080"
	}
	log.Printf("Сервер запущен на порту %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}