package handlers

import (
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
	// TODO
}
