package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/the-arcade-01/golang-jwt-authentication/internal/config"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/models"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/service"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/utils"
)

type AuthHandlers struct {
	svc service.AuthServiceInterface
}

func NewAuthHandlers() AuthHandlersInterface {
	return &AuthHandlers{
		svc: service.NewAuthService(),
	}
}

func (h *AuthHandlers) Greet(w http.ResponseWriter, r *http.Request) {
	models.ResponseWithJSON(w, http.StatusOK, []byte("Hello, World"))
}

func (h *AuthHandlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	var body *models.AuthReqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		models.ResponseWithJSON(w, http.StatusBadRequest, models.NewErrorResponse(http.StatusBadRequest, fmt.Errorf("please provide valid input")))
		return
	}
	defer r.Body.Close()

	result, err := h.svc.CreateUser(r.Context(), body)
	if err != nil {
		models.ResponseWithJSON(w, err.Status, err)
		return
	}
	models.ResponseWithJSON(w, result.Status, result)
}

func (h *AuthHandlers) LogoutUser(w http.ResponseWriter, r *http.Request) {
	userID := int(r.Context().Value(utils.UserIDCtxKey).(float64))
	result, err := h.svc.LogoutUser(r.Context(), userID)
	if err != nil {
		models.ResponseWithJSON(w, err.Status, err)
		return
	}

	h.setCookie(w, "")
	models.ResponseWithJSON(w, result.Status, result)
}

func (h *AuthHandlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := int(r.Context().Value(utils.UserIDCtxKey).(float64))
	result, err := h.svc.DeleteUser(r.Context(), userID)
	if err != nil {
		models.ResponseWithJSON(w, err.Status, err)
		return
	}
	models.ResponseWithJSON(w, result.Status, result)
}

func (h *AuthHandlers) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userID := int(r.Context().Value(utils.UserIDCtxKey).(float64))
	result, err := h.svc.GetUserByID(r.Context(), userID)
	if err != nil {
		models.ResponseWithJSON(w, err.Status, err)
		return
	}
	models.ResponseWithJSON(w, result.Status, result)
}

func (h *AuthHandlers) LoginUser(w http.ResponseWriter, r *http.Request) {
	var body *models.AuthReqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		models.ResponseWithJSON(w, http.StatusBadRequest, models.NewErrorResponse(http.StatusBadRequest, fmt.Errorf("please provide valid input")))
		return
	}
	defer r.Body.Close()

	tokensResponse, err := h.svc.LoginUser(r.Context(), body)
	if err != nil {
		models.ResponseWithJSON(w, err.Status, err)
		return
	}

	h.setCookie(w, tokensResponse.RefreshToken)
	models.ResponseWithJSON(w, http.StatusOK, &models.Response{Success: true, Status: http.StatusOK, Data: tokensResponse})
}

// RefreshToken handles the token refresh process for authenticated users.
// It retrieves the JWT cookie from the request, validates it, and generates new access and refresh tokens.
// If the JWT cookie is missing or invalid, it responds with an unauthorized status.
// If the token generation is successful, it sets the new refresh token in the cookie and responds with the new access token.
//
// @param w http.ResponseWriter - the response writer to send the response
// @param r *http.Request - the incoming HTTP request containing the JWT cookie
func (h *AuthHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt")
	if err != nil || cookie.Value == "" {
		models.ResponseWithJSON(w, http.StatusUnauthorized, models.NewErrorResponse(http.StatusUnauthorized, fmt.Errorf("please login again")))
		return
	}
	userID := int(r.Context().Value(utils.UserIDCtxKey).(float64))
	tokensResponse, er := h.svc.GenerateTokens(r.Context(), userID, cookie.Value)
	if er != nil {
		models.ResponseWithJSON(w, er.Status, er)
		return
	}

	h.setCookie(w, tokensResponse.RefreshToken)
	models.ResponseWithJSON(w, http.StatusOK, &models.Response{Success: true, Status: http.StatusOK, Data: tokensResponse})
}

func (h *AuthHandlers) setCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		HttpOnly: config.Envs.HTTP_COOKIE_HTTPONLY,
		Secure:   config.Envs.HTTP_COOKIE_SECURE,
		SameSite: http.SameSiteLaxMode,
	})
}
