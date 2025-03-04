package service

import (
	"context"

	"github.com/the-arcade-01/golang-jwt-authentication/internal/models"
)

type AuthServiceInterface interface {
	CreateUser(ctx context.Context, body *models.AuthReqBody) (*models.Response, *models.ErrorResponse)
	LogoutUser(ctx context.Context, userID int) (*models.Response, *models.ErrorResponse)
	DeleteUser(ctx context.Context, userID int) (*models.Response, *models.ErrorResponse)
	GetUserByID(ctx context.Context, userID int) (*models.Response, *models.ErrorResponse)
	LoginUser(ctx context.Context, body *models.AuthReqBody) (*models.TokenResponse, *models.ErrorResponse)
	GenerateTokens(ctx context.Context, userID int, oldRefreshToken string) (*models.TokenResponse, *models.ErrorResponse)
}
