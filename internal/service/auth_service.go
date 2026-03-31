package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"tracker/internal/domain"
	"tracker/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrEmailTaken         = errors.New("email already in use")
)

type AuthService struct {
	users     UserRepo
	tokens    TokenRepo
	jwtSecret []byte
}

func NewAuthService(users UserRepo, tokens TokenRepo, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		tokens:    tokens,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (domain.TokenPair, error) {
	existing, err := s.users.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return domain.TokenPair{}, err
	}
	if existing != nil {
		return domain.TokenPair{}, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.TokenPair{}, err
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Timezone:     "UTC",
	}
	if err := s.users.Create(ctx, user); err != nil {
		return domain.TokenPair{}, err
	}
	return s.issueTokenPair(ctx, user.ID)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (domain.TokenPair, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.TokenPair{}, ErrInvalidCredentials
		}
		return domain.TokenPair{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return domain.TokenPair{}, ErrInvalidCredentials
	}
	return s.issueTokenPair(ctx, user.ID)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error) {
	hash := hashToken(refreshToken)
	stored, err := s.tokens.GetByHash(ctx, hash)
	if err != nil {
		return domain.TokenPair{}, ErrInvalidToken
	}
	if err := s.tokens.Delete(ctx, hash); err != nil {
		return domain.TokenPair{}, err
	}
	return s.issueTokenPair(ctx, stored.UserID)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.tokens.Delete(ctx, hashToken(refreshToken))
}

func (s *AuthService) ParseAccessToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return uuid.Nil, ErrInvalidToken
	}
	return uuid.Parse(claims.UserID)
}

func (s *AuthService) issueTokenPair(ctx context.Context, userID uuid.UUID) (domain.TokenPair, error) {
	accessToken, err := s.signAccessToken(userID)
	if err != nil {
		return domain.TokenPair{}, err
	}

	raw, err := generateRawToken()
	if err != nil {
		return domain.TokenPair{}, err
	}

	rt := &repository.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hashToken(raw),
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
	}
	if err := s.tokens.Create(ctx, rt); err != nil {
		return domain.TokenPair{}, err
	}

	return domain.TokenPair{AccessToken: accessToken, RefreshToken: raw}, nil
}

type jwtClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *AuthService) signAccessToken(userID uuid.UUID) (string, error) {
	claims := jwtClaims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}

func generateRawToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}