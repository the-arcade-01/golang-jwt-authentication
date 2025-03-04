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

// mountMiddlewares sets up the middleware stack for the server's router.
// It includes the following middlewares:
// - Heartbeat: Responds to /ping requests to check server health.
// - Timeout: Sets a timeout for requests to 1 minute.
// - Recoverer: Recovers from panics and returns a 500 status code.
// - requestLogger: Logs incoming requests.
// - CORS: Configures Cross-Origin Resource Sharing with specified options.
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

// mountHandlers sets up the routing for the authentication-related endpoints.
// It initializes the authentication handlers and defines the routes for user
// creation, login, and greeting. It also sets up a group of routes that require
// JWT authentication, including routes for getting user information, logging out,
// deleting a user, and refreshing tokens.
func (s *Server) mountHandlers() {
	authHandlers := handlers.NewAuthHandlers()
	authRouter := chi.NewRouter()
	authRouter.Get("/greet", authHandlers.Greet)
	authRouter.Post("/users", authHandlers.CreateUser)
	authRouter.Post("/login", authHandlers.LoginUser)
	authRouter.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(config.NewAppConfig().JWTAuth))
		r.Use(jwtauth.Authenticator(config.NewAppConfig().JWTAuth))
		r.Use(parseClaims)

		r.Get("/users/me", authHandlers.GetUserByID)
		r.Post("/logout", authHandlers.LogoutUser)
		r.Delete("/users", authHandlers.DeleteUser)
		r.Post("/tokens/refresh", authHandlers.RefreshToken)
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
