package store

import (
	"context"
	"github.com/Brain-Wave-Ecosystem/go-common/pkg/dbx"
	apperrors "github.com/Brain-Wave-Ecosystem/go-common/pkg/error"
	"github.com/Brain-Wave-Ecosystem/users-service/internal/models"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"time"
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	builder := dbx.StatementBuilder.
		Select("id", "email", "avatar_url", "full_name", "slug", "bio", "last_login_at", "role", "created_at", "updated_at").
		From("users").
		Where(squirrel.Eq{"id": userID}).
		Where(squirrel.Eq{"deleted_at": nil})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var user models.User
	err = s.db.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.AvatarURL,
		&user.FullName,
		&user.Slug,
		&user.Bio,
		&user.LastLoginAt,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	switch {
	case dbx.IsNoRows(err):
		return nil, apperrors.NotFound("user", "id", userID)
	case err != nil:
		return nil, apperrors.Internal(err)
	}

	return &user, nil
}

func (s *Store) GetUserBySlug(ctx context.Context, slug string) (*models.User, error) {
	builder := dbx.StatementBuilder.
		Select("id", "email", "avatar_url", "full_name", "slug", "bio", "last_login_at", "role", "created_at", "updated_at").
		From("users").
		Where(squirrel.Eq{"slug": slug}).
		Where(squirrel.Eq{"deleted_at": nil})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var user models.User
	err = s.db.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.AvatarURL,
		&user.FullName,
		&user.Slug,
		&user.Bio,
		&user.LastLoginAt,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	switch {
	case dbx.IsNoRows(err):
		return nil, apperrors.NotFound("user", "slug", slug)
	case err != nil:
		return nil, apperrors.Internal(err)
	}

	return &user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*models.UserWithPassword, error) {
	builder := dbx.StatementBuilder.
		Select("id", "email", "avatar_url", "full_name", "slug", "bio", "role", "pass_hash", "is_verified", "updated_at", "created_at").
		From("users").
		Where(squirrel.Eq{"email": email}).
		Where(squirrel.Eq{"deleted_at": nil})

	loginAt := time.Now()
	loginAtBuilder := dbx.StatementBuilder.
		Update("users").
		Set("last_login_at", loginAt).
		Where(squirrel.Eq{"email": email}).
		Where(squirrel.Eq{"deleted_at": nil})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, apperrors.Internal(err)
	}

	loginAtQuery, loginAtArgs, loginAtErr := loginAtBuilder.ToSql()
	if loginAtErr != nil {
		return nil, apperrors.Internal(loginAtErr)
	}

	var user models.UserWithPassword
	user.User = &models.User{}
	user.UserPassword = &models.UserPassword{}
	user.LastLoginAt = &loginAt
	err = s.db.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.AvatarURL,
		&user.FullName,
		&user.Slug,
		&user.Bio,
		&user.Role,
		&user.PasswordHash,
		&user.IsVerified,
		&user.User.UpdatedAt,
		&user.CreatedAt,
	)

	cmd, loginTimeErr := s.db.Exec(ctx, loginAtQuery, loginAtArgs...)

	switch {
	case dbx.IsNoRows(err) && cmd.RowsAffected() == 0:
		return nil, apperrors.NotFound("user", "email", email)
	case err != nil:
		return nil, apperrors.Internal(err)
	case loginTimeErr != nil:
		return nil, apperrors.Internal(loginTimeErr)
	}

	return &user, nil
}

func (s *Store) CreateUser(ctx context.Context, user *models.UserWithPassword) (*models.User, error) {
	builder := dbx.StatementBuilder.
		Insert("users").
		Columns("full_name", "slug", "email", "pass_hash").
		Values(user.FullName, user.Slug, user.Email, user.PasswordHash).
		Suffix("RETURNING id, role, created_at")

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	tracer := otel.Tracer("pgxpool")
	ctx, span := tracer.Start(ctx, "CreateUser",
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.statement", query),
		),
	)
	defer span.End()

	err = s.db.QueryRow(ctx, query, args...).Scan(&user.ID, &user.Role, &user.CreatedAt)

	switch {
	case dbx.IsUniqueViolation(err, "email"):
		span.RecordError(err)
		return nil, apperrors.AlreadyExists("user", "email", user.Email)
	case dbx.IsUniqueViolation(err, "slug"):
		return nil, apperrors.AlreadyExists("user", "slug", user.Slug)
	case err != nil:
		span.RecordError(err)
		return nil, apperrors.Internal(err)
	}

	return user.User, nil
}

func (s *Store) ConfirmUser(ctx context.Context, userID int64) error {
	builder := dbx.StatementBuilder.
		Update("users").
		Set("role", "user").
		Set("is_verified", true).
		Where(squirrel.Eq{"id": userID}).
		Where(squirrel.Eq{"deleted_at": nil})

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	cmd, err := s.db.Exec(ctx, query, args...)

	switch {
	case cmd.RowsAffected() == 0:
		return apperrors.NotFound("user", "id", userID)
	case err != nil:
		return apperrors.Internal(err)
	}

	return nil
}

func (s *Store) UpdateUser(ctx context.Context, userID int, data *models.UpdateUser) error {
	builder := dbx.StatementBuilder.
		Update("users").
		Where(squirrel.Eq{"id": userID}).
		Where(squirrel.Eq{"deleted_at": nil})

	hasSet := false
	if data.AvatarURL != nil {
		builder = builder.Set("avatar_url", *data.AvatarURL)
		hasSet = true
	}
	if data.FullName != nil {
		builder = builder.Set("full_name", *data.FullName)
		builder = builder.Set("slug", data.Slug)
		hasSet = true
	}
	if data.Bio != nil {
		builder = builder.Set("bio", *data.Bio)
		hasSet = true
	}

	if !hasSet {
		builder = builder.Set("id", userID)
	} else {
		builder = builder.Set("updated_at", time.Now())
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	cmd, err := s.db.Exec(ctx, query, args...)

	switch {
	case cmd.RowsAffected() == 0:
		return apperrors.NotFound("user", "id", userID)
	case err != nil:
		return apperrors.Internal(err)
	}

	return nil
}

func (s *Store) UpdatePassword(ctx context.Context, userID int, password string) error {
	builder := dbx.StatementBuilder.
		Update("users").
		Set("pass_hash", password).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": userID})

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	cmd, err := s.db.Exec(ctx, query, args...)

	switch {
	case cmd.RowsAffected() == 0:
		return apperrors.NotFound("user", "id", userID)
	case err != nil:
		return apperrors.Internal(err)
	}

	return nil
}

func (s *Store) DeleteUser(ctx context.Context, userID int) error {
	builder := dbx.StatementBuilder.
		Update("users").
		Set("deleted_at", time.Now()).
		Where(squirrel.Eq{"id": userID})

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	cmd, err := s.db.Exec(ctx, query, args...)

	switch {
	case cmd.RowsAffected() == 0:
		return apperrors.NotFound("user", "id", userID)
	case err != nil:
		return apperrors.Internal(err)
	}

	return nil
}
