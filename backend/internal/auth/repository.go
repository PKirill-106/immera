package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Register(ctx context.Context, registration RegisterParams) error
	GetCredentialsByEmail(ctx context.Context, email string) (UserCredentials, error)
	CreateRefreshSession(
		ctx context.Context,
		id uuid.UUID,
		refreshTokenHash string,
		refreshExpiresAt time.Time,
	) error
	// Logout(ctx context.Context, id uuid.UUID) error
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool: pool,
	}
}

func (r *PostgresRepository) Register(ctx context.Context, registration RegisterParams) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin registration transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var userID uuid.UUID

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO users (
			name,
			email,
			phone_number,
			password_hash
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id
		`,
		registration.Name,
		registration.Email,
		registration.PhoneNumber,
		registration.PasswordHash,
	).Scan(&userID)
	if err != nil {
		return fmt.Errorf("insert registered user: %w", err)
	}

	tag, err := tx.Exec(
		ctx,
		`
		INSERT INTO user_settings (user_id)
		VALUES ($1)
		`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("insert default user settings: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return errors.New("insert default user settings: unexpected row count")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit registration transaction: %w", err)
	}

	return nil

}

func (r *PostgresRepository) GetCredentialsByEmail(ctx context.Context, email string) (UserCredentials, error) {
	var userCredentials UserCredentials

	err := r.pool.QueryRow(
		ctx,
		`
    SELECT id, email, password_hash
    FROM users
    WHERE email = $1
    `,
		email,
	).Scan(
		&userCredentials.ID,
		&userCredentials.Email,
		&userCredentials.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserCredentials{}, ErrUserNotFound
		}

		return UserCredentials{}, fmt.Errorf("get user credentials by email: %w", err)
	}

	return userCredentials, nil
}

func (r *PostgresRepository) CreateRefreshSession(
	ctx context.Context,
	id uuid.UUID,
	refreshTokenHash string,
	refreshExpiresAt time.Time,
) error {
	tag, err := r.pool.Exec(
		ctx,
		`
    INSERT into auth_refresh_tokens(
      user_id,
      token_hash,
      expires_at
    )
    VALUES($1, $2, $3)
    `,
		id,
		refreshTokenHash,
		refreshExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("create refresh session: %w", err)
	}

	if tag.RowsAffected() != 1 {
    return errors.New("insert refresh session: unexpected row count")
	}

	return nil
}

// func (r *PostgresRepository) Logout(ctx context.Context, id uuid.UUID) error {}

// func (r *PostgresRepository) RefreshToken(ctx context.Context, token string) (string, error) {}
