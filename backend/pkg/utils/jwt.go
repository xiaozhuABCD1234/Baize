package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	AccessToken  TokenType = "access_token"
	RefreshToken TokenType = "refresh_token"
)

type Claims struct {
	UserID   uint      `json:"user_id"`
	Email    string    `json:"email"`
	UserType string    `json:"user_type"`
	Type     TokenType `json:"type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func GenerateAccessToken(userID uint, email string, userType string) (string, error) {
	expiresIn := GetEnvInt("JWT_ACCESS_EXPIRES", 15*60)
	claims := Claims{
		UserID:   userID,
		Email:    email,
		UserType: userType,
		Type:     AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiresIn) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(GetEnv("JWT_SECRET_KEY", "default_secret")))
}

func GenerateRefreshToken(userID uint, email string, userType string) (string, error) {
	expiresIn := GetEnvInt("JWT_REFRESH_EXPIRES", 7*24*60*60)
	claims := Claims{
		UserID:   userID,
		Email:    email,
		UserType: userType,
		Type:     RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiresIn) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(GetEnv("JWT_SECRET_KEY", "default_secret")))
}

func GenerateTokenPair(userID uint, email string, userType string) (*TokenPair, error) {
	accessToken, err := GenerateAccessToken(userID, email, userType)
	if err != nil {
		return nil, err
	}
	refreshToken, err := GenerateRefreshToken(userID, email, userType)
	if err != nil {
		return nil, err
	}
	expiresIn := GetEnvInt("JWT_ACCESS_EXPIRES", 15*60)
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(expiresIn),
	}, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(GetEnv("JWT_SECRET_KEY", "default_secret")), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

func RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
	claims, err := ValidateToken(refreshTokenString)
	if err != nil {
		return nil, err
	}
	if claims.Type != RefreshToken {
		return nil, jwt.ErrSignatureInvalid
	}
	return GenerateTokenPair(claims.UserID, claims.Email, claims.UserType)
}
