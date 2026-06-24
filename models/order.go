package models

import (
	"time"
)

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


