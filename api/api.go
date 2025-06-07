package api

import (
	"errors"
	"net/http"

	"github.com/chaeanthony/go-pos/internal/database"
	"github.com/charmbracelet/log"
)

var ErrAuthorizeUser = errors.New("couldn't authorize user")
var ErrAuthorizeUserRole = errors.New("couldn't authorize user role")

type APIConfig struct {
	DB             *database.Client
	Port           string
	JWTSecret      string
	Logger         *log.Logger
	Hub            *Hub
	CookieSecure   bool
	CookieSameSite http.SameSite
}

func (cfg *APIConfig) HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
