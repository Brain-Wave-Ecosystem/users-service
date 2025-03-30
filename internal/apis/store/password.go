package store

import (
	"context"
	"fmt"
	"github.com/Brain-Wave-Ecosystem/go-common/pkg/dbx"
	apperrors "github.com/Brain-Wave-Ecosystem/go-common/pkg/error"
	"github.com/Brain-Wave-Ecosystem/users-service/internal/models"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

func (s *Store) AddPasswordHistory(ctx context.Context, userID int64, passwordHash string) error {
	builder := dbx.StatementBuilder.
		Insert("users_password_history").
		Columns("user_id", "pass_hash").
		Values(userID, passwordHash)

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	cmd, err := s.db.Exec(ctx, query, args...)

	switch {
	case cmd.RowsAffected() == 0:
		return apperrors.BadRequest(fmt.Errorf("no rows affected user id: %d", userID))
	case err != nil:
		return apperrors.InternalWithoutStackTrace(err)
	}

	return nil
}

func (s *Store) GetPasswordHistory(ctx context.Context, userID int64) ([]*models.UserPasswordHistory, error) {
	builder := dbx.StatementBuilder.
		Select("pass_hash", "created_at").
		From("users_password_history").
		Where(squirrel.Eq{"user_id": userID})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, apperrors.Internal(err)
	}

	defer rows.Close()

	histories, err := pgx.CollectRows[*models.UserPasswordHistory](rows, pgx.RowToAddrOfStructByPos)
	if err != nil {
		return nil, apperrors.Internal(err)
	}

	return histories, nil
}
