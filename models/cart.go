package models

type Cart struct {
	ID int `json: id`
	UserID int `json:"user_id"`
}

type CartItem struct {
	ID int `json: id`
	CartID int `json: cart_id`
	ProductID int `json: product_id`
	Quantity int `json: quantity`
}

type CreateCartItem struct {
	ProductID int `json:"product_id"`
	Quantity int `json:"quantity"`
}

type CreateCartRequest struct {
	Items []CreateCartItem `json:"items"`
}
