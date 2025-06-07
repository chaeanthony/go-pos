package api

import (
	"net/http"

	"github.com/chaeanthony/go-pos/internal/auth"
	"github.com/chaeanthony/go-pos/utils"
)

// type contextKey string

// const (
// userIDKey contextKey = "userID"
// roleKey   contextKey = "role"
// )

func (cfg *APIConfig) StoreAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r, auth.AccessToken)
		if err != nil {
			utils.RespondError(w, cfg.Logger, http.StatusUnauthorized, "Couldn't find token", err)
			return
		}

		userID, role, err := auth.ValidateJWT(token, cfg.JWTSecret)
		if err != nil {
			utils.RespondError(w, cfg.Logger, http.StatusUnauthorized, "Couldn't validate token for store access", err)
			return
		}

		_, err = cfg.DB.GetUserById(userID)
		if err != nil {
			utils.RespondError(w, cfg.Logger, http.StatusUnauthorized, "Couldn't find user", err)
			return
		}

		if role != "store" {
			utils.RespondError(w, cfg.Logger, http.StatusForbidden, "You are not authorized to access this resource", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}
