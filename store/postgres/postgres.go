package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"
	"wdiet/store"

	_ "github.com/lib/pq"
)

var _ store.Store = (*PG)(nil)

type PG struct { //PG struct implements the Store sinterface. Store is gonna keep all your db function definitions, and you're going to implement them in postgres.
	db *sql.DB
}

func New() (*PG, error) {
	dbString, err := getDBStringFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("error creating new postgres store: %w", err)
	}

	db, err := sql.Open("postgres", dbString)
	if err != nil {
		return nil, fmt.Errorf("error creating new postgres store: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return &PG{
		db: db,
	}, nil
}

func getDBStringFromEnvironment() (string, error) {
	dbHost, exists := os.LookupEnv("WDIET_DB_HOST")
	if !exists {
		return "", fmt.Errorf("error getting WDIET_DB_HOST from environment")
	}

	dbPort, exists := os.LookupEnv("WDIET_DB_PORT")
	if !exists {
		return "", fmt.Errorf("error getting WDIET_DB_PORT from environment")
	}

	dbUser, exists := os.LookupEnv("WDIET_DB_USER")
	if !exists {
		return "", fmt.Errorf("error getting WDIET_DB_USER from environment")
	}

	dbPass, exists := os.LookupEnv("WDIET_DB_PASS")
	if !exists {
		return "", fmt.Errorf("error getting WDIET_DB_PASS from environment")
	}

	dbName, exists := os.LookupEnv("WDIET_DB_NAME")
	if !exists {
		return "", fmt.Errorf("error getting WDIET_DB_NAME from environment")
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		dbHost, dbPort, dbUser, dbPass, dbName), nil
}
