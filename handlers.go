package main 

import (
	"net/http"
	"encoding/json"
	"strconv"
	"errors"
	"time"
	"database/sql"
)

func ValidateProduct(product Product) error {
	if product.Name == "" {
		return errors.New("name is required")
	}
	if product.Price < 0 {
		return errors.New("price cannot be negative")
	}
	if product.Stock < 0 {
		return errors.New("stock cannot be negative")
	}
	return nil
}

func ValidateOrderItem(item CreateOrderItem) error {
	if item.ProductID <= 0 {
		return errors.New("invalid product id")
	}

	if item.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	return nil
}

func ValidateCreateOrder(req CreateOrderRequest) error {
	if len(req.Items) == 0 {
		return errors.New("order must contain at least one item")
	}

	for _, item := range req.Items {
		if err := ValidateOrderItem(item); err != nil {
			return err
		}
	}

	return nil
}

func GetProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")	

	rows, err := db.Query(`
		SELECT id, name, description, price, stock 
		FROM products
	`)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()	

	var products []Product 
	for rows.Next() {
		var p Product 
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		products = append(products, p)
	}

	err = json.NewEncoder(w).Encode(products)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func GetProductById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	} 
	
	row := db.QueryRow(
		`SELECT * FROM products
		WHERE id = $1
		`,
		id, 
	)
	
	var p Product 
	err = row.Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func PostProduct(w http.ResponseWriter, r *http.Request) {
	var newProduct Product 
	err := json.NewDecoder(r.Body).Decode(&newProduct)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}	
	er := ValidateProduct(newProduct)
	if er != nil {
		http.Error(w, er.Error(), http.StatusBadRequest)
		return
	}

	err = db.QueryRow(
		`
		INSERT INTO products
		(name, description, price, stock)
		VALUES($1, $2, $3, $4)
		RETURNING id
		`,
		newProduct.Name,
		newProduct.Description,
		newProduct.Price,
		newProduct.Stock,
	).Scan(&newProduct.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(newProduct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	var updatedProduct Product 
	er := json.NewDecoder(r.Body).Decode(&updatedProduct)
	if er != nil {
		http.Error(w, "Invalid Input", http.StatusBadRequest)
		return
	}
	
	err = ValidateProduct(updatedProduct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec(
		`
		UPDATE products
		SET 
			name = $1,
			description = $2,
			price = $3,
			stock = $4
		WHERE id = $5
		`,
		updatedProduct.Name,
		updatedProduct.Description,
		updatedProduct.Price,
		updatedProduct.Stock,
		id,
	)	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}
	updatedProduct.ID = id
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(updatedProduct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}	
}

func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	
	result, err := db.Exec(
		`
		DELETE FROM products 
		WHERE id = $1
		`,
		id,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	rowsAffected, err := result.RowsAffected()
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := db.Query(`
		SELECT *
		FROM order_items
	`)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var order_items []OrderItem 

	for rows.Next() {
		var o OrderItem 
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

	row := db.QueryRow(
		`SELECT * FROM order_items
		WHERE id = $1
		`,
		id, 
	)
	
	var o OrderItem
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

func PostOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var orderItems CreateOrderRequest

	err := json.NewDecoder(r.Body).Decode(&orderItems)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = ValidateCreateOrder(orderItems)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	defer tx.Rollback()

	var totalAmount float64
	var quantities []int // to store quantity such that we don't have to call db to get this info
	var prices []float64 // to store prices 
	for _, item := range orderItems.Items {
		id := item.ProductID
		var p Product
		err = tx.QueryRow(
			`
			SELECT * FROM products
			WHERE id = $1;
			`,
			id,
		).Scan(
			&p.ID, 
			&p.Name,
			&p.Description,
			&p.Price,
			&p.Stock,
		)
			
		if err == sql.ErrNoRows {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if item.Quantity > p.Stock {
			http.Error(w, "Product out of stock", http.StatusBadRequest)
			return
		}
		quantities = append(quantities, p.Stock)	
		prices = append(prices, p.Price)
		totalAmount += p.Price * float64(item.Quantity)
	}

	for idx, item := range orderItems.Items {
		q := quantities[idx] - item.Quantity
		_, err := tx.Exec(
			`
			UPDATE products
			SET 
				stock = $1
			WHERE id = $2
			`,
			q,
			item.ProductID,
		)	
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}	
	}

	newOrder := Order{
		TotalAmount: totalAmount,
		CreatedAt:   time.Now(),
	}

	err = tx.QueryRow(
		`
		INSERT INTO orders
		(total_amount, created_at)
		VALUES($1, $2)
		RETURNING id
		`,
		newOrder.TotalAmount,
		newOrder.CreatedAt,
	).Scan(&newOrder.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for idx, item := range orderItems.Items {
		price := prices[idx]	
		order := OrderItem{
			OrderID: newOrder.ID,
			ProductID: item.ProductID,
			Quantity: item.Quantity,
			Price: price,
		}

			err = tx.QueryRow(
			`
			INSERT INTO order_items
			(order_id, product_id, quantity, price)
			VALUES($1, $2, $3, $4)
			RETURNING id
			`,
			order.OrderID,
			order.ProductID,
			order.Quantity,
			order.Price,
		).Scan(&order.ID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	
	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(newOrder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func UpdateOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orderId, err := strconv.Atoi(r.PathValue("orderId"))
	if err != nil {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	productId, err := strconv.Atoi(r.PathValue("productId"))
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}

	var updatedOrderItem CreateOrderItem

	err = json.NewDecoder(r.Body).Decode(&updatedOrderItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if updatedOrderItem.Quantity <= 0 {
		http.Error(w, "quantity must be greater than 0", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin();
	if (err != nil) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer tx.Rollback()
	var currentQuantity int 
	var currentStock int 
	err = tx.QueryRow(
		`
		SELECT quantity FROM order_items
		WHERE order_id = $1 and product_id = $2
		`,
		orderId,
		productId,
	).Scan(
		&currentQuantity,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "No such order found", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	err = tx.QueryRow(
		`
		SELECT stock from products
		WHERE id = $1
		`,
		productId,
	).Scan(
		&currentStock, 
	)

	if err == sql.ErrNoRows {
		http.Error(w, "No such product found", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	requiredQuantity := updatedOrderItem.Quantity - currentQuantity
	if (requiredQuantity > currentStock) {
		http.Error(w, "Item out of stock", http.StatusBadRequest)
		return
	}
	
	_, err = tx.Exec(
		`
		UPDATE order_items
		SET
			quantity = $1
		WHERE order_id = $2 and product_id = $3
		`,
		updatedOrderItem.Quantity,
		orderId,
		productId,
	)	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var totalAmount float64

	err = tx.QueryRow(
		`
		SELECT COALESCE(SUM(quantity * price), 0)
		FROM order_items
		WHERE order_id = $1
		`,
		orderId,
	).Scan(&totalAmount)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = tx.Exec(
		`
		UPDATE products
		SET	
			stock = $1
		WHERE id = $2
		`,
		currentStock - requiredQuantity,
		productId,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}	
	
	_, err = tx.Exec(
		`
		UPDATE orders
		SET 
			total_amount = $1
		WHERE id = $2
		`,
		totalAmount,
		orderId,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func DeleteOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orderID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}	
	
	tx, err := db.Begin()

	if (err != nil) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()	

	var updatedOrderItems []OrderItem
	
	rows, err := tx.Query(
		`
		SELECT * FROM order_items
		WHERE order_id = $1
		`,
		orderID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item OrderItem
		err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		updatedOrderItems = append(updatedOrderItems, item)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(updatedOrderItems) == 0 {
		http.Error(w, "No such orders found", http.StatusBadRequest)
		return
	}

	for _, item := range updatedOrderItems {
		var currentStock int 
		err = tx.QueryRow(
			`
			SELECT stock FROM products 
			WHERE id = $1
			`,
			item.ProductID,
		).Scan(
			&currentStock,
		)
		if err == sql.ErrNoRows {
			http.Error(w, "No such product found", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		updatedStock := currentStock + item.Quantity

		_, err = tx.Exec(
			`
			UPDATE products 
			SET 
				stock = $1
			WHERE id = $2
			`,
			updatedStock,
			item.ProductID,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	
	_, err = tx.Exec(
		`
		DELETE FROM order_items 
		WHERE order_id = $1
		`,
		orderID,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(
		`
		DELETE FROM orders 
		WHERE id = $1
		`,
		orderID, 
	)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
