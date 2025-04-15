package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type createProductRequest struct {
	Name  string  `json:"name" binding:"required"`
	Price float64 `json:"Price" binding:"required,gt=0"`
}

type updateProductRequest struct {
	Name  string  `json:"name" binding:"required"`
	Price float64 `json:"Price"`
}

// global Variables (Clients)
var (
	db  *sql.DB
	rdb *redis.Client
	ctx = context.Background()
)

func initRedis() {
	redis_addr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")

	redis_db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatal("Unable to get DB value for redis")
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     redis_addr,
		Password: os.Getenv("REDIS_PASS"),
		DB:       redis_db,
	})

	_, err = rdb.Ping(ctx).Result()
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

	router.GET("/products", getAllProducts)
	router.POST("/products", createProduct)
	router.GET("/products/:id", getProductById)
	router.PUT("/products/:id", updateProduct)
	router.DELETE("/products/:id", deleteProduct)

	log.Println("Server starting on :8080")
	if err := router.Run(":8000"); err != nil {
		log.Fatal("Failed to start server")
	}
}

func getAllProducts(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM products")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to Fetch Products"})
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan products"})
			return
		}
		// fmt.Println(product)
		products = append(products, product)
	}
	c.JSON(http.StatusOK, products)
}

func getProductById(c *gin.Context) {
	id := c.Param("id")

	// fetch from redis
	cacheKey := fmt.Sprintf("product:%s", id)
	cachedProduct, err := rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var product Product
		if err := json.Unmarshal([]byte(cachedProduct), &product); err == nil {
			log.Printf("Cache Hit for product id: %s", id)
			c.JSON(http.StatusOK, product)
			return
		}
		log.Printf("Cache HIT but failed to unmarshal product ID: %s - %v", id, err)
	} else {
		if err == redis.Nil {
			log.Printf("Cache Miss for product Id: %s", id)
		} else {
			log.Printf("Redis ERROR for product ID: %s - %v", id, err)
		}
	}

	// if no cache query DB
	var product Product
	err = db.QueryRow(`SELECT * FROM products WHERE id=$1`, id).Scan(&product.ID, &product.Name, &product.Price)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// updating cache
	if productJSON, err := json.Marshal(product); err == nil {
		if err := rdb.Set(ctx, cacheKey, productJSON, time.Minute).Err(); err != nil {
			log.Printf("Failed to CACHE product ID: %s - %v", id, err)
		}
	} else {
		log.Printf("Failed to marshal Product for caching: %v", err)
	}

	log.Printf("Data retrieved from DB for product ID: %s", id)
	c.JSON(http.StatusOK, product)
}

func createProduct(c *gin.Context) {
	var req createProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	var product Product

	err := db.QueryRow(
		`INSERT INTO products (name,price) values ($1,$2) RETURNING id,name,price`,
		req.Name, req.Price).Scan(&product.ID, &product.Name, &product.Price)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, product)
}

func updateProduct(c *gin.Context) {
	id := c.Param("id")
	var req updateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var product Product

	err := db.QueryRow("UPDATE products SET name=$1, price=$2 WHERE id=$3 RETURNING id,name,price", req.Name, req.Price, id).Scan(&product.ID, &product.Name, &product.Price)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		}
		return
	}

	rdb.Del(ctx, "product:"+id)
	c.JSON(http.StatusOK, product)
}

func deleteProduct(c *gin.Context) {
	id := c.Param("id")

	result, err := db.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	rdb.Del(ctx, "product:"+id)
	c.Status(http.StatusNoContent)
}
