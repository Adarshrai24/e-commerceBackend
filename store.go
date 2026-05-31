package main

import (
	"time"
)

type Product struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
	Price float64 `json:"pice"`
	Stock int `json:"stock"`
}

type Order struct {
	ID int `json:"id"`
	TotalAmount float64 `json:"total_amount"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderItem struct {
	ID int `json:"id"`
	OrderID int `json:"order_id"`
	ProductID int `json:"product_id"`
	Quantity int `json:"quantity"`
	Price float64 `json:"price"`
}

//mock database
var mockProducts = []Product{
	{ID: 1, Name: "Mechanical Keyboard", Description: "Tactile switches", Price: 120.50, Stock: 10},
	{ID: 2, Name: "Monitor", Description: "4K IPS display", Price: 350.00, Stock: 5},
}

