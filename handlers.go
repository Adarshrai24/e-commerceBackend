package main 

import (
	"net/http"
	"encoding/json"
	"strconv"
	"errors"
	"time"
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

	err := json.NewEncoder(w).Encode(mockProducts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetProductById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(r.PathValue("id"))
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
	id, err := strconv.Atoi(r.PathValue("id"))
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

func GetOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(mockOrders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetOrderById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idString := r.PathValue("id")

	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	for _, order := range mockOrders {
		if order.ID != id {
			continue
		}

		var items []OrderItem

		for _, item := range mockOrderItems {
			if item.OrderID == id {
				items = append(items, item)
			}
		}

		response := OrderResponse{
			ID:          order.ID,
			TotalAmount: order.TotalAmount,
			CreatedAt:   order.CreatedAt,
			Items:       items,
		}

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	http.Error(w, "order not found", http.StatusNotFound)
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

	var totalAmount float64

	for _, item := range orderItems.Items {
		productFound := false

		for _, product := range mockProducts {
			if item.ProductID == product.ID {
				productFound = true

				if item.Quantity > product.Stock {
					http.Error(w, "product out of stock", http.StatusBadRequest)
					return
				}

				totalAmount += product.Price * float64(item.Quantity)
				break
			}
		}

		if !productFound {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}
	}

	for _, item := range orderItems.Items {
		for idx := range mockProducts {
			if mockProducts[idx].ID == item.ProductID {
				mockProducts[idx].Stock -= item.Quantity
				break
			}
		}
	}

	newOrderID := len(mockOrders) + 1

	newOrder := Order{
		ID:          newOrderID,
		TotalAmount: totalAmount,
		CreatedAt:   time.Now(),
	}

	mockOrders = append(mockOrders, newOrder)

	for _, item := range orderItems.Items {
		var price float64

		for _, product := range mockProducts {
			if product.ID == item.ProductID {
				price = product.Price
				break
			}
		}

		mockOrderItems = append(mockOrderItems, OrderItem{
			ID:        len(mockOrderItems) + 1,
			OrderID:   newOrderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     price,
		})
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

	itemFound := false

	for idx := range mockOrderItems {
		item := &mockOrderItems[idx]

		if item.OrderID == orderId && item.ProductID == productId {
			itemFound = true

			currentQuantity := item.Quantity
			updatedQuantity := updatedOrderItem.Quantity
			requiredQuantity := updatedQuantity - currentQuantity

			productFound := false

			for j := range mockProducts {
				product := &mockProducts[j]

				if product.ID == productId {
					productFound = true

					if requiredQuantity > product.Stock {
						http.Error(w, "product out of stock", http.StatusBadRequest)
						return
					}

					product.Stock -= requiredQuantity
					break
				}
			}

			if !productFound {
				http.Error(w, "product not found", http.StatusNotFound)
				return
			}

			item.Quantity = updatedQuantity
			break
		}
	}

	if !itemFound {
		http.Error(w, "order item not found", http.StatusNotFound)
		return
	}

	var totalAmount float64

	for _, item := range mockOrderItems {
		if item.OrderID == orderId {
			totalAmount += item.Price * float64(item.Quantity)
		}
	}

	for idx := range mockOrders {
		if mockOrders[idx].ID == orderId {
			mockOrders[idx].TotalAmount = totalAmount

			w.WriteHeader(http.StatusOK)

			err = json.NewEncoder(w).Encode(mockOrders[idx])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return
		}
	}

	http.Error(w, "order not found", http.StatusNotFound)
}

func DeleteOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orderID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	orderFound := false

	for _, order := range mockOrders {
		if order.ID == orderID {
			orderFound = true
			break
		}
	}

	if !orderFound {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	for _, item := range mockOrderItems {
		if item.OrderID == orderID {
			for idx := range mockProducts {
				if mockProducts[idx].ID == item.ProductID {
					mockProducts[idx].Stock += item.Quantity
					break
				}
			}
		}
	}

	var updatedOrderItems []OrderItem

	for _, item := range mockOrderItems {
		if item.OrderID != orderID {
			updatedOrderItems = append(updatedOrderItems, item)
		}
	}

	mockOrderItems = updatedOrderItems

	var updatedOrders []Order

	for _, order := range mockOrders {
		if order.ID != orderID {
			updatedOrders = append(updatedOrders, order)
		}
	}

	mockOrders = updatedOrders

	w.WriteHeader(http.StatusNoContent)
}
