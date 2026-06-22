package main 

import (
	"log"
	"fmt"
	"os"
	"net/http"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

//var db *sql.DB

func main() {
	
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	
	user := os.Getenv("DB_USER")
	host := os.Getenv("DB_HOST")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	connStr := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=disable",
		host, user, password, name,
	)
	db, err := sql.Open("postgres", connStr)
	
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	
	mux.HandleFunc("GET /products", GetProduct)
	mux.HandleFunc("GET /products/{id}", GetProductById)
	mux.HandleFunc("POST /products", PostProduct)
	mux.HandleFunc("PUT /products/{id}", UpdateProduct)
	mux.HandleFunc("DELETE /products/{id}", DeleteProduct)
	mux.HandleFunc("GET /orders", GetOrder)
	mux.HandleFunc("GET /orders/{id}", GetOrderById)
	mux.HandleFunc("POST /orders", PostOrder)
	mux.HandleFunc("PATCH /orders/{orderId}/products/{productId}", UpdateOrder)	
	mux.HandleFunc("DELETE /orders/{id}", DeleteOrder)
	log.Println("Server running on :8080")

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}	
}
