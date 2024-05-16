package token

import (
	"github.com/dgrijalva/jwt-go"
	"job_tasks/calendar/config"
	"time"
)

// CustomClaims represents custom claims for JWT
type CustomClaims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.StandardClaims
}

func GenerateTokens(email, role string) (accessToken string, refreshToken string, err error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return "", "", err
	}
	accessToken, err = generateToken(email, role, []byte(cfg.Token.Secret), time.Minute*15)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = generateToken(email, role, []byte(cfg.Token.Secret), time.Hour*24*7)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func generateToken(email, role string, secretKey []byte, expiry time.Duration) (string, error) {
	// Define custom claims
	claims := CustomClaims{
		Email: email,
		Role:  role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiry).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "nodirbek",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
