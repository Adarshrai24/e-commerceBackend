package handlers

import (
	"net/http"
	"encoding/json"
	"strconv"
	"database/sql"
	"ecommerce_api/models"
	"ecommerce_api/db"
)

func GetOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := db.DB.Query(`
		SELECT *
		FROM order_items
	`)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var order_items []models.OrderItem 

	for rows.Next() {
		var o models.OrderItem 
		err := rows.Scan(
			&o.ID, &o.OrderID, &o.ProductID, &o.Quantity, &o.Price,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		order_items = append(order_items, o)
	}

	err = json.NewEncoder(w).Encode(order_items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetOrderById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	row := db.DB.QueryRow(
		`SELECT * FROM order_items
		WHERE id = $1
		`,
		id, 
	)
	
	var o models.OrderItem
	err = row.Scan(
		&o.ID, &o.OrderID, &o.ProductID, &o.Quantity, &o.Price,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(o)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

