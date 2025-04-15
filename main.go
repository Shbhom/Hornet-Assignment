package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Product struct {
	ID    int     `json:id`
	name  string  `json:name`
	price float64 `json:price`
}

// global Variables (Clients)
var (
	db  *sql.DB
	rdb *redis.Client
	ctx = context.Background()
)

func initRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to redis", err)
	}
	log.Println("Connected To Redis")
}

func initPg() {
	connStr := os.Getenv("DB_URL")
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Connected to PostgreSQL")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	initPg()
	defer db.Close()

	initRedis()
	defer rdb.Close()

	router := gin.Default()

	log.Println("Server starting on :8080")
	if err := router.Run(":8000"); err != nil {
		log.Fatal("Failed to start server")
	}
}
