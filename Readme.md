## Golang JWT Authentication

This project implements a JWT authentication flow in Golang, featuring both access and refresh tokens. The following APIs have been implemented:

### API Endpoints

#### Public Endpoints

- `GET /api/auth/greet` - Greet endpoint
- `POST /api/auth/users` - Create a new user
- `POST /api/auth/sessions` - Login user

#### Protected Endpoints

- `GET /api/auth/users/me` - Get current user information
- `POST /api/auth/logout` - Logout user
- `DELETE /api/auth/users` - Delete user
- `POST /api/auth/tokens/refresh` - Refresh access token

### Middleware

- JWT verification and authentication
- Request logging
- Claims parsing

### Postman Collection

A Postman collection is included in the repository to help you test the APIs.<br>
`golang-jwt-authentication.postman_collection.json`

### Getting Started

#### Prerequisites

- Go 1.16+
- A `.env` file with the necessary environment variables

#### Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/the-arcade-01/golang-jwt-authentication.git
   cd golang-jwt-authentication
   ```

2. Install dependencies:

   ```sh
   go mod tidy
   ```

3. Create a `.env` file in the root directory and add the required environment variables:
   ```env
    ENV=development
    WEB_URL=http://localhost:5173
    JWT_SECRET_KEY=<secret>
    DB_DRIVER=mysql
    DB_URL=<user>:<password>@tcp(<mysql_container_name>:3306)/<db_name>?parseTime=true
    DB_MAX_IDLE_CONN=10
    DB_MAX_OPEN_CONN=10
    DB_MAX_CONN_TIME_SEC=180
    MYSQL_ROOT_PASSWORD=<password>
    MYSQL_DATABASE=<db_name>
    HTTP_COOKIE_HTTPONLY=false
    HTTP_COOKIE_SECURE=false
    HTTP_REFRESH_TOKEN_EXPIRE=720
    HTTP_ACCESS_TOKEN_EXPIRE=15
   ```

#### Running the Server

1.  Run the db using docker compose in the `scripts` folder
    ```sh
    cd scripts
    docker compose --env-file ../../.env up
    ```
2.  Start the server by running:
    `go run cmd/main.go` OR `make run`
