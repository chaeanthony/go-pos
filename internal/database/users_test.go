package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUser(t *testing.T) {
	c, err := CreateTestClient(t)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer c.db.Close()

	t.Run("Successful user", func(t *testing.T) {
		user, err := c.CreateUser(CreateUserParams{
			Email:    "test@test.com",
			Password: "testpassword",
			Role:     "user",
		},
		)
		require.NoError(t, err, "Failed to create user")
		require.NotEmpty(t, user.ID, "User ID should not be empty")
		require.Equal(t, "test@test.com", user.Email, "User email should match")

		_, err = c.GetUserById(user.ID)
		require.NoError(t, err, "Failed to get user by ID")

		_, err = c.GetUserByEmail("test@test.com")
		require.NoError(t, err, "Failed to get user by email")

		token, err := c.CreateRefreshToken(CreateRefreshTokenParams{
			Token:     "testtoken",
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(1 * time.Minute),
		})
		require.NoError(t, err, "Failed to create refresh token")
		require.Equal(t, "testtoken", token.Token, "Refresh token should match")
		require.Empty(t, token.RevokedAt, "Refresh token should not be revoked")

		err = c.RevokeRefreshToken(token.Token)
		require.NoError(t, err, "Failed to revoke refresh token")
		revokedToken, err := c.GetRefreshToken(token.Token)
		require.NoError(t, err, "Failed to get revoked refresh token")
		require.NotEmpty(t, revokedToken.RevokedAt, "RevokedAt should not be empty")
	})
}
