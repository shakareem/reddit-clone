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

	mux := http.NewServeMux()
	registerStaticHandlers(mux)
	handlers.ReqisterAPIHandlers(mux, storage)

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
