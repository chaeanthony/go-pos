package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	CreateRefreshTokenParams
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	RevokedAt *time.Time `json:"revoked_at"`
}

type CreateRefreshTokenParams struct {
	Token     string    `json:"token"`
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (c *Client) CreateRefreshToken(params CreateRefreshTokenParams) (RefreshToken, error) {
	query := `
		INSERT INTO refresh_tokens (
			token,
			created_at,
			updated_at,
			user_id,
			expires_at
		) VALUES (?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?)
	`
	_, err := c.db.Exec(query, params.Token, params.UserID.String(), params.ExpiresAt.Format(TIME_LAYOUT))
	if err != nil {
		return RefreshToken{}, fmt.Errorf("couldn't create refresh token: %w", err)
	}

	return c.GetRefreshToken(params.Token)
}

func (c *Client) RevokeRefreshToken(token string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = CURRENT_TIMESTAMP
		WHERE token = ?
	`
	_, err := c.db.Exec(query, token)
	return err
}

func (c *Client) GetRefreshToken(token string) (RefreshToken, error) {
	query := `
		SELECT token, created_at, updated_at, user_id, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token = ?
	`
	var rt RefreshToken
	var userID string
	var created_at, updated_at, expires_at string
	var revoked_at *string

	err := c.db.QueryRow(query, token).
		Scan(&rt.Token, &created_at, &updated_at, &userID, &expires_at, &revoked_at)
	if err != nil {
		if err == sql.ErrNoRows {
			return RefreshToken{}, nil
		}
		return RefreshToken{}, err
	}

	rt.UserID, err = uuid.Parse(userID)
	if err != nil {
		return RefreshToken{}, err
	}

	rt.CreatedAt, err = time.Parse(TIME_LAYOUT, created_at)
	if err != nil {
		return RefreshToken{}, err
	}
	rt.UpdatedAt, err = time.Parse(TIME_LAYOUT, updated_at)
	if err != nil {
		return RefreshToken{}, err
	}
	rt.ExpiresAt, err = time.Parse(TIME_LAYOUT, expires_at)
	if err != nil {
		return RefreshToken{}, err
	}
	if revoked_at != nil {
		t, err := time.Parse(TIME_LAYOUT, *revoked_at)
		if err != nil {
			return rt, fmt.Errorf("couldn't parse revoked_at: %w", err)
		}
		rt.RevokedAt = &t
	}

	return rt, nil
}

func (c *Client) DeleteRefreshToken(token string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE token = ?
	`
	_, err := c.db.Exec(query, token)
	return err
}
