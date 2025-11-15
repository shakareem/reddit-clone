package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"redditclone/internal/storage"
	"sort"
	"time"
)

type PostHandler struct {
	Storage storage.Storage
}

type key string

const USER key = "user"

func NewPostHandler(storage storage.Storage) PostHandler {
	return PostHandler{
		Storage: storage,
	}
}

func (h *PostHandler) handleNewPost(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(USER).(UserClaims)

	rawPost := &storage.RawPost{}
	err := json.NewDecoder(r.Body).Decode(rawPost)
	if err != nil {
		jsonError(w, http.StatusBadRequest, []RequestError{{
			Location: "post",
			Message:  err.Error(),
		}})
		return
	}

	post := h.Storage.AddPost(*rawPost, user.Name, user.ID)

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(&post)
	if err != nil {
		http.Error(w, "could not encode response", http.StatusInternalServerError)
	}
}

func (h *PostHandler) handleDeletePost(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(USER).(UserClaims)
	postID := r.PathValue("id")

	err := h.Storage.DeletePost(postID, user.ID)
	if err != nil {
		var statusCode int
		if errors.As(err, storage.ErrPostNotFound) {
			statusCode = http.StatusBadRequest
		} else if errors.As(err, storage.ErrPermissionDenied) {
			statusCode = http.StatusForbidden
		}

		http.Error(w, fmt.Sprintf("{\"message\":\"%s\"}", err.Error()), statusCode)
		return
	}

	w.Write([]byte(`{"message":"success"}`))
}

func (h *PostHandler) handleGetPosts(w http.ResponseWriter, r *http.Request) {
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

func (h *PostHandler) handleGetCategoryPosts(w http.ResponseWriter, r *http.Request) {
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

func (h *PostHandler) handleGetUserPosts(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

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

func (h *PostHandler) handleGetPostDetails(w http.ResponseWriter, r *http.Request) {
	postID := r.PathValue("id")

	post, err := h.Storage.GetPost(postID)
	if err != nil {
		http.Error(w, `{"message":"invalid post id"}`, http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(&post)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, []RequestError{{
			Location: "post",
			Message:  "Failed to encode post",
		}})
	}
}

func (h *PostHandler) handleUpvote(w http.ResponseWriter, r *http.Request) {
	handleVote(w, r, h.Storage.UpvotePost)
}
func (h *PostHandler) handleDownvote(w http.ResponseWriter, r *http.Request) {
	handleVote(w, r, h.Storage.DownvotePost)
}
func (h *PostHandler) handleUnvote(w http.ResponseWriter, r *http.Request) {
	handleVote(w, r, h.Storage.UnvotePost)
}

func handleVote(w http.ResponseWriter, r *http.Request, voteFunc func(id, userID string) (storage.Post, error)) {
	user := r.Context().Value(USER).(UserClaims)

	post, err := voteFunc(r.PathValue("id"), user.ID)
	if err != nil {
		http.Error(w, `{"message":"invalid post id"}`, http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(&post)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, []RequestError{{
			Location: "post",
			Message:  "Failed to encode post",
		}})
	}
}

type Comment struct {
	Comment string `json:"comment"`
}

func (h *PostHandler) handleAddComment(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(USER).(UserClaims)

	var comment struct {
		Message string `json:"comment"`
	}
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, `{"message":"invalid comment POST body"}`, http.StatusBadRequest)
		return
	}

	postID := r.PathValue("id")
	post, err := h.Storage.AddComment(postID, user.ID, user.Name, comment.Message)
	if err != nil {
		http.Error(w, `{"message":"invalid post id"}`, http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(&post)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, []RequestError{{
			Location: "post",
			Message:  "Failed to encode post",
		}})
	}
}

func (h *PostHandler) handleDeleteComment(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(USER).(UserClaims)
	postID, commentID := r.PathValue("postID"), r.PathValue("commentID")

	post, err := h.Storage.DeleteComment(postID, user.ID, commentID)
	if err != nil {
		var statusCode int
		if errors.As(err, storage.ErrPostNotFound) {
			statusCode = http.StatusBadRequest
		} else if errors.As(err, storage.ErrPermissionDenied) {
			statusCode = http.StatusForbidden
		}

		http.Error(w, fmt.Sprintf("{\"message\":\"%s\"}", err.Error()), statusCode)
		return
	}

	err = json.NewEncoder(w).Encode(&post)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, []RequestError{{
			Location: "post",
			Message:  "Failed to encode post",
		}})
	}
}
