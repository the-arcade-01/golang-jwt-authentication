package service

import (
	"context"

	"github.com/the-arcade-01/golang-jwt-authentication/internal/models"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/repository"
)

type AuthService struct {
	repo repository.AuthRepositoryInterface
}

func NewAuthService() AuthServiceInterface {
	return &AuthService{
		repo: repository.NewAuthRepo(),
	}
}

func (svc *AuthService) CreateUser(ctx context.Context, body *models.AuthReqBody) (*models.Response, *models.ErrorResponse) {
	user := &models.User{Email: body.Email, Password: body.Password}
	status, err := svc.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, models.NewErrorResponse(status, err)
	}
	return &models.Response{Success: true, Status: status}, nil
}

func (svc *AuthService) LogoutUser(ctx context.Context, userID int) (*models.Response, *models.ErrorResponse) {
	status, err := svc.repo.LogoutUser(ctx, userID)
	if err != nil {
		return nil, models.NewErrorResponse(status, err)
	}
	return &models.Response{Success: true, Status: status}, nil
}

func (svc *AuthService) DeleteUser(ctx context.Context, userID int) (*models.Response, *models.ErrorResponse) {
	status, err := svc.repo.DeleteUser(ctx, userID)
	if err != nil {
		return nil, models.NewErrorResponse(status, err)
	}
	return &models.Response{Success: true, Status: status}, nil
}

func (svc *AuthService) GetUserByID(ctx context.Context, userID int) (*models.Response, *models.ErrorResponse) {
	user, status, err := svc.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, models.NewErrorResponse(status, err)
	}
	return &models.Response{Success: true, Status: status, Data: []*models.User{user}}, nil
}

func (svc *AuthService) LoginUser(ctx context.Context, body *models.AuthReqBody) (*models.TokenResponse, *models.ErrorResponse) {
	user := &models.User{Email: body.Email, Password: body.Password}
	tokenRes, status, err := svc.repo.LoginUser(ctx, user)
	if err != nil {
		return nil, models.NewErrorResponse(status, err)
	}
	return tokenRes, nil
}

func (svc *AuthService) GenerateTokens(ctx context.Context, userID int, oldRefreshToken string) (*models.TokenResponse, *models.ErrorResponse) {
	tokenRes, status, err := svc.repo.GenerateTokens(ctx, userID, oldRefreshToken)
	if err != nil {
		return nil, models.NewErrorResponse(status, err)
	}
	return tokenRes, nil
}
