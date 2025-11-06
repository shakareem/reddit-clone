package handlers

import (
	"encoding/json"
	"errors"
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

type LogInRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

func NewUserHandler(storage storage.Storage) *UserHandler {
	return &UserHandler{Storage: storage}
}

func (h *UserHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req LogInRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		jsonError(w, http.StatusBadRequest, []RequestError{{
			Location: "body",
			Message:  "wrong request body, username & password expected",
		}})
		return
	}

	user, err := h.Storage.AddUser(req.UserName, req.Password)
	if errors.Is(err, storage.ErrUserAlreadyExists) {
		jsonError(w, http.StatusUnprocessableEntity, []RequestError{{
			Location: "body",
			Param:    "username",
			Value:    req.UserName,
			Message:  "already exists",
		}})
		return
	}

	token, err := generateJWT(user)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, []RequestError{{
			Message: fmt.Sprintf("could not create token: %v", err),
		}})
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

func (h *UserHandler) HandleLogIn(w http.ResponseWriter, r *http.Request) {
	var req LogInRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		jsonError(w, http.StatusBadRequest, []RequestError{{
			Location: "body",
			Message:  "wrong request body, username & password expected",
		}})
		return
	}

	user, err := h.Storage.GetUser(req.UserName, req.Password)
	if err != nil {
		msg := fmt.Sprintf("{\"message\":\"%s\"}", err.Error())
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(user)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, []RequestError{{
			Message: fmt.Sprintf("could not create token: %v", err),
		}})
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
