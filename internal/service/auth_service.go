package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/auth"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/mohdrashid9678/xlink/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

const refreshTokenTTL = 10 * 24 * time.Hour

type AuthService interface {
	Register(context.Context, models.RegisterRequest) (*models.AuthResponse, error)
	Login(context.Context, models.LoginRequest) (*models.AuthResponse, error)
	Refresh(context.Context, string) (*models.TokenPair, error)
	Logout(context.Context, string) error
}

type DefaultAuthService struct {
	users   repository.UserRepository
	refresh repository.RefreshTokenRepository
	jwt     *auth.JWTManager
}

func NewAuthService(users repository.UserRepository, refresh repository.RefreshTokenRepository, jwt *auth.JWTManager) *DefaultAuthService {
	return &DefaultAuthService{users: users, refresh: refresh, jwt: jwt}
}

func (s *DefaultAuthService) Register(ctx context.Context, request models.RegisterRequest) (*models.AuthResponse, error) {
	email, name, err := validateRegistration(request)
	if err != nil {
		return nil, err
	}
	hash, err := auth.HashPassword(request.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user, err := s.users.Create(ctx, &models.User{ID: uuid.New(), Email: email, Name: name, PasswordHash: hash})
	if errors.Is(err, repository.ErrConflict) {
		return nil, ErrConflict
	}
	if err != nil {
		return nil, err
	}
	tokens, err := s.issueTokenPair(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return &models.AuthResponse{User: *user, Tokens: *tokens}, nil
}

func (s *DefaultAuthService) Login(ctx context.Context, request models.LoginRequest) (*models.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(request.Email))
	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if auth.ComparePassword(user.PasswordHash, request.Password) != nil {
		return nil, ErrInvalidCredentials
	}
	tokens, err := s.issueTokenPair(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return &models.AuthResponse{User: *user, Tokens: *tokens}, nil
}

func (s *DefaultAuthService) Refresh(ctx context.Context, rawToken string) (*models.TokenPair, error) {
	if strings.TrimSpace(rawToken) == "" {
		return nil, ErrValidation
	}
	newRawToken, err := auth.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	userID, err := s.refresh.Rotate(ctx, auth.HashRefreshToken(rawToken), &models.RefreshToken{TokenHash: auth.HashRefreshToken(newRawToken), ExpiresAt: time.Now().UTC().Add(refreshTokenTTL)})
	if errors.Is(err, repository.ErrInvalidRefreshToken) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	accessToken, err := s.jwt.Issue(userID)
	if err != nil {
		return nil, err
	}
	return &models.TokenPair{AccessToken: accessToken, RefreshToken: newRawToken, TokenType: "Bearer", ExpiresIn: int64(auth.AccessTokenTTL().Seconds())}, nil
}

func (s *DefaultAuthService) Logout(ctx context.Context, rawToken string) error {
	if strings.TrimSpace(rawToken) == "" {
		return ErrValidation
	}
	return s.refresh.Revoke(ctx, auth.HashRefreshToken(rawToken))
}

func (s *DefaultAuthService) issueTokenPair(ctx context.Context, userID uuid.UUID) (*models.TokenPair, error) {
	rawToken, err := auth.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	if err = s.refresh.Create(ctx, &models.RefreshToken{TokenHash: auth.HashRefreshToken(rawToken), UserID: userID, ExpiresAt: time.Now().UTC().Add(refreshTokenTTL)}); err != nil {
		return nil, err
	}
	accessToken, err := s.jwt.Issue(userID)
	if err != nil {
		return nil, err
	}
	return &models.TokenPair{AccessToken: accessToken, RefreshToken: rawToken, TokenType: "Bearer", ExpiresIn: int64(auth.AccessTokenTTL().Seconds())}, nil
}

func validateRegistration(request models.RegisterRequest) (string, string, error) {
	email := strings.ToLower(strings.TrimSpace(request.Email))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(email) > 254 {
		return "", "", fmt.Errorf("email is invalid: %w", ErrValidation)
	}
	name := strings.TrimSpace(request.Name)
	if name == "" || len(name) > 120 {
		return "", "", fmt.Errorf("name must contain 1-120 characters: %w", ErrValidation)
	}
	if len(request.Password) < 12 || len(request.Password) > 72 {
		return "", "", fmt.Errorf("password must contain 12-72 characters: %w", ErrValidation)
	}
	return email, name, nil
}
