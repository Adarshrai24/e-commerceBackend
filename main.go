package main 

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	
	mux.HandleFunc("GET /products", GetProduct)
	mux.HandleFunc("GET /products/{id}", GetProductById)
	mux.HandleFunc("POST /products", PostProduct)
	mux.HandleFunc("PUT /products/{id}", UpdateProduct)
	mux.HandleFunc("DELETE /products/{id}", DeleteProduct)

	log.Println("Server running on :8080")

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
