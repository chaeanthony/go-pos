package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/chaeanthony/go-pos/internal/auth"
	"github.com/chaeanthony/go-pos/internal/database"
	"github.com/chaeanthony/go-pos/utils"
)

const (
	JWT_EXPIRATION = 15 * time.Minute
)

func (cfg *APIConfig) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		database.User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.DB.GetUserByEmail(params.Email)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.Password)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	jwtRole := user.Role

	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.JWTSecret,
		JWT_EXPIRATION,
		jwtRole,
	)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusInternalServerError, "Couldn't create access JWT", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusInternalServerError, "Couldn't create refresh token", err)
		return
	}

	_, err = cfg.DB.CreateRefreshToken(database.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusInternalServerError, "Couldn't save refresh token", err)
		return
	}

	auth.SetTokenCookie(w, accessToken, auth.AccessToken, "/", JWT_EXPIRATION, cfg.CookieSameSite, cfg.CookieSecure)
	auth.SetTokenCookie(w, refreshToken, auth.RefreshToken, "/", JWT_EXPIRATION, cfg.CookieSameSite, cfg.CookieSecure)
	utils.RespondJSON(w, cfg.Logger, http.StatusOK, response{
		User:         user,
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}

func (cfg *APIConfig) HandlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r, auth.RefreshToken)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusBadRequest, "Couldn't find token in header", err)
		return
	}

	user, err := cfg.DB.GetUserByRefreshToken(refreshToken)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusUnauthorized, "Couldn't get user for refresh token", err)
		return
	}

	jwtRole := user.Role

	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.JWTSecret,
		JWT_EXPIRATION,
		jwtRole,
	)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusUnauthorized, "Couldn't validate token for refresh", err)
		return
	}

	auth.SetTokenCookie(w, accessToken, auth.AccessToken, "/", JWT_EXPIRATION, http.SameSiteLaxMode, false)
	utils.RespondJSON(w, cfg.Logger, http.StatusOK, response{
		Token: accessToken,
	})
}

func (cfg *APIConfig) HandlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r, auth.RefreshToken)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusBadRequest, "Couldn't find token", err)
		return
	}

	err = cfg.DB.RevokeRefreshToken(refreshToken)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusInternalServerError, "Couldn't revoke session", err)
		return
	}

	auth.ClearTokenCookie(w, auth.RefreshToken, "/", http.SameSiteLaxMode, false)

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *APIConfig) HandlerSession(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r, auth.AccessToken)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusBadRequest, "Couldn't find token", err)
		return
	}
	cfg.Logger.Printf("Checking token: %s", token)
	_, _, err = auth.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		utils.RespondError(w, cfg.Logger, http.StatusUnauthorized, "Invalid session. Couldn't validate token", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
