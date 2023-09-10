package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
)

var (
	DATABASE_URL, DB_DRIVER, PORT string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Coudn't load env file!!")
	}

	DATABASE_URL = os.Getenv("DATABASE_URL")
	DB_DRIVER = os.Getenv("DB_DRIVER")
	PORT = os.Getenv("PORT")
}

func DBClient() (*sql.DB, error) {
	db, err := sql.Open(DB_DRIVER, DATABASE_URL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	fmt.Println("Connected to DB!!")
	return db, nil
}

type Server struct {
	Router *chi.Mux
	DB     *sql.DB
}

func CreateServer(db *sql.DB) *Server {
	server := &Server{
		Router: chi.NewRouter(),
		DB:     db,
	}
	return server
}

func (server *Server) MountHandlers() {
	server.Router.Route("/user", func(userRouter chi.Router) {
		userRouter.Post("/login", server.LoginUser)
		userRouter.Post("/", server.CreateUser)
		userRouter.Get("/{id}", server.GetUser)
		// userRouter.Group(func(r chi.Router) {
		// 	r.Get("/{id}", server.GetUser)
		// })
	})
}

func main() {
	db, err := DBClient()
	if err != nil {
		log.Fatal(err)
	}

	server := CreateServer(db)
	server.MountHandlers()
	fmt.Printf("server running on port%v\n", PORT)
	http.ListenAndServe(PORT, server.Router)
}

type User struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
	Hash  string `json:"hash"`
}

type UserRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Response struct {
	Id int `json:"id"`
}

func (server *Server) LoginUser(w http.ResponseWriter, r *http.Request) {
	userReqBody := new(UserRequestBody)
	if err := json.NewDecoder(r.Body).Decode(userReqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please provide the correct input!!"))
		return
	}
	var hashPassword string

	query := `SELECT hash FROM User where email = ?`
	err := server.DB.QueryRow(query, userReqBody.Email).Scan(&hashPassword)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please provide the correct input!!"))
		return
	}

	if !checkPassword(hashPassword, userReqBody.Password) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect password please check again"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login successful"))
}

func (server *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	userReqBody := new(UserRequestBody)
	if err := json.NewDecoder(r.Body).Decode(userReqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please provide the correct input!!"))
		return
	}

	hashPassword, err := getHashPassword(userReqBody.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something bad happened on the server :("))
		return
	}

	query := `INSERT INTO User (email, hash) VALUES (?, ?)`
	result, err := server.DB.Exec(query, userReqBody.Email, hashPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something bad happened on the server :("))
		return
	}
	recordId, _ := result.LastInsertId()
	response := Response{
		Id: int(recordId),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (server *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	query := `SELECT * FROM User WHERE id = ?`

	rows, err := server.DB.Query(query, id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please provide the correct input!!"))
		return
	}

	user := new(User)
	for rows.Next() {
		err := rows.Scan(&user.Id, &user.Email, &user.Hash)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Please provide the correct input!!"))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
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

// func checkHash() {
// 	for {
// 		var password string

// 		fmt.Scan(&password)

// 		bytePassword := []byte(password)

// 		hash, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
// 		if err != nil {
// 			log.Fatal("meh")
// 		}

// 		fmt.Println("your hash", string(hash))
// 	}
// }
