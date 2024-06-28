package main

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/mike-kimani/whitepointinventory/auth"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mike-kimani/whitepointinventory/internal/database"
)

const SecretKey = "secret"

func (apiCfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		Email    string `json:"email_address"`
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error parsing json: %v", err))
		return
	}

	if params.Name == "" {
		respondWithError(w, 400, "Name field is empty")
		return
	}

	if params.Email == "" {
		respondWithError(w, 400, "Email field is empty")
		return
	}

	if len(params.Password) < 5 {
		respondWithError(w, 400, "Password field is empty or too short. Must be 5 or more characters")
		return
	}

	if !strings.Contains(params.Email, "@") {
		respondWithError(w, 400, "Invalid email format")
		return
	}

	if !strings.Contains(params.Email, ".") {
		respondWithError(w, 400, "Invalid email format")
		return
	}

	user, err := apiCfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      params.Name,
		Password:  params.Password,
		Email:     params.Email,
	})

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't create user: %v", err))
		return
	}
	respondWithJSON(w, 201, user)
}

func (apiCfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {

	email, password, err := auth.GetPasswordAndEmailFromBody(r)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Auth error: %v", err))
		return
	}

	user, err := apiCfg.DB.GetUserByPasswordAndEmail(r.Context(), database.GetUserByPasswordAndEmailParams{
		Password: password,
		Email:    email,
	})
	if err != nil {
		respondWithError(w, 404, fmt.Sprintf("User not found: %v", err))
		return
	}
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    user.ID.String(),
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	})
	token, err := claims.SignedString([]byte(SecretKey))
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't get token: %v", err))
	}

	cookie := http.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	respondWithJSON(w, 200, user)
}

func (apiCfg *apiConfig) handlerGetUserFromCookie(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		respondWithError(w, 404, fmt.Sprintf("Couldn't get cookie: %v", err))
		return
	}
	tokenString := cookie.Value
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		respondWithError(w, 404, fmt.Sprintf("Couldn't parse token: %v", err))
	}
	claims := token.Claims.(*jwt.StandardClaims)
	userID, err := uuid.Parse(claims.Issuer)

	user, err := apiCfg.DB.GetUserByID(r.Context(), userID)
	if err != nil {
		respondWithError(w, 404, fmt.Sprintf("User not found: %v", err))
	}
	respondWithJSON(w, 200, user)
}

func (apiCfg *apiConfig) handlerUserLogout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
}

func (apiCfg *apiConfig) handlerGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := apiCfg.DB.GetUsers(r.Context())
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Couldn't get users: %v", err))
		return
	}
	respondWithJSON(w, 200, users)
}
