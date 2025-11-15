package handlers

import (
	"net/http"
	"redditclone/internal/storage"
)

func ReqisterAPIHandlers(mux *http.ServeMux, storage storage.Storage) {
	userHandler := NewUserHandler(storage)
	postHandler := NewPostHandler(storage)

	apiMux := http.NewServeMux()
	apiMux.HandleFunc("POST /register", userHandler.handleRegister)
	apiMux.HandleFunc("POST /login", userHandler.handleLogIn)
	apiMux.HandleFunc("GET /posts/", postHandler.handleGetPosts)
	apiMux.HandleFunc("GET /posts/{category}", postHandler.handleGetCategoryPosts)
	apiMux.HandleFunc("GET /user/{username}", postHandler.handleGetUserPosts)
	apiMux.HandleFunc("GET /post/{id}", postHandler.handleGetPostDetails)
	apiMux.Handle("POST /posts", withAuth(http.HandlerFunc(postHandler.handleNewPost)))
	apiMux.Handle("DELETE /post/{id}", withAuth(http.HandlerFunc(postHandler.handleDeletePost)))
	apiMux.Handle("GET /post/{id}/upvote", withAuth(http.HandlerFunc(postHandler.handleUpvote)))
	apiMux.Handle("GET /post/{id}/downvote", withAuth(http.HandlerFunc(postHandler.handleDownvote)))
	apiMux.Handle("GET /post/{id}/unvote", withAuth(http.HandlerFunc(postHandler.handleUnvote)))
	apiMux.Handle("POST /post/{id}", withAuth(http.HandlerFunc(postHandler.handleAddComment)))
	apiMux.Handle("DELETE /post/{postID}/{commentID}", withAuth(http.HandlerFunc(postHandler.handleDeleteComment)))

	mux.Handle("/api/", http.StripPrefix("/api", apiMux))
}
