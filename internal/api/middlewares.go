package api

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/models"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/utils"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)

		micro := time.Since(start).Microseconds()
		utils.Log.Info("rtt",
			"method", r.Method,
			"url", r.URL.String(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"micro", micro,
			"status", rec.status,
		)
	})
}

func parseClaims(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		userID, ok := claims["userID"].(float64)
		if err != nil && !ok {
			models.ResponseWithJSON(w, http.StatusUnauthorized, &models.ErrorResponse{
				Success: false,
				Status:  http.StatusUnauthorized,
				Error:   "Invalid token, please login again",
			})
			return
		}

		ctx := context.WithValue(r.Context(), utils.UserIDCtxKey, userID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
