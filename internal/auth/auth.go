package auth

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessIssuer  = "chirpy-access"
	RefreshIssuer = "chirpy-refresh"
)

func ValidateToken(r *http.Request) (int, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, errors.New("No authorization header")
	}

	tokenString := authHeader[len("Bearer "):]
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("Invalid token")
	}

	userIdString := claims.Subject
	issuer := claims.Issuer

	if issuer != AccessIssuer {
		return 0, errors.New("Invalid token")
	}
	// convert string to int
	return strconv.Atoi(userIdString)
}

func ValidateRefreshToken(authHeader string) (int, string, error) {
	tokenString := authHeader[len("Bearer "):]
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return 0, "", err
	}

	if !token.Valid {
		return 0, "", errors.New("Invalid token")
	}

	userIdString := claims.Subject
	issuer := claims.Issuer

	if issuer != RefreshIssuer {
		return 0, "", errors.New("Invalid token")
	}

	// convert string to int
	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		return 0, "", err
	}

	return userId, tokenString, nil
}

func GetAccessToken(userId int) (string, error) {
	expiry := time.Duration(1 * time.Hour)
	claims := jwt.RegisteredClaims{
		Issuer:    AccessIssuer,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		Subject:   strconv.Itoa(userId), // convert int to string
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiry)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	return token.SignedString([]byte(secret))

}

func GetRefreshToken(userId int) (string, error) {
	expiry := time.Duration(60 * 24 * time.Hour)
	claims := jwt.RegisteredClaims{
		Issuer:    RefreshIssuer,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		Subject:   strconv.Itoa(userId), // convert int to string
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiry)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	return token.SignedString([]byte(secret))
}
