package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gaba-bouliva/chirpy-1/internal/auth"
	"github.com/gaba-bouliva/chirpy-1/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
	}

	params := parameters{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil || len(params.Email) < 5 {
		errMsg := "error invalid payload"
		respondWithError(w, http.StatusBadRequest, errMsg, err)
		return
	}

	if len(params.Password) < 5 {
		respondWithError(w, http.StatusBadRequest, "invalid password enter at least a 6 digit long password", err)
		return
	}

	pwd, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error hashing pasword", err)
		return
	}

	CreateUserParams := database.CreateUserParams{
		ID:             uuid.New(),
		Email:          params.Email,
		HashedPassword: pwd,
	}

	user, err := cfg.db.CreateUser(r.Context(), CreateUserParams)
	if err != nil {
		errMsg := "error creating new user in db"
		respondWithError(w, http.StatusInternalServerError, errMsg, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	})

}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type loginDetails struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	loginDetail := loginDetails{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&loginDetail)
	if err != nil || len(loginDetail.Email) < 5 {
		errMsg := "error invalid login credentials provided"
		respondWithError(w, http.StatusBadRequest, errMsg, err)
		return
	}
	usr, err := cfg.db.GetUserByEmail(r.Context(), loginDetail.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "incorrect email or password", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "unexpected error login failed", err)
		return
	}

	err = auth.CheckPasswordHash(loginDetail.Password, usr.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password", err)
		return
	}

	accessToken, err := auth.MakeJWT(usr.ID, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error generating access token", err)
		return
	}

	refreshTokenArgs := database.CreateRefreshTokenParams{
		Token:     auth.MakeRefreshToken(),
		UserID:    usr.ID,
		ExpiresAt: time.Now().Add(time.Hour * 1440), // expires after 60 days
	}

	refreshToken, err := cfg.db.CreateRefreshToken(r.Context(), refreshTokenArgs)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error generating refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:           usr.ID,
		CreatedAt:    usr.CreatedAt,
		UpdatedAt:    usr.UpdatedAt,
		Email:        usr.Email,
		Token:        accessToken,
		RefreshToken: refreshToken.Token,
	})

}

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	// get refresh token from request Authorization headers with format `Bearer <refresh_token>`
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find token", err)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get user from refresh token", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't valid token", err)
		return
	}

	type payload struct {
		Token string `json:"token"`
	}

	respondWithJSON(w, http.StatusOK, payload{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find token", err)
		return
	}

	_, err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error revoking session", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
