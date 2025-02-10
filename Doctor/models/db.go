package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB

func InitDb() {

	// dbUser := "root"
	// dbPass := "123456789"
	// dbHost := "localhost"
	// dbPort := 3306
	// dbName := "doctor_service"

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v", dbUser, dbPass, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error opening database:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	Db = db
	fmt.Println("Connected With Database")
}
