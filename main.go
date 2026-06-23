package main 

import (
	"log"
	"fmt"
	"os"
	"net/http"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"ecommerce_api/handlers"
	"ecommerce_api/db"
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
	conn, err := sql.Open("postgres", connStr)
	db.DB = conn	
	if err != nil {
		panic(err)
	}

	err = conn.Ping()
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	
	mux.HandleFunc("GET /products", handlers.GetProduct)
	mux.HandleFunc("GET /products/{id}", handlers.GetProductById)
	mux.HandleFunc("POST /products", handlers.PostProduct)
	mux.HandleFunc("PUT /products/{id}", handlers.UpdateProduct)
	mux.HandleFunc("DELETE /products/{id}", handlers.DeleteProduct)
	mux.HandleFunc("GET /orders", handlers.GetOrder)
	mux.HandleFunc("GET /orders/{id}", handlers.GetOrderById)
	mux.HandleFunc("POST /orders", handlers.PostOrder)
	mux.HandleFunc("PATCH /orders/{orderId}/products/{productId}", handlers.UpdateOrder)	
	mux.HandleFunc("DELETE /orders/{id}", handlers.DeleteOrder)
	mux.HandleFunc("GET /cart", handlers.GetCart)
	mux.HandleFunc("GET /cart/{idx}", handlers.GetCartByID)
	mux.HandleFunc("POST /cart/", handlers.PostCart)
	log.Println("Server running on :8080")

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}	
}
