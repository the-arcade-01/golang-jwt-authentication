package repository

import (
	"context"

	"github.com/the-arcade-01/golang-jwt-authentication/internal/models"
)

type AuthRepositoryInterface interface {
	CreateUser(ctx context.Context, user *models.User) (int, error)
	LogoutUser(ctx context.Context, userID int) (int, error)
	DeleteUser(ctx context.Context, userID int) (int, error)
	GetUserByID(ctx context.Context, userID int) (*models.User, int, error)
	LoginUser(ctx context.Context, user *models.User) (*models.TokenResponse, int, error)
	GenerateTokens(ctx context.Context, userID int, oldRefreshToken string) (*models.TokenResponse, int, error)
	getAuthTokens(ctx context.Context, userID int) (*models.TokenResponse, int, error)
}
