package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	types "github.com/pedro/10db-launch/apps/server/internal/types"
)

type Service struct {
	store *Store
}

func New(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) Register(ctx context.Context, name, email, password string) (types.User, error) {
	normalizedName := strings.TrimSpace(name)
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	if normalizedName == "" {
		return types.User{}, errors.New("name is required")
	}
	if normalizedEmail == "" {
		return types.User{}, errors.New("email is required")
	}
	if len(password) < 8 {
		return types.User{}, errors.New("password must be at least 8 characters")
	}

	if _, err := s.store.GetByEmail(ctx, normalizedEmail); err == nil {
		return types.User{}, errors.New("an account with that email already exists")
	} else if !errors.Is(err, sql.ErrNoRows) {
		return types.User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return types.User{}, err
	}

	now := time.Now().UTC()
	user := types.User{
		ID:           uuid.NewString(),
		Email:        normalizedEmail,
		Name:         normalizedName,
		Role:         types.UserRoleUser,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.store.Create(ctx, user); err != nil {
		return types.User{}, err
	}
	return user, nil
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (types.User, error) {
	user, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return types.User{}, errors.New("invalid email or password")
		}
		return types.User{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return types.User{}, errors.New("invalid email or password")
	}
	return user, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (types.User, error) {
	return s.store.GetByID(ctx, id)
}
