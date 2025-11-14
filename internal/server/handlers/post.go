package handlers

import (
	"encoding/json"
	"net/http"
	"redditclone/internal/storage"
	"sort"
	"time"
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

func (h *PostHandler) HandleGetPosts(w http.ResponseWriter, r *http.Request) {
	posts := h.Storage.GetPosts()

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Score < posts[j].Score
	})

	err := json.NewEncoder(w).Encode(&posts)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, []RequestError{{
			Location: "post",
			Message:  "Failed to encode posts",
		}})
	}
}

func (h *PostHandler) HandleGetCategoryPosts(w http.ResponseWriter, r *http.Request) {
	category := (r.PathValue("category"))

	posts := []storage.Post{}
	for _, p := range h.Storage.GetPosts() {
		if p.Category == category {
			posts = append(posts, p)
		}
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Score < posts[j].Score
	})

	err := json.NewEncoder(w).Encode(&posts)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, []RequestError{{
			Location: "post",
			Message:  "Failed to encode posts",
		}})
	}
}

func (h *PostHandler) HandleGetUserPosts(w http.ResponseWriter, r *http.Request) {
	username := (r.PathValue("username"))

	posts := []storage.Post{}
	for _, p := range h.Storage.GetPosts() {
		if p.Author.Name == username {
			posts = append(posts, p)
		}
	}

	sort.Slice(posts, func(i, j int) bool {
		iTime, _ := time.Parse(time.RFC3339, posts[i].CreatedTime)
		jTime, _ := time.Parse(time.RFC3339, posts[j].CreatedTime)
		return iTime.Before(jTime)
	})

	err := json.NewEncoder(w).Encode(&posts)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, []RequestError{{
			Location: "post",
			Message:  "Failed to encode posts",
		}})
	}
}
