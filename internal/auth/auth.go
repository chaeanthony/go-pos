package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TokenType string

const (
	TokenTypeAccess TokenType = "go-home-pos"
	AccessToken     TokenType = "access_token"
	RefreshToken    TokenType = "refresh_token"
)

type CustomClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

var ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")

func HashPassword(password string) (string, error) {
	dat, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(dat), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration, role string) (string, error) {
	signingKey := []byte(tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &CustomClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    string(TokenTypeAccess),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
			Subject:   userID.String(),
		},
	})
	return token.SignedString(signingKey)
}

func ValidateJWT(tokenString, tokenSecret string) (userID uuid.UUID, role string, err error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return uuid.Nil, "", err
	}

	if !token.Valid {
		return uuid.Nil, "", fmt.Errorf("token invalid")
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, "", err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, "", err
	}
	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, "", fmt.Errorf("invalid issuer")
	}

	id, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("invalid user ID: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return uuid.Nil, "", fmt.Errorf("failed to parse claims")
	}

	return id, claims.Role, nil
}

func GetBearerToken(r *http.Request, tokenType TokenType) (string, error) {
	// 1. Check Authorization header for Bearer token
	if r.Method == http.MethodPost {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			return token, nil
		}
	}

	// 2. Check cookie for access token
	cookie, err := r.Cookie(string(tokenType))
	if err == nil {
		return cookie.Value, nil
	}

	return "", nil
}

func MakeRefreshToken() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

func SetTokenCookie(w http.ResponseWriter, token string, tokenType TokenType, path string, expireTime time.Duration, sameSite http.SameSite, secure bool) {
	accessCookie := &http.Cookie{
		Name:     string(tokenType),
		Value:    token,
		Path:     path,
		Expires:  time.Now().Add(expireTime), // e.g., 15 minutes expiry
		HttpOnly: true,                       // prevents JavaScript access (mitigates XSS)
		Secure:   secure,                     // set to true in production (HTTPS)
		SameSite: sameSite,                   // helps mitigate CSRF
	}
	http.SetCookie(w, accessCookie)
}

func ClearTokenCookie(w http.ResponseWriter, tokenType TokenType, path string, sameSite http.SameSite, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     string(tokenType),
		Value:    "",
		Path:     path,            // match the cookie path used when setting it
		Expires:  time.Unix(0, 0), // set expiration to the past
		MaxAge:   -1,              // delete cookie immediately
		HttpOnly: true,            // match flags used when setting cookie
		Secure:   secure,          // set to true in production (HTTPS)
		SameSite: sameSite,        // helps mitigate CSRF
	})
}
