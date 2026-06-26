package handlers

import (
	"net/http"
	"encoding/json"
	"strconv"
	"database/sql"
	"ecommerce_api/models"
)

func GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var users []models.User 

	rows, err := db.DB.Query(
		`
		SELECT * FROM users 
		`,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var u models.User 
		err = rows.Scan(&u.ID, &u.Name, &u.Email, &u.UserName, &u.Address)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	if len(users) == 0 {
		http.Error(w, "No users present currently", http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetUserById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, err := strconv.Atoi(r.PathValue("userID"))
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	var users models.User

	err := db.DB.QueryRow(
		`
		SELECT * FROM users 
		WHERE id = $1
		`,
		userID,
	).Scan(
		&users.ID, &users.Name, &users.Email, &users.UserName, &users.Address
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
