package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"redditclone/internal/storage"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("abc") // tmp

type UserHandler struct {
	Storage storage.Storage
}

func NewUserHandler(storage storage.Storage) *UserHandler {
	return &UserHandler{Storage: storage}
}

func (h *UserHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserName string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		Error(w, "wrong request body, username & password expected", http.StatusBadRequest)
		return
	}

	user, err := h.Storage.AddUser(req.UserName, req.Password)
	if err != nil {
		msg := fmt.Sprintf("could not add user: %v", err)
		Error(w, msg, http.StatusConflict) // TODO: понять что возвращать, если юзер уже есть
		return
	}

	token, err := generateJWT(user)
	if err != nil {
		msg := fmt.Sprintf("could not create token: %v", err)
		Error(w, msg, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(
		struct {
			Token string `json:"token"`
		}{token},
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func generateJWT(user storage.User) (string, error) {
	claims := jwt.MapClaims{
		"user": map[string]string{
			"id":       user.ID,
			"username": user.Name,
		},
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// category, text, title, type, как минимум айдишник
//
