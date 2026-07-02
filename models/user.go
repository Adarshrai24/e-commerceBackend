package models

type User struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"-"`
	Address string `json:"address"`
}

type UserRequest struct {	
	ID int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
	Address string `json:"address"`
}

type UserResponse struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	Address string `json:"address"`
}

type UpdateUserRequest struct {
	Name string `json:"name"`
	Password string `json:"password"`
	Address string `json:"address"`	
}
