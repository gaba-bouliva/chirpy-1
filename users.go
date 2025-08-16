package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gaba-bouliva/chirpy-1/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
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

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*3)
	defer cancelFunc()

	CreateUserParams := database.CreateUserParams{
		ID:    uuid.New(),
		Email: params.Email,
	}

	user, err := cfg.db.CreateUser(ctx, CreateUserParams)
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
