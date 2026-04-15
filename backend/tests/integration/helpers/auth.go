package helpers

import (
	"testing"

	"backend/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func GenerateTestToken(t *testing.T, userID uint, email, userType string) string {
	token, err := utils.GenerateAccessToken(userID, email, userType)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	return token
}

func GenerateTestTokenPair(t *testing.T, userID uint, email, userType string) *utils.TokenPair {
	pair, err := utils.GenerateTokenPair(userID, email, userType)
	assert.NoError(t, err)
	assert.NotNil(t, pair)
	return pair
}
