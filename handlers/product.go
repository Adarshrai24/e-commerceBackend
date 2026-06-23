package handlers

import (
	"net/http"
	"encoding/json"
	"strconv"
	"errors"
	"database/sql"
	"ecommerce_api/models"
	"ecommerce_api/db"
)

func ValidateProduct(product models.Product) error {
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

func GetProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")	

	rows, err := db.DB.Query(`
		SELECT id, name, description, price, stock 
		FROM products
	`)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()	

	var products []models.Product 
	for rows.Next() {
		var p models.Product 
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
	
	row := db.DB.QueryRow(
		`SELECT * FROM products
		WHERE id = $1
		`,
		id, 
	)
	
	var p models.Product 
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
	var newProduct models.Product 
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

	err = db.DB.QueryRow(
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

	var updatedProduct models.Product 
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

	result, err := db.DB.Exec(
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
	
	result, err := db.DB.Exec(
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
