package main 

import (
	"net/http"
	"encoding/json"
	"strconv"
	"errors"
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

func GetProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")	

	err := json.NewEncoder(w).Encode(mockProducts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetProductById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	} else {
		for _, i := range mockProducts {
			if i.ID == id {
				er := json.NewEncoder(w).Encode(i)
				if er != nil {
					http.Error(w, er.Error(), http.StatusInternalServerError)
					return
				}
				return
			}
		}
		http.Error(w, "Product not found", http.StatusNotFound)
	}
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
	newProduct.ID = len(mockProducts) + 1
	mockProducts = append(mockProducts, newProduct)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newProduct)
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)

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

	for idx, i := range mockProducts {
		if i.ID == id {
			updatedProduct.ID = id
			mockProducts[idx] = updatedProduct
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(updatedProduct)
			return
		}
	}
	http.Error(w, "ID not found", http.StatusNotFound)
}

func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	for idx, i := range mockProducts {
		if i.ID == id {
			var afterDeleteProduct []Product 
			afterDeleteProduct = append(afterDeleteProduct, mockProducts[:idx]...)
			if idx+1 < len(mockProducts) {
				afterDeleteProduct = append(afterDeleteProduct, mockProducts[idx+1:]...)
			}
			mockProducts = afterDeleteProduct
			w.WriteHeader(http.StatusNoContent)
			json.NewEncoder(w).Encode(mockProducts)
			return
		}
	}
	http.Error(w, "ID not found", http.StatusNotFound)
}
