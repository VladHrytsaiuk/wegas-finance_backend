package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type JWTClaims struct {
	UserID   string `json:"user_id"`
	FamilyID string `json:"family_id"`
	RoleID   string `json:"role_id"`
	jwt.RegisteredClaims
}

type JWTService interface {
	GenerateAccessToken(userID, familyID, roleID string) (string, error)
	GenerateRefreshToken(userID, familyID, roleID string) (string, error)
	ValidateToken(tokenStr string) (*JWTClaims, error)
}

type jwtService struct {
	secretKey []byte
}

func NewJWTService(secretKey string) JWTService {
	return &jwtService{secretKey: []byte(secretKey)}
}

func (s *jwtService) GenerateAccessToken(userID, familyID, roleID string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		FamilyID: familyID,
		RoleID:   roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *jwtService) GenerateRefreshToken(userID, familyID, roleID string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		FamilyID: familyID,
		RoleID:   roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *jwtService) ValidateToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
