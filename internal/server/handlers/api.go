package handlers

import (
	"net/http"
	"redditclone/internal/storage"
)

func ReqisterAPIHandlers(mux *http.ServeMux, storage storage.Storage) {
	userHandler := NewUserHandler(storage)
	postHandler := NewPostHandler(storage)

	apiMux := http.NewServeMux()
	apiMux.HandleFunc("POST /register", userHandler.HandleRegister)
	apiMux.HandleFunc("POST /login", userHandler.HandleLogIn)
	apiMux.HandleFunc("GET /posts/", postHandler.HandleGetPosts)
	apiMux.HandleFunc("GET /posts/{category}", postHandler.HandleGetCategoryPosts)
	apiMux.HandleFunc("GET /user/{username}", postHandler.HandleGetUserPosts)
	apiMux.HandleFunc("GET /post/{id}", postHandler.HandleGetPostDetails)
	apiMux.Handle("POST /posts", withAuth(http.HandlerFunc(postHandler.HandleNewPost)))
	apiMux.Handle("DELETE /post/{id}", withAuth(http.HandlerFunc(postHandler.HandleDeletePost)))
	apiMux.Handle("GET /post/{id}/upvote", withAuth(http.HandlerFunc(postHandler.HandleUpvote)))
	apiMux.Handle("GET /post/{id}/downvote", withAuth(http.HandlerFunc(postHandler.HandleDownvote)))
	apiMux.Handle("GET /post/{id}/unvote", withAuth(http.HandlerFunc(postHandler.HandleUnvote)))
	apiMux.Handle("POST /post/{id}", withAuth(http.HandlerFunc(postHandler.handleAddComment)))
	apiMux.Handle("DELETE /post/{postID}/{commentID}", withAuth(http.HandlerFunc(postHandler.handleDeleteComment)))

	mux.Handle("/api/", http.StripPrefix("/api", apiMux))
}
