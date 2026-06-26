package models

type User struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	UserName string `json:"user_name"`
	Password string `json:"-"`
	Address string `json:"address"`
}
