package handlers

import(
	"net/http"
	"encoding/json"
	"database/sql"
	"strconv"
	"errors"
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
	
	var cartItems []models.CartItem 
	
	rows, err := db.DB.Query(
		`
		SELECT * FROM order_items 
		`,
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
		http.Error(w, "Cart is empty", http.StatusBadRequest)
		return
	}
	err = json.NewEncoder(w).Encode(cartItems)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var cartItems []models.CartItem 
	
	rows, err := db.DB.Query(
		`
		SELECT * FROM order_items 
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
