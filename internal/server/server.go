package server

import (
	"log"
	"net/http"
	"redditclone/internal/server/handlers"
	"redditclone/internal/storage"
	"time"
)

type Service struct {
	Server  *http.Server
	Storage storage.Storage
}

const PORT = ":8081"

func NewService() Service {
	storage := storage.NewInMemStorage()
	userHandler := handlers.NewUserHandler(storage)
	postHandler := handlers.NewPostHandler(storage)

	mux := http.NewServeMux()
	registerStaticHandlers(mux)
	reqisterAPIHandlers(mux, userHandler, &postHandler)

	log.Println("Starting server on :8081")
	server := &http.Server{
		Addr:         PORT,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return Service{
		Server:  server,
		Storage: storage,
	}
}

func (s *Service) Run() error {
	return s.Server.ListenAndServe()
}

func registerStaticHandlers(mux *http.ServeMux) {
	staticHTMLHandler := http.FileServer(http.Dir("./web/html"))
	mux.Handle("/", staticHTMLHandler)

	staticHandler := http.FileServer(http.Dir("./web"))
	mux.Handle("/static/", http.StripPrefix("/static/", staticHandler))
}

func reqisterAPIHandlers(mux *http.ServeMux, userHandler *handlers.UserHandler, postHandler *handlers.PostHandler) {
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("POST /register", userHandler.HandleRegister)
	apiMux.HandleFunc("POST /login", userHandler.HandleLogIn)
	apiMux.Handle("POST /posts", handlers.WithAuth(http.HandlerFunc(postHandler.HandleNewPost)))
	apiMux.HandleFunc("GET /posts/", postHandler.HandleGetPosts)
	apiMux.HandleFunc("GET /posts/{category}", postHandler.HandleGetCategoryPosts)
	apiMux.HandleFunc("GET /user/{username}", postHandler.HandleGetUserPosts)
	apiMux.HandleFunc("GET /post/{id}", postHandler.HandleGetPostDetails)
	apiMux.Handle("GET /post/{id}/upvote", handlers.WithAuth(http.HandlerFunc(postHandler.HandleUpvote)))
	apiMux.Handle("GET /post/{id}/downvote", handlers.WithAuth(http.HandlerFunc(postHandler.HandleDownvote)))
	apiMux.Handle("GET /post/{id}/unvote", handlers.WithAuth(http.HandlerFunc(postHandler.HandleUnvote)))

	mux.Handle("/api/", http.StripPrefix("/api", apiMux))
}
