package api

import (
	"encoding/json"
	"net/http"

	"github.com/chaeanthony/go-pos/internal/auth"
	"github.com/chaeanthony/go-pos/internal/database"
	"github.com/chaeanthony/go-pos/utils"
)

func (cfg *APIConfig) HandlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password  string `json:"password"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.Password == "" || params.Email == "" {
		utils.RespondError(w, cfg.Logger, http.StatusBadRequest, "Email and password are required", nil)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := cfg.DB.CreateUser(database.CreateUserParams{
		Email:     params.Email,
		Password:  hashedPassword,
		FirstName: params.FirstName,
		LastName:  params.LastName,
		Role:      "user", // Default role, can be changed later
	})
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	utils.RespondJSON(w, cfg.Logger, http.StatusCreated, user)
}
