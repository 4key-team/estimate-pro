package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewService(secret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *Service) GeneratePair(userID string) (*TokenPair, error) {
	access, err := s.generateToken(userID, s.accessTTL, "access")
	if err != nil {
		return nil, fmt.Errorf("jwt.GeneratePair access: %w", err)
	}
	refresh, err := s.generateToken(userID, s.refreshTTL, "refresh")
	if err != nil {
		return nil, fmt.Errorf("jwt.GeneratePair refresh: %w", err)
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *Service) ValidateAccess(tokenStr string) (*Claims, error) {
	return s.validate(tokenStr, "access")
}

func (s *Service) ValidateRefresh(tokenStr string) (*Claims, error) {
	return s.validate(tokenStr, "refresh")
}

func (s *Service) generateToken(userID string, ttl time.Duration, subject string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *Service) validate(tokenStr, expectedSubject string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt.validate: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("jwt.validate: invalid token")
	}
	if claims.Subject != expectedSubject {
		return nil, fmt.Errorf("jwt.validate: unexpected token type %q", claims.Subject)
	}
	return claims, nil
}
