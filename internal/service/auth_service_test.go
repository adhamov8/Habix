package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// --- helpers ---

func newTestAuthService(users *mockUserRepo, tokens *mockTokenRepo) *AuthService {
	return NewAuthService(users, tokens, "test-secret-key")
}

// --- Register tests ---

func TestRegister_Success(t *testing.T) {
	users := newMockUserRepo()
	tokens := newMockTokenRepo()
	svc := newTestAuthService(users, tokens)

	pair, err := svc.Register(context.Background(), "alice@test.com", "password123", "Alice")
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)

	assert.Len(t, users.users, 1)
	u := users.users["alice@test.com"]
	assert.Equal(t, "Alice", u.Name)
	assert.Equal(t, "UTC", u.Timezone)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("password123")))

	assert.Len(t, tokens.tokens, 1)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	users := newMockUserRepo()
	tokens := newMockTokenRepo()
	svc := newTestAuthService(users, tokens)

	_, err := svc.Register(context.Background(), "alice@test.com", "pass1", "Alice")
	require.NoError(t, err)

	_, err = svc.Register(context.Background(), "alice@test.com", "pass2", "Alice2")
	assert.ErrorIs(t, err, ErrEmailTaken)
}

// --- Login tests ---

func TestLogin_Success(t *testing.T) {
	users := newMockUserRepo()
	tokens := newMockTokenRepo()
	svc := newTestAuthService(users, tokens)

	_, err := svc.Register(context.Background(), "bob@test.com", "secret", "Bob")
	require.NoError(t, err)

	pair, err := svc.Login(context.Background(), "bob@test.com", "secret")
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
}

func TestLogin_WrongPassword(t *testing.T) {
	users := newMockUserRepo()
	tokens := newMockTokenRepo()
	svc := newTestAuthService(users, tokens)

	_, err := svc.Register(context.Background(), "bob@test.com", "secret", "Bob")
	require.NoError(t, err)

	_, err = svc.Login(context.Background(), "bob@test.com", "wrong")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_UserNotFound(t *testing.T) {
	users := newMockUserRepo()
	tokens := newMockTokenRepo()
	svc := newTestAuthService(users, tokens)

	_, err := svc.Login(context.Background(), "nobody@test.com", "pass")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

// --- ParseAccessToken tests ---

func TestParseAccessToken_Valid(t *testing.T) {
	users := newMockUserRepo()
	tokens := newMockTokenRepo()
	svc := newTestAuthService(users, tokens)

	pair, err := svc.Register(context.Background(), "eve@test.com", "pass", "Eve")
	require.NoError(t, err)

	userID, err := svc.ParseAccessToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, users.users["eve@test.com"].ID, userID)
}

func TestParseAccessToken_Invalid(t *testing.T) {
	svc := newTestAuthService(newMockUserRepo(), newMockTokenRepo())

	_, err := svc.ParseAccessToken("garbage")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestParseAccessToken_WrongSecret(t *testing.T) {
	users := newMockUserRepo()
	tokens := newMockTokenRepo()
	svc1 := NewAuthService(users, tokens, "secret-1")
	svc2 := NewAuthService(users, tokens, "secret-2")

	pair, err := svc1.Register(context.Background(), "eve@test.com", "pass", "Eve")
	require.NoError(t, err)

	_, err = svc2.ParseAccessToken(pair.AccessToken)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

// --- Refresh tests ---

func TestRefresh_Success(t *testing.T) {
	users := newMockUserRepo()
	tokens := newMockTokenRepo()
	svc := newTestAuthService(users, tokens)

	pair, err := svc.Register(context.Background(), "carol@test.com", "pass", "Carol")
	require.NoError(t, err)

	newPair, err := svc.Refresh(context.Background(), pair.RefreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newPair.AccessToken)
	assert.NotEmpty(t, newPair.RefreshToken)
	assert.NotEqual(t, pair.RefreshToken, newPair.RefreshToken)
}

func TestRefresh_InvalidToken(t *testing.T) {
	svc := newTestAuthService(newMockUserRepo(), newMockTokenRepo())

	_, err := svc.Refresh(context.Background(), "bad-token")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

// --- Logout tests ---

func TestLogout_DeletesToken(t *testing.T) {
	users := newMockUserRepo()
	tokens := newMockTokenRepo()
	svc := newTestAuthService(users, tokens)

	pair, err := svc.Register(context.Background(), "dan@test.com", "pass", "Dan")
	require.NoError(t, err)
	assert.Len(t, tokens.tokens, 1)

	err = svc.Logout(context.Background(), pair.RefreshToken)
	require.NoError(t, err)
	assert.Len(t, tokens.tokens, 0)
}