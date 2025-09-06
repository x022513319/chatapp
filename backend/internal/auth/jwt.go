package auth

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
}

func SignAccessToken(userID int64) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET is empty")
	}

	ttlMin := 30
	if v := os.Getenv("ACCESS_TOKEN_TTL_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			ttlMin = n
		}
	}

	now := time.Now()
	sub := strconv.FormatInt(userID, 10)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			ID:        uuid.New().String(), // jti
			Issuer:    "chatapp",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttlMin) * time.Minute)),
			Audience:  []string{"web"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))

	if err != nil {
		return "", err
	}

	return signed, nil
}

func ParseAndValidate(tokenStr string) (*Claims, error) {
	secret := os.Getenv("JWT_SECRET")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok || t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		},

		jwt.WithAudience("web"),
		jwt.WithIssuer("chatapp"),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),

		jwt.WithLeeway(10*time.Second),
	)

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
