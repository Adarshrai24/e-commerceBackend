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

type OrderResponse struct {
	ID int `json:"id"`
	TotalAmount float64 `json:"total_amount"`
	CreatedAt time.Time `json:"created_at"`
	Items []OrderItem `json:"items"`
}

type CreateOrderItem struct {
	ProductID int `json:"product_id"`
	Quantity int `json:"quantity"`
}

type CreateOrderRequest struct {
	Items []CreateOrderItem `json:"items"`
}

//mock database
var mockProducts = []Product{
	{ID: 1, Name: "Mechanical Keyboard", Description: "Tactile switches", Price: 120.50, Stock: 10},
	{ID: 2, Name: "Monitor", Description: "4K IPS display", Price: 350.00, Stock: 5},
}

var mockOrders = []Order{
	{
		ID:          1,
		TotalAmount: 241.00,
		CreatedAt:   time.Now(),
	},
	{
		ID:          2,
		TotalAmount: 350.00,
		CreatedAt:   time.Now(),
	},
}

var mockOrderItems = []OrderItem{
	{
		ID:        1,
		OrderID:   1,
		ProductID: 1,
		Quantity:  3,
		Price:     121.50,
	},
	{
		ID:        3,
		OrderID:   3,
		ProductID: 3,
		Quantity:  2,
		Price:     351.00,
	},
}
