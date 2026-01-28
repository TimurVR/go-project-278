package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	db, err := sql.Open("postgres", "postgresql://go_project_user:5crLwQD0QYVCjkppXQ5Dtjn2IPWvoBz5@dpg-d5svobu3jp1c738v4g40-a.oregon-postgres.render.com:5432/go_project_db?sslmode=require")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}
	if err := goose.Up(db, "db/migrations"); err != nil {
		log.Fatal(err)
	}
	log.Println("Миграции успешно применены!")
}