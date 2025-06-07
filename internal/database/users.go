package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreateUserParams
}

type CreateUserParams struct {
	Email     string `json:"email"`
	Password  string `json:"-"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

type UpdateUserParams struct {
	ID        uuid.UUID
	Email     string
	Password  string
	FirstName string
	LastName  string
}

func (c *Client) CreateUser(params CreateUserParams) (User, error) {
	id := uuid.New()

	query := `
		INSERT INTO users
			(id, created_at, updated_at, email, password_hash, first_name, last_name, role)
		VALUES
			(?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?, ?, ?)
	`
	_, err := c.db.Exec(query, id.String(), params.Email, params.Password, params.FirstName, params.LastName, params.Role)
	if err != nil {
		return User{}, fmt.Errorf("couldn't create user: %w", err)
	}

	user, err := c.GetUserById(id)
	if err != nil {
		return User{}, fmt.Errorf("couldn't retreive user after creation: %w", err)
	}

	return user, nil
}

func (c *Client) GetUserById(id uuid.UUID) (User, error) {
	query := `
		SELECT id, created_at, updated_at, email, password_hash, role, first_name, last_name
		FROM users
		WHERE id = ?
	`
	var user User
	var idStr string
	var created_at, updated_at string

	err := c.db.QueryRow(query, id.String()).Scan(&idStr, &created_at, &updated_at, &user.Email, &user.Password, &user.Role, &user.FirstName, &user.LastName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, nil
		}
		return User{}, err
	}

	user.CreatedAt, err = time.Parse(TIME_LAYOUT, created_at)
	if err != nil {
		return User{}, err
	}
	user.UpdatedAt, err = time.Parse(TIME_LAYOUT, updated_at)
	if err != nil {
		return User{}, err
	}

	user.ID, err = uuid.Parse(idStr)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (c *Client) GetUserByEmail(email string) (User, error) {
	query := `
		SELECT id, email, password_hash, role, first_name, last_name
		FROM users
		WHERE email = ?
	`
	var user User
	var id string
	err := c.db.QueryRow(query, email).Scan(&id, &user.Email, &user.Password, &user.Role, &user.FirstName, &user.LastName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, nil
		}
		return User{}, err
	}
	user.ID, err = uuid.Parse(id)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (c *Client) GetUserByRefreshToken(token string) (User, error) {
	query := `
		SELECT u.id, u.email, u.role
		FROM users u
		JOIN refresh_tokens rt ON u.id = rt.user_id
		WHERE rt.token = ?
	`

	var user User
	var id string
	err := c.db.QueryRow(query, token).Scan(&id, &user.Email, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, nil
		}
		return User{}, err
	}
	user.ID, err = uuid.Parse(id)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (c *Client) UpdateUser(params UpdateUserParams) (User, error) {
	query := `UPDATE users SET email = ?, password_hash = ?, first_name =? , last_name = ? WHERE id = ?`
	_, err := c.db.Exec(query, params.ID.String(), params.Email, params.Password, params.FirstName, params.LastName)
	if err != nil {
		return User{}, err
	}

	return c.GetUserById(params.ID)
}

func (c *Client) DeleteUser(id uuid.UUID) error {
	query := `
		DELETE FROM users
		WHERE id = ?
	`
	_, err := c.db.Exec(query, id.String())
	return err
}
