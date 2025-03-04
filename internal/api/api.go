package api

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/config"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/handlers"
)

type Server struct {
	Router *chi.Mux
}

func (s *Server) mountMiddlewares() {
	s.Router.Use(middleware.Heartbeat("/ping"))
	s.Router.Use(middleware.Timeout(1 * time.Minute))
	s.Router.Use(middleware.Recoverer)
	s.Router.Use(requestLogger)
	s.Router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{config.Envs.WEB_URL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

func (s *Server) mountHandlers() {
	authHandlers := handlers.NewAuthHandlers()
	authRouter := chi.NewRouter()
	authRouter.Get("/greet", authHandlers.Greet)
	authRouter.Post("/signup", authHandlers.CreateUser)
	authRouter.Post("/login", authHandlers.LoginUser)
	authRouter.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(config.NewAppConfig().JWTAuth))
		r.Use(jwtauth.Authenticator(config.NewAppConfig().JWTAuth))
		r.Use(parseClaims)

		r.Get("/users/me", authHandlers.GetUserByID)
		r.Post("/logout", authHandlers.LogoutUser)
		r.Delete("/delete", authHandlers.DeleteUser)
		r.Post("/refresh-token", authHandlers.RefreshToken)
	})
	s.Router.Mount("/api/auth", authRouter)
}

func NewServer() *Server {
	s := &Server{
		Router: chi.NewMux(),
	}
	s.mountMiddlewares()
	s.mountHandlers()
	return s
}
