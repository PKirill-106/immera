package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	GetUserSettings(ctx context.Context, id uuid.UUID) (UserSettings, error)
	UpdateUser(ctx context.Context, id uuid.UUID, user UpdateUserParams) error
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool: pool,
	}
}

func (r *PostgresRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (User, error) {
	var user User

	err := r.pool.QueryRow(
		ctx,
		`
    SELECT
        id,
        name,
        email,
        phone_number,
			  created_at,
			  updated_at
    FROM users
    WHERE id = $1
    `,
		id,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PhoneNumber,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}

		return User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (r *PostgresRepository) GetUserSettings(
	ctx context.Context,
	id uuid.UUID,
) (UserSettings, error) {
	var (
		userID          uuid.UUID
		settingsID      *uuid.UUID
		defaultLanguage *string
		theme           *string
	)

	err := r.pool.QueryRow(
		ctx,
		`
		SELECT
    		u.id,
				us.id,
				us.default_language,
				us.theme
		FROM users u
		LEFT JOIN user_settings us ON us.user_id = u.id
		WHERE u.id = $1;
		`,
		id,
	).Scan(
		&userID,
		&settingsID,
		&defaultLanguage,
		&theme,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserSettings{}, ErrUserNotFound
		}

		return UserSettings{}, fmt.Errorf("get user settings by id: %w", err)
	}

	if settingsID == nil || defaultLanguage == nil || theme == nil {
		return UserSettings{}, fmt.Errorf("user settings invariant violated for user %s", userID)
	}

	userSettings := UserSettings{
		ID:              *settingsID,
		DefaultLanguage: *defaultLanguage,
		Theme:           *theme,
	}

	return userSettings, nil
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, id uuid.UUID, user UpdateUserParams) error {
	tag, err := r.pool.Exec(
		ctx,
		`
		UPDATE users
		SET 
			name = $1,
			email = $2,
			phone_number = $3,
			updated_at = now()
		WHERE id = $4
		`,
		user.Name,
		user.Email,
		user.PhoneNumber,
		id,
	)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			switch postgresError.ConstraintName {
			case "users_email_key":
				return ErrEmailAlreadyExists
			case "users_phone_number_key":
				return ErrPhoneNumberAlreadyExists
			default:
				return ErrUserConflict
			}
		}

		return fmt.Errorf("update user by id: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}
