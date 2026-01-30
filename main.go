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
		log.Fatal("Failed to open DB:", err)
	}
	defer database.Close()	
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	log.Println("Database connected successfully")
	
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000", "https://go-project-278-14-xlxj.onrender.com"},
    AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "Accept", "Range"},
    ExposeHeaders:    []string{"Content-Length", "Content-Range"},
    AllowCredentials: true,
    MaxAge: 12 * time.Hour,
	}))
	r.Use(gin.Recovery())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"service": "go-project-278",
			"time":    time.Now().Format(time.RFC3339),
		})
	})
	appCtx := context.Background()
	a := app.NewApp(appCtx, database)
	a.Routes(r)
	port := "8080"
	log.Printf("Сервер запущен на порту %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}