package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"todo-backend/pkg/utils"
)

func TestGenerateAndParseToken(t *testing.T) {
	token, err := utils.GenerateToken(1, "alice")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := utils.ParseToken(token)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "alice", claims.Username)
}

func TestParseToken_Invalid(t *testing.T) {
	claims, err := utils.ParseToken("invalid.token.string")
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, utils.ErrInvalidToken)
}

func TestParseToken_Empty(t *testing.T) {
	claims, err := utils.ParseToken("")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestGenerateToken_DifferentUsers(t *testing.T) {
	token1, err := utils.GenerateToken(1, "alice")
	assert.NoError(t, err)

	token2, err := utils.GenerateToken(2, "bob")
	assert.NoError(t, err)

	assert.NotEqual(t, token1, token2)

	claims1, err := utils.ParseToken(token1)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), claims1.UserID)
	assert.Equal(t, "alice", claims1.Username)

	claims2, err := utils.ParseToken(token2)
	assert.NoError(t, err)
	assert.Equal(t, uint(2), claims2.UserID)
	assert.Equal(t, "bob", claims2.Username)
}
