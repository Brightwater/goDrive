package service

import (
	"context"
	"fmt"
	"log"
	"time"
	"github.com/jackc/pgx/v4/pgxpool"
)

var Pool *pgxpool.Pool

func InitPgPool() {

	config := AppConfig
	// err := godotenv.Load(".env")
	// if err != nil {
	// 	log.Fatalf("Error loading .env file %s", err)
	// }

	// host := os.Getenv("HOST")
	// port := os.Getenv("PORT")
	// user := os.Getenv("DBUSER")
	// password := os.Getenv("PASSWORD")
	// dbname := os.Getenv("DBNAME")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
			"password=%s dbname=%s sslmode=disable",
			config.Host, config.Port, config.DBUser, config.Password, config.DBName)

	poolConfig, err := pgxpool.ParseConfig(psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	poolConfig.MaxConns = 10
	poolConfig.MaxConnIdleTime = 15 * time.Minute
	poolConfig.MaxConnLifetime = 30 * time.Minute

	Pool, err = pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Database connected")
}

func CloseDb() {
	if Pool != nil {
		Pool.Close()
	}
}