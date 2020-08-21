package config

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// DB of type sql.DB
var DB *sql.DB

//ConnectMySQL to bot db
func ConnectMySQL() *sql.DB {

	dbDriver := "mysql"
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbAddress := os.Getenv("DB_ADDRESS")
	dbPort := os.Getenv("DB_PORT")
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp("+dbAddress+":"+dbPort+")/"+dbName)
	// Handling if the db object is created or no
	if err != nil {
		log.Panicf("Unable to Open the database %s", err.Error())
	}
	err = db.Ping()
	// handling if the communication is available to the database or no
	if err != nil {
		log.Panicf("Connection to database failed %s", err.Error())
	}
	log.Println("Connected to MySQL Successfully ...")
	return db
}
