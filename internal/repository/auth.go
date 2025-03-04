package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/config"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/models"
	"github.com/the-arcade-01/golang-jwt-authentication/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

const (
	INSERT_USER                = `INSERT INTO users (email, password) VALUES (?,?)`
	COUNT_USER_BY_EMAIL        = `SELECT count(email) FROM users WHERE email = ?`
	FETCH_USER_BY_EMAIL        = `SELECT id, email, password FROM users WHERE email = ?`
	DELETE_TOKEN_REFRESH_TABLE = `DELETE from refresh_tokens_table WHERE user_id = ?`
	FETCH_USER                 = `SELECT id, email, created_at FROM users WHERE id = ?`
	FETCH_REFRESH_TOKEN        = `SELECT refresh_token FROM refresh_tokens_table WHERE user_id = ?`
	INSERT_REFRESH_TOKEN       = `
		INSERT INTO refresh_tokens_table (user_id, refresh_token, expire_time) 
		VALUES (?, ?, ?) 
		ON DUPLICATE KEY UPDATE 
		refresh_token = VALUES(refresh_token), 
		expire_time = VALUES(expire_time), 
		created_at = CURRENT_TIMESTAMP
	`
	DELETE_USER = `DELETE FROM users WHERE id = ?`
)

type AuthRepo struct {
	db   *sql.DB
	auth *jwtauth.JWTAuth
}

func NewAuthRepo() AuthRepositoryInterface {
	return &AuthRepo{
		db:   config.NewAppConfig().DB,
		auth: config.NewAppConfig().JWTAuth,
	}
}

func (r *AuthRepo) CreateUser(ctx context.Context, user *models.User) (int, error) {
	var row int
	err := r.db.QueryRowContext(ctx, COUNT_USER_BY_EMAIL, user.Email).Scan(&row)
	if err != nil && err != sql.ErrNoRows {
		utils.Log.ErrorContext(ctx, "error fetching user", "function", "Create", "error", err)
		return http.StatusInternalServerError, fmt.Errorf("please try again later")
	}

	if row != 0 {
		return http.StatusBadRequest, fmt.Errorf("email already taken, please use different email")
	}

	hashPassword, err := getHashPassword(user.Password)
	if err != nil {
		utils.Log.ErrorContext(ctx, "error on generating hash password", "function", "Create", "error", err)
		return http.StatusInternalServerError, fmt.Errorf("please try again later")
	}

	_, err = r.db.ExecContext(ctx, INSERT_USER, user.Email, hashPassword)
	if err != nil {
		utils.Log.ErrorContext(ctx, "error on saving user in db", "function", "Create", "error", err)
		return http.StatusInternalServerError, fmt.Errorf("please try again later")
	}

	return http.StatusOK, nil
}

func (r *AuthRepo) LogoutUser(ctx context.Context, userID int) (int, error) {
	_, err := r.db.ExecContext(ctx, DELETE_TOKEN_REFRESH_TABLE, userID)
	if err != nil {
		utils.Log.ErrorContext(ctx, "error on deleting from refresh_tokens_table", "function", "Logout", "error", err)
		return http.StatusInternalServerError, fmt.Errorf("please try again later")
	}
	return http.StatusOK, nil
}

func (r *AuthRepo) GetUserByID(ctx context.Context, userID int) (*models.User, int, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, FETCH_USER, userID).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		utils.Log.ErrorContext(ctx, "error on fetching user", "function", "Read", "error", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("please try again later")
	}
	return user, http.StatusOK, nil
}

func (r *AuthRepo) DeleteUser(ctx context.Context, userID int) (int, error) {
	_, err := r.db.ExecContext(ctx, DELETE_USER, userID)
	if err != nil {
		utils.Log.ErrorContext(ctx, "error on deleting user", "function", "Delete", "error", err)
		return http.StatusInternalServerError, fmt.Errorf("please try again later")
	}
	return http.StatusAccepted, nil
}

func (r *AuthRepo) LoginUser(ctx context.Context, user *models.User) (*models.TokenResponse, int, error) {
	existUser := &models.User{}
	err := r.db.QueryRowContext(ctx, FETCH_USER_BY_EMAIL, user.Email).Scan(&existUser.ID, &existUser.Email, &existUser.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, http.StatusBadRequest, fmt.Errorf("please check credentials")
		}
		utils.Log.ErrorContext(ctx, "error on fetching user", "function", "Login", "error", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("please try again later")
	}

	isValid := checkPassword(existUser.Password, user.Password)
	if !isValid {
		return nil, http.StatusUnauthorized, fmt.Errorf("incorrect password, please try again")
	}

	return r.getAuthTokens(ctx, existUser.ID)
}

func (r *AuthRepo) GenerateTokens(ctx context.Context, userID int, oldRefreshToken string) (*models.TokenResponse, int, error) {
	var dbRefreshToken string
	err := r.db.QueryRowContext(ctx, FETCH_REFRESH_TOKEN, userID).Scan(&dbRefreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, http.StatusUnauthorized, fmt.Errorf("please login again")
		}
		utils.Log.ErrorContext(ctx, "error on fetching from refresh_tokens_table", "function", "GenerateAuthTokens", "error", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("please login again")
	}
	if dbRefreshToken != oldRefreshToken {
		return nil, http.StatusUnauthorized, fmt.Errorf("invalid token, please login again")
	}

	return r.getAuthTokens(ctx, userID)
}

func (r *AuthRepo) getAuthTokens(ctx context.Context, userID int) (*models.TokenResponse, int, error) {
	accessToken, err := getToken(userID, r.auth, time.Now().Add(time.Duration(config.Envs.HTTP_ACCESS_TOKEN_EXPIRE)*time.Minute).Unix())
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("please try again later")
	}
	refreshTokenExpire := time.Now().Add(time.Duration(config.Envs.HTTP_REFRESH_TOKEN_EXPIRE) * time.Minute).Unix()
	refreshToken, err := getToken(userID, r.auth, refreshTokenExpire)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("please try again later")
	}

	_, err = r.db.ExecContext(ctx, INSERT_REFRESH_TOKEN, userID, refreshToken, refreshTokenExpire)
	if err != nil {
		utils.Log.ErrorContext(ctx, "error on saving refresh token in db", "function", "LoginUser", "error", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("please try again later")
	}
	return &models.TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken}, http.StatusOK, err
}

func getToken(userID int, auth *jwtauth.JWTAuth, expireTime int64) (string, error) {
	claims := map[string]any{
		"userID": userID,
		"exp":    expireTime,
	}
	_, token, err := auth.Encode(claims)
	if err != nil {
		utils.Log.Error("error on generating auth token", "function", "generateAuthToken", "error", err)
		return "", err
	}
	return token, nil
}

func getHashPassword(password string) (string, error) {
	bytePassword := []byte(password)
	hash, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func checkPassword(hashPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}
