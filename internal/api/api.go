package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Router *chi.Mux
}

func (s *Server) mountMiddlewares() {
	s.Router.Use(middleware.Heartbeat("/ping"))
	s.Router.Use(middleware.Timeout(1 * time.Minute))
	s.Router.Use(middleware.Recoverer)
}

func (s *Server) mountHandlers() {
	s.Router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World"))
	})
}

func NewServer() *Server {
	s := &Server{}
	s.mountMiddlewares()
	s.mountHandlers()
	return s
}
