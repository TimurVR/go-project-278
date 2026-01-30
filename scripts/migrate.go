package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	db, err := sql.Open("postgres", "postgresql://go_project_user:Fj2SbLdlar3a4l48bXHObp5r6ewZEzpO@dpg-d5u8jbh4tr6s739dbca0-a/go_project_db_h0do")
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