package handlers

import(
	"net/http"
	"encoding/json"
	"database/sql"
	"strconv"
	"errors"
	"time"
	"ecommerce_api/models"
	"ecommerce_api/db"
)

func ValidateCartItem(item models.CreateCartItem) error {
	if item.ProductID <= 0 {
		return errors.New("invalid product id")
	}

	if item.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	return nil
}

func ValidateCreateCart(req models.CreateCartRequest) error {
	if len(req.Items) == 0 {
		return errors.New("cart must contain at least one item")
	}

	for _, item := range req.Items {
		if err := ValidateCartItem(item); err != nil {
			return err
		}
	}

	return nil
}

func GetCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var cart []models.Cart
	
	rows, err := db.DB.Query(
		`
		SELECT * FROM carts
		`,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var c models.Cart
		err = rows.Scan(
			&c.ID, &c.UserID,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cart = append(cart, c)
	}
	
	if (len(cart) == 0) {
		http.Error(w, "Cart is empty", http.StatusBadRequest)
		return
	}
	err = json.NewEncoder(w).Encode(cart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetCartByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cartId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid cart id", http.StatusBadRequest)
		return
	}

	var cartItems []models.CartItem 
	
	rows, err := db.DB.Query(
		`
		SELECT * FROM cart_items 
		WHERE cart_id = $1
		`,
		cartId,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var c models.CartItem
		err = rows.Scan(
			&c.ID, &c.CartID, &c.ProductID, &c.Quantity,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cartItems = append(cartItems, c)
	}
	if (len(cartItems) == 0) {
		http.Error(w, "no cart items present for this cart id", http.StatusBadRequest)
		return
	}
	err = json.NewEncoder(w).Encode(cartItems)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func PostCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var cartItems models.CreateCartRequest
	err := json.NewDecoder(r.Body).Decode(&cartItems)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = ValidateCreateCart(cartItems)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, item := range cartItems.Items {
		var p models.Product 
		err = tx.QueryRow(
			`
			SELECT * FROM products 
			WHERE id = $1
			`,
			item.ProductID,
		).Scan(
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

		if item.Quantity > p.Stock {
			http.Error(w, "Product out of stock", http.StatusBadRequest)
			return
		}
	}

	var cartID int 
	err = tx.QueryRow(
		`
		INSERT INTO carts
		DEFAULT VALUES
		RETURNING id
		`,
	).Scan(&cartID)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, item := range cartItems.Items {
		c := models.CartItem{
			CartID: cartID,
			ProductID: item.ProductID,
			Quantity: item.Quantity,	
		}
		err = tx.QueryRow(
			`
			INSERT INTO cart_items 
			(cart_id, product_id, quantity)
			VALUES($1, $2, $3)
			RETURNING id
			`,
			c.CartID, c.ProductID, c.Quantity,
		).Scan(&c.ID)

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
	response := models.Cart{
		ID: cartID,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func UpdateCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cartID, err := strconv.Atoi(r.PathValue("cartID"))
	if err != nil {
		http.Error(w, "invalid cart id", http.StatusBadRequest)
		return
	}

	productID, err := strconv.Atoi(r.PathValue("productID"))
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}

	var updateCartItem models.CreateCartItem 
	err = json.NewDecoder(r.Body).Decode(&updateCartItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if updateCartItem.Quantity <= 0 {
		http.Error(w, "Quantity of product can't be <= 0", http.StatusBadRequest)
		return
	}

	var quantity int 
	err = db.DB.QueryRow(
		`
		SELECT stock FROM products 
		WHERE id = $1
		`,
		productID,
	).Scan(&quantity)

	if err == sql.ErrNoRows {
		http.Error(w, "No product found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if quantity < updateCartItem.Quantity {
		http.Error(w, "required quantity is greater than stock present", http.StatusBadRequest)
		return
	}

	_, err = db.DB.Exec(
		`
		UPDATE cart_items
		SET 
			quantity = $1
		WHERE cart_id = $2 and product_id = $3
		`,
		quantity,
		cartID,
		productID,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cartID, err := strconv.Atoi(r.PathValue("cartID"))

	if err != nil {
		http.Error(w, "invalid cart id", http.StatusBadRequest)
		return
	}

	result, err := db.DB.Exec(
		`
		DELETE FROM carts
		WHERE id = $1
		`,
		cartID,
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
		http.Error(w, "cart not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteProductFromCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cartID, err := strconv.Atoi(r.PathValue("cartID"))
	if err != nil {
		http.Error(w, "invalid cart id", http.StatusBadRequest)
		return
	}

	productID, err := strconv.Atoi(r.PathValue("productID"))
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}

	result, err := db.DB.Exec(
		`
		DELETE FROM cart_items
		WHERE cart_id = $1 and product_id = $2
		`,
		cartID,
		productID,
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
		http.Error(w, "Product not found in cart", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func CheckoutFromCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cartID, err := strconv.Atoi(r.PathValue("cartID"))

	if err != nil {
		http.Error(w, "invalid cart id", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	
	var cartItems []models.CartItem
	rows, err := tx.Query(
		`
		SELECT * FROM cart_items 
		WHERE cart_id = $1
		`,
		cartID,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var c models.CartItem
		err = rows.Scan(
			&c.ID, &c.CartID, &c.ProductID, &c.Quantity,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cartItems = append(cartItems, c)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}	
	if (len(cartItems) == 0) {
		http.Error(w, "Cart is empty", http.StatusBadRequest)
		return
	}		
	var totalAmount float64	
	for _, item := range cartItems {
		pid := item.ProductID
		quantity := item.Quantity
		var price float64
		var stock int
		//FOR UPDATE in query prevents race condition
		err = tx.QueryRow(
			`
			SELECT price, stock FROM products 
			WHERE id = $1
			FOR UPDATE 
			`,
			pid,
		).Scan(
			&price, &stock,
		)
		if err == sql.ErrNoRows {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if stock < quantity {
			http.Error(w, "product is out of stock", http.StatusBadRequest)
			return
		}
		totalAmount += float64(quantity) * price
			_, err = tx.Exec(
			`
			UPDATE products
			SET	
				stock = $1
			WHERE id = $2
			`,
			stock-quantity,
			pid,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	
	_, err = tx.Exec(
		`
		DELETE FROM carts 
		WHERE id = $1
		`,
		cartID,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// from here we are creating new order
	newOrder := models.Order{
		TotalAmount: totalAmount,
		CreatedAt: time.Now(),
	}

	err = tx.QueryRow(
		`
		INSERT INTO orders 
		(total_amount, created_at)
		VALUES($1, $2)
		RETURNING id
		`,
		newOrder.TotalAmount, newOrder.CreatedAt,
	).Scan(&newOrder.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, item := range cartItems {	
		pid := item.ProductID
		var price float64
		err = tx.QueryRow(
			`
			SELECT price FROM products 
			WHERE id = $1
			`,
			pid,
		).Scan(&price)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		orderItems := models.OrderItem{
			OrderID: newOrder.ID,
			ProductID: pid,
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
			newOrder.ID, pid, item.Quantity, price,
		).Scan(&orderItems.ID)
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
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&newOrder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}	
}
