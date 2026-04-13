package utils

import (
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET_KEY", "test_secret_key_for_jwt_testing")
	os.Setenv("JWT_ACCESS_EXPIRES", "900")
	os.Setenv("JWT_REFRESH_EXPIRES", "604800")
	code := m.Run()
	os.Exit(code)
}

func TestGenerateAccessToken(t *testing.T) {
	userID := uint(1)
	email := "test@example.com"
	role := "user"

	token, err := GenerateAccessToken(userID, email, role)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("GenerateAccessToken returned empty token")
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID mismatch: got %d, want %d", claims.UserID, userID)
	}
	if claims.Email != email {
		t.Errorf("Email mismatch: got %s, want %s", claims.Email, email)
	}
	if claims.Role != role {
		t.Errorf("Role mismatch: got %s, want %s", claims.Role, role)
	}
	if claims.Type != AccessToken {
		t.Errorf("TokenType mismatch: got %s, want %s", claims.Type, AccessToken)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	userID := uint(1)
	email := "test@example.com"
	role := "user"

	token, err := GenerateRefreshToken(userID, email, role)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("GenerateRefreshToken returned empty token")
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID mismatch: got %d, want %d", claims.UserID, userID)
	}
	if claims.Email != email {
		t.Errorf("Email mismatch: got %s, want %s", claims.Email, email)
	}
	if claims.Type != RefreshToken {
		t.Errorf("TokenType mismatch: got %s, want %s", claims.Type, RefreshToken)
	}
}

func TestGenerateTokenPair(t *testing.T) {
	userID := uint(1)
	email := "test@example.com"
	role := "user"

	pair, err := GenerateTokenPair(userID, email, role)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	if pair.AccessToken == "" {
		t.Fatal("AccessToken is empty")
	}
	if pair.RefreshToken == "" {
		t.Fatal("RefreshToken is empty")
	}
	if pair.ExpiresIn == 0 {
		t.Fatal("ExpiresIn is zero")
	}

	accessClaims, err := ValidateToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken failed for access token: %v", err)
	}
	if accessClaims.Type != AccessToken {
		t.Errorf("Access token type mismatch: got %s, want %s", accessClaims.Type, AccessToken)
	}

	refreshClaims, err := ValidateToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("ValidateToken failed for refresh token: %v", err)
	}
	if refreshClaims.Type != RefreshToken {
		t.Errorf("Refresh token type mismatch: got %s, want %s", refreshClaims.Type, RefreshToken)
	}
}

func TestValidateToken_ValidToken(t *testing.T) {
	userID := uint(123)
	email := "valid@example.com"
	role := "user"

	token, _ := GenerateAccessToken(userID, email, role)

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed for valid token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID mismatch: got %d, want %d", claims.UserID, userID)
	}
	if claims.Email != email {
		t.Errorf("Email mismatch: got %s, want %s", claims.Email, email)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	_, err := ValidateToken("invalid.token.string")
	if err == nil {
		t.Fatal("ValidateToken should fail for invalid token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", "secret_a")
	token, _ := GenerateAccessToken(1, "a@test.com", "user")
	os.Setenv("JWT_SECRET_KEY", "secret_b")

	_, err := ValidateToken(token)
	if err == nil {
		t.Fatal("ValidateToken should fail when using wrong secret")
	}
}

func TestRefreshAccessToken_ValidRefreshToken(t *testing.T) {
	userID := uint(1)
	email := "test@example.com"
	role := "user"

	refreshToken, _ := GenerateRefreshToken(userID, email, role)

	newPair, err := RefreshAccessToken(refreshToken)
	if err != nil {
		t.Fatalf("RefreshAccessToken failed: %v", err)
	}

	if newPair.AccessToken == "" {
		t.Fatal("New access token is empty")
	}

	claims, _ := ValidateToken(newPair.AccessToken)
	if claims.UserID != userID {
		t.Errorf("UserID mismatch: got %d, want %d", claims.UserID, userID)
	}
	if claims.Type != AccessToken {
		t.Errorf("Token type should be access: got %s", claims.Type)
	}
}

func TestRefreshAccessToken_InvalidToken(t *testing.T) {
	_, err := RefreshAccessToken("invalid.token")
	if err == nil {
		t.Fatal("RefreshAccessToken should fail for invalid token")
	}
}

func TestRefreshAccessToken_WrongTokenType(t *testing.T) {
	accessToken, _ := GenerateAccessToken(1, "test@example.com", "user")

	_, err := RefreshAccessToken(accessToken)
	if err == nil {
		t.Fatal("RefreshAccessToken should fail when given access token instead of refresh token")
	}
}

func TestRefreshAccessToken_ExpiredRefreshToken(t *testing.T) {
	os.Setenv("JWT_REFRESH_EXPIRES", "1")

	refreshToken, _ := GenerateRefreshToken(1, "test@example.com", "user")

	time.Sleep(2 * time.Second)

	_, err := RefreshAccessToken(refreshToken)
	if err == nil {
		t.Fatal("RefreshAccessToken should fail for expired token")
	}

	os.Setenv("JWT_REFRESH_EXPIRES", "604800")
}

func TestClaims_TokenTypes(t *testing.T) {
	access, _ := GenerateAccessToken(1, "a@b.com", "user")
	refresh, _ := GenerateRefreshToken(1, "a@b.com", "user")

	accessClaims, _ := ValidateToken(access)
	refreshClaims, _ := ValidateToken(refresh)

	if accessClaims.Type == refreshClaims.Type {
		t.Fatal("Access and refresh tokens should have different types")
	}

	if accessClaims.Type != AccessToken {
		t.Error("Access token should have type AccessToken")
	}
	if refreshClaims.Type != RefreshToken {
		t.Error("Refresh token should have type RefreshToken")
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	_, err := ValidateToken("")
	if err == nil {
		t.Fatal("ValidateToken should fail for empty token")
	}
}

func TestValidateToken_MalformedToken(t *testing.T) {
	testCases := []string{
		"not.a.jwt",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		"abc.def",
	}

	for _, tc := range testCases {
		_, err := ValidateToken(tc)
		if err == nil {
			t.Errorf("ValidateToken should fail for malformed token: %s", tc)
		}
	}
}
