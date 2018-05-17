package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// User is the  data structure for the users table
type User struct {
	ID       int `gorm:"primary-key;AUTO_INCREMENT"`
	Username string
	Password string
	Name     string
	Mobile   string
	Email    string
}

var tokenAuth *jwtauth.JWTAuth

func init() {
	jwtSecret := getEnv("JWT_SECRET", "thissecretshouldbesecret")
	tokenAuth = jwtauth.New("HS256", []byte(jwtSecret), nil)
}

func main() {
	// listen on localhost:LISTEN_PORT
	listenPort := getEnv("LISTEN_PORT", "3333")
	dbHost := getEnv("DB_HOST", "localhost")

	addr := fmt.Sprintf(":%s", listenPort)

	// database
	dbaddr := fmt.Sprintf("postgresql://tester@%s:5432/rpapoc?sslmode=disable", dbHost)
	db, err := gorm.Open("postgres", dbaddr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Automatically create the "accounts" table based on the Account model.
	db.AutoMigrate(&User{})

	// Insert default user
	var defaultUser User
	db.FirstOrCreate(&defaultUser, User{
		Username: "tester",
		Password: "password",
		Name:     "Bob Smith",
		Email:    "bob@email.com",
		Mobile:   "123-456-7890",
	})

	// http router
	r := chi.NewRouter()

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)
	r.Use(middleware.Logger)
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("We hit the post")
		var authTry = struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{}
		err := json.NewDecoder(r.Body).Decode(&authTry)
		if err != nil {
			fmt.Println("error parsing json request for login. ", err)
			http.Error(w, "Something was wrong with the request body.", http.StatusBadRequest)
			return
		}
		var user User
		fmt.Println("Attempting login for user: ", authTry.Username)
		lookup := db.Where("username = ?", authTry.Username).First(&user)
		if lookup.Error != nil {
			// user not found
			http.Error(w, "User Not Found", http.StatusNotFound)
		} else {
			// user found, check password
			// in reality, the password would be encrypted/encoded before storing in the database
			if (user.Password == "") || (user.Password != authTry.Password) {
				// wrong password try again
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}
			// good to go buddy, return jwt token
			_, tokenString, _ := tokenAuth.Encode(jwtauth.Claims{"username": user.Username})
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, "{\"token\": \"%s\"}", tokenString)
		}
	})

	fmt.Printf("Starting server on %v\n", addr)
	http.ListenAndServe(addr, r)
}

// this is a helper func to fetch environment variables
func getEnv(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}
