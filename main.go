package main

import (
	"context"
	"database/sql"
	"go-project-278/Internal/app"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://go_project_user:5crLwQD0QYVCjkppXQ5Dtjn2IPWvoBz5@dpg-d5svobu3jp1c738v4g40-a/go_project_db"
	}
	
	if !strings.Contains(databaseURL, "sslmode") {
		databaseURL += "?sslmode=require"
	}
	
	database, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	r := gin.Default()
	
	corsConfig := cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"https://go-project-278-24.onrender.com",
			"http://go-project-278-24.onrender.com",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH", "HEAD"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-API-Key",
		},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "production"
	}
	
	if env == "development" {
		corsConfig.AllowOrigins = append(corsConfig.AllowOrigins, 
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		)
	}
	
	r.Use(cors.New(corsConfig))
	r.Use(gin.Recovery())
	
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	
	r.GET("/health", func(c *gin.Context) {
		if err := database.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "go-project-278",
		})
	})
	
	appCtx := context.Background()
	a := app.NewApp(appCtx, database)
	a.Routes(r)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}
	
	log.Printf("Сервер запущен на порту %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}