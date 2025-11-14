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
	apiMux.Handle("POST /posts", WithAuth(http.HandlerFunc(postHandler.HandleNewPost)))
	apiMux.Handle("DELETE /post/{id}", WithAuth(http.HandlerFunc(postHandler.HandleDeletePost)))
	apiMux.Handle("GET /post/{id}/upvote", WithAuth(http.HandlerFunc(postHandler.HandleUpvote)))
	apiMux.Handle("GET /post/{id}/downvote", WithAuth(http.HandlerFunc(postHandler.HandleDownvote)))
	apiMux.Handle("GET /post/{id}/unvote", WithAuth(http.HandlerFunc(postHandler.HandleUnvote)))

	mux.Handle("/api/", http.StripPrefix("/api", apiMux))
}
