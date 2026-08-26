package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	RegisterUser(ctx context.Context, newUser RegisterUserParams) error
	// LoginUser(ctx context.Context, email string, passwordHash string) (string, error)
	// LogoutUser(ctx context.Context, id uuid.UUID) error
	// RefreshToken(ctx context.Context, token string) (string, error)
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool: pool,
	}
}

func (r *PostgresRepository) RegisterUser(ctx context.Context, newUser RegisterUserParams) error {
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
		newUser.Name,
		newUser.Email,
		newUser.PhoneNumber,
		newUser.PasswordHash,
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

// func (r *PostgresRepository) LoginUser(ctx context.Context, email string, passwordHash string) (string, error) {
// }

// func (r *PostgresRepository) LogoutUser(ctx context.Context, id uuid.UUID) error {}

// func (r *PostgresRepository) RefreshToken(ctx context.Context, token string) (string, error) {}
