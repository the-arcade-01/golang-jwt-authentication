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

// CreateUser creates a new user in the database.
// It first checks if a user with the given email already exists.
// If the email is already taken, it returns a BadRequest status with an appropriate error message.
// If the email is not taken, it hashes the user's password and saves the user in the database.
// It returns an OK status if the user is successfully created, or an InternalServerError status if there is an error during the process.
//
// Parameters:
//   - ctx: The context for the request.
//   - user: A pointer to the User model containing the user's details.
//
// Returns:
//   - int: The HTTP status code indicating the result of the operation.
//   - error: An error message if there was an issue during the operation.
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

// LogoutUser logs out a user by deleting their refresh token from the database.
// It takes a context and a userID as parameters and returns an HTTP status code and an error.
//
// Parameters:
//   - ctx: The context for the request, used for timeout and cancellation.
//   - userID: The ID of the user to log out.
//
// Returns:
//   - int: HTTP status code indicating the result of the operation.
//   - error: An error message if the operation fails, otherwise nil.
func (r *AuthRepo) LogoutUser(ctx context.Context, userID int) (int, error) {
	_, err := r.db.ExecContext(ctx, DELETE_TOKEN_REFRESH_TABLE, userID)
	if err != nil {
		utils.Log.ErrorContext(ctx, "error on deleting from refresh_tokens_table", "function", "Logout", "error", err)
		return http.StatusInternalServerError, fmt.Errorf("please try again later")
	}
	return http.StatusOK, nil
}

// GetUserByID retrieves a user from the database by their user ID.
// It takes a context and a user ID as parameters and returns a pointer to a User model,
// an HTTP status code, and an error if any occurred during the process.
//
// Parameters:
//   - ctx: The context for the request, used for timeout and cancellation.
//   - userID: The ID of the user to be retrieved.
//
// Returns:
//   - *models.User: A pointer to the User model if found.
//   - int: An HTTP status code indicating the result of the operation.
//   - error: An error if any occurred during the process.
//
// Possible errors:
//   - If the user is not found or any other error occurs during the database query,
//     an error is logged and a generic error message is returned with an HTTP 500 status code.
func (r *AuthRepo) GetUserByID(ctx context.Context, userID int) (*models.User, int, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, FETCH_USER, userID).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		utils.Log.ErrorContext(ctx, "error on fetching user", "function", "Read", "error", err)
		return nil, http.StatusInternalServerError, fmt.Errorf("please try again later")
	}
	return user, http.StatusOK, nil
}

// DeleteUser deletes a user from the database based on the provided userID.
// It returns an HTTP status code and an error if any occurs during the deletion process.
//
// Parameters:
//   - ctx: The context for the request, used for cancellation and deadlines.
//   - userID: The ID of the user to be deleted.
//
// Returns:
//   - int: An HTTP status code indicating the result of the operation.
//   - error: An error message if the deletion fails, otherwise nil.
func (r *AuthRepo) DeleteUser(ctx context.Context, userID int) (int, error) {
	_, err := r.db.ExecContext(ctx, DELETE_USER, userID)
	if err != nil {
		utils.Log.ErrorContext(ctx, "error on deleting user", "function", "Delete", "error", err)
		return http.StatusInternalServerError, fmt.Errorf("please try again later")
	}
	return http.StatusAccepted, nil
}

// LoginUser authenticates a user by verifying their email and password.
// It fetches the user details from the database using the provided email,
// checks if the password is correct, and returns authentication tokens if successful.
//
// Parameters:
//   - ctx: The context for the request, used for timeout and cancellation.
//   - user: A pointer to the User model containing the email and password for authentication.
//
// Returns:
//   - A pointer to the TokenResponse model containing the authentication tokens if login is successful.
//   - An integer representing the HTTP status code.
//   - An error if any issue occurs during the login process.
//
// Possible HTTP status codes:
//   - http.StatusBadRequest: If the user does not exist.
//   - http.StatusUnauthorized: If the password is incorrect.
//   - http.StatusInternalServerError: If there is an error during the database query or token generation.
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

// GenerateTokens generates new authentication tokens for a user.
// It validates the provided old refresh token against the stored token in the database.
// If the tokens match, it generates and returns new authentication tokens.
// If the tokens do not match or an error occurs during the process, it returns an appropriate error and status code.
//
// Parameters:
//   - ctx: The context for the request.
//   - userID: The ID of the user for whom the tokens are being generated.
//   - oldRefreshToken: The old refresh token provided by the user.
//
// Returns:
//   - *models.TokenResponse: The new authentication tokens if successful.
//   - int: The HTTP status code indicating the result of the operation.
//   - error: An error message if the operation fails.
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

// getAuthTokens generates and returns new access and refresh tokens for a given user ID.
// It stores the refresh token in the database and returns a TokenResponse containing both tokens.
//
// Parameters:
//   - ctx: The context for the request, used for timeout and cancellation.
//   - userID: The ID of the user for whom the tokens are being generated.
//
// Returns:
//   - *models.TokenResponse: A struct containing the generated access and refresh tokens.
//   - int: The HTTP status code indicating the result of the operation.
//   - error: An error object if an error occurred, otherwise nil.
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

// getToken generates a JWT token for a given user ID with an expiration time.
// It takes the user ID, a JWTAuth instance, and the expiration time as parameters.
// It returns the generated token as a string and an error if the token generation fails.
//
// Parameters:
//   - userID: The ID of the user for whom the token is being generated.
//   - auth: A pointer to a jwtauth.JWTAuth instance used for encoding the token.
//   - expireTime: The expiration time of the token in Unix time format.
//
// Returns:
//   - string: The generated JWT token.
//   - error: An error if the token generation fails.
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
