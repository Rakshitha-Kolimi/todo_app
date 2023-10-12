package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"todo-project/constants"
)

var DB *sql.DB

func ConnectToDb() *sql.DB {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get database connection parameters from environment variables.
	dbUsername := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_DATABASE")
	dbPort := os.Getenv("DB_PORT")

	// Create the database connection string.
	dbConnectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUsername, dbPassword, dbName)

	// Initialize the database connection.
	DB, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Fatalf(constants.DATABASE_CONNECTION_ERROR, err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf(constants.DATABASE_CONNECTION_ERROR, err)
	}

	_, err = DB.Exec(UserTableQuery)
	if err != nil {
		log.Fatalf(constants.CANNOT_CREATE_TABLE_ERROR)
	}
	_, err = DB.Exec(CreateTableIfNotExistsQuery)
	if err != nil {
		log.Fatalf(constants.CANNOT_CREATE_TABLE_ERROR)
	}

	return DB
}
