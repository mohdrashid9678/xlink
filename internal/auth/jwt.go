package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	accessTokenTTL = time.Hour
	jwtIssuer      = "xlink"
)

type Claims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	signingKey []byte
}

func NewJWTManager(signingKey string) *JWTManager {
	return &JWTManager{signingKey: []byte(signingKey)}
}

func (m *JWTManager) Issue(userID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtIssuer,
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.signingKey)
}

func (m *JWTManager) Verify(tokenString string) (uuid.UUID, error) {
	claims := new(Claims)
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(jwtIssuer),
		jwt.WithExpirationRequired(),
	)
	token, err := parser.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return m.signingKey, nil
	})
	if err != nil || !token.Valid || claims.Subject != claims.UserID {
		return uuid.Nil, fmt.Errorf("invalid access token")
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid access token")
	}
	return userID, nil
}

func AccessTokenTTL() time.Duration { return accessTokenTTL }
