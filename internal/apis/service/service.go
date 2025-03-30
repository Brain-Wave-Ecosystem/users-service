package service

import (
	"context"
	apperrors "github.com/Brain-Wave-Ecosystem/go-common/pkg/error"
	"github.com/Brain-Wave-Ecosystem/users-service/internal/apis/store"
	"github.com/Brain-Wave-Ecosystem/users-service/internal/models"
	"golang.org/x/crypto/bcrypt"
	"strconv"
)

type Service struct {
	store *store.Store
}

func NewService(store *store.Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) GetUserByIdentifier(ctx context.Context, identifier string) (*models.User, error) {
	id, ok := extractID(identifier)
	if ok {
		return s.store.GetUserByID(ctx, id)
	}
	return s.store.GetUserBySlug(ctx, identifier)
}

func (s *Service) GetUserByEmail(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, apperrors.BadRequestHidden(err, "invalid password")
	}

	return user.User, nil
}

func (s *Service) CreateUser(ctx context.Context, user *models.UserWithPassword) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.Internal(err)
	}

	user.PasswordHash = string(hash)

	newUser, err := s.store.CreateUser(ctx, user.PrepareUser())
	if err != nil {
		return nil, err
	}

	err = s.store.AddPasswordHistory(ctx, newUser.ID, user.PasswordHash)

	return newUser, err
}

func (s *Service) ConfirmUser(ctx context.Context, userID int64) error {
	return s.store.ConfirmUser(ctx, userID)
}

func (s *Service) UpdateUser(ctx context.Context, userID int64, user *models.UpdateUser) error {
	return s.store.UpdateUser(ctx, int(userID), user.PrepareUser())
}

func (s *Service) UpdateUserPassword(ctx context.Context, userID int64, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return apperrors.Internal(err)
	}

	passwordHash := string(hash)

	histories, err := s.store.GetPasswordHistory(ctx, userID)
	if err != nil {
		return apperrors.Internal(err)
	}

	for _, history := range histories {
		if err = bcrypt.CompareHashAndPassword([]byte(history.PasswordHash), []byte(password)); err == nil {
			return apperrors.BadRequestHidden(err, "this password is already used")
		}
	}

	err = s.store.UpdatePassword(ctx, int(userID), passwordHash)
	if err != nil {
		return err
	}

	return s.store.AddPasswordHistory(ctx, userID, passwordHash)
}

func (s *Service) DeleteUser(ctx context.Context, userID int64) error {
	return s.store.DeleteUser(ctx, int(userID))
}

func extractID(identifier string) (id int, ok bool) {
	var err error

	id, err = strconv.Atoi(identifier)
	if err != nil {
		return 0, false
	}
	return id, true
}
