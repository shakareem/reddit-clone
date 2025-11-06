package server

import (
	"log"
	"net/http"
	"redditclone/internal/server/handlers"
	"redditclone/internal/storage"
	"time"
)

type RedditCloneServer struct {
	Server  *http.Server
	Storage storage.Storage
}

func NewServer() RedditCloneServer {
	storage := storage.NewInMemoryStorage()
	userHandler := handlers.NewUserHandler(storage)

	mux := http.NewServeMux()
	registerStaticHandlers(mux)
	reqisterAPIHandlers(mux, userHandler)

	log.Println("Starting server on :8081")
	server := &http.Server{
		Addr:         ":8081",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return RedditCloneServer{
		Server:  server,
		Storage: storage,
	}
}

func (s *RedditCloneServer) Run() error {
	return s.Server.ListenAndServe()
}

func registerStaticHandlers(mux *http.ServeMux) {
	staticHTMLHandler := http.FileServer(http.Dir("./web/html"))
	mux.Handle("/", staticHTMLHandler)

	staticHandler := http.FileServer(http.Dir("./web"))
	mux.Handle("/static/", http.StripPrefix("/static/", staticHandler))
}

func reqisterAPIHandlers(mux *http.ServeMux, userHandler *handlers.UserHandler) {
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("POST /register", userHandler.HandleRegister)
	apiMux.HandleFunc("POST /login", userHandler.HandleLogIn)

	mux.Handle("/api/", http.StripPrefix("/api", apiMux))
}
