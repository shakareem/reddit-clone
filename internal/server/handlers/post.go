package handlers

import (
	"encoding/json"
	"net/http"
	"redditclone/internal/storage"
)

type PostHandler struct {
	Storage storage.Storage
}

func NewPostHandler(storage storage.Storage) PostHandler {
	return PostHandler{
		Storage: storage,
	}
}

func (h *PostHandler) HandleNewPost(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")

	inToken := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		inToken = authHeader[7:]
	}

	claims, err := parseJWT(inToken)
	if err != nil {
		http.Error(w, "bad token", http.StatusUnauthorized)
		return
	}

	rawPost := &storage.RawPost{}
	err = json.NewDecoder(r.Body).Decode(rawPost)
	if err != nil {
		jsonError(w, http.StatusBadRequest, []RequestError{{
			Location: "post",
			Message:  err.Error(),
		}})
		return
	}

	post := h.Storage.AddPost(*rawPost, claims.User.Name, claims.User.ID)

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(&post)
	if err != nil {
		http.Error(w, "could not encode response", http.StatusInternalServerError)
	}
}
