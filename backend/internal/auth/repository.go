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
	GetRefreshSessionByTokenHash(ctx context.Context, tokenHash string) (RefreshSession, error)
	RotateRefreshSession(
		ctx context.Context,
		oldTokenHash string,
		userID uuid.UUID,
		newTokenHash string,
		newExpiresAt time.Time,
	) error
	DeleteRefreshSessionByTokenHash(ctx context.Context, tokenHash string) error
	GetEmailVerificationByTokenHash(ctx context.Context, tokenHash string) (EmailVerification, error)
	VerifyEmail(
		ctx context.Context,
		verificationID uuid.UUID,
		userID uuid.UUID,
		verifiedAt time.Time,
	) error
	GetUserVerificationStatusByEmail(ctx context.Context, email string) (UserVerificationStatus, error)
	ReplaceEmailVerificationToken(
		ctx context.Context,
		userID uuid.UUID,
		tokenHash string,
		expiresAt time.Time,
	) error
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

	tag, err = tx.Exec(
		ctx,
		`
		INSERT INTO email_verification_tokens (
			user_id,
			token_hash,
			expires_at
		)
		VALUES ($1, $2, $3)
		`,
		userID,
		registration.VerificationTokenHash,
		registration.VerificationTokenExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert email verification token: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return errors.New("insert email verification token: unexpected row count")
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

func (r *PostgresRepository) GetRefreshSessionByTokenHash(
	ctx context.Context,
	tokenHash string,
) (RefreshSession, error) {
	var session RefreshSession

	err := r.pool.QueryRow(
		ctx,
		`
		SELECT user_id, expires_at
		FROM auth_refresh_tokens
		WHERE token_hash = $1
		`,
		tokenHash,
	).Scan(&session.UserID, &session.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RefreshSession{}, ErrRefreshTokenNotFound
		}

		return RefreshSession{}, fmt.Errorf("get refresh session by token hash: %w", err)
	}

	return session, nil
}

func (r *PostgresRepository) RotateRefreshSession(
	ctx context.Context,
	oldTokenHash string,
	userID uuid.UUID,
	newTokenHash string,
	newExpiresAt time.Time,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin refresh rotation transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	tag, err := tx.Exec(
		ctx,
		`DELETE FROM auth_refresh_tokens WHERE token_hash = $1`,
		oldTokenHash,
	)
	if err != nil {
		return fmt.Errorf("delete rotated refresh session: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return ErrRefreshTokenNotFound
	}

	tag, err = tx.Exec(
		ctx,
		`
		INSERT INTO auth_refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		`,
		userID,
		newTokenHash,
		newExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert rotated refresh session: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return errors.New("insert rotated refresh session: unexpected row count")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit refresh rotation transaction: %w", err)
	}

	return nil
}

func (r *PostgresRepository) DeleteRefreshSessionByTokenHash(
	ctx context.Context,
	tokenHash string,
) error {
	_, err := r.pool.Exec(
		ctx,
		`DELETE FROM auth_refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("delete refresh session by token hash: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetEmailVerificationByTokenHash(
	ctx context.Context,
	tokenHash string,
) (EmailVerification, error) {
	var verification EmailVerification

	err := r.pool.QueryRow(
		ctx,
		`
		SELECT id, user_id, expires_at, used_at
		FROM email_verification_tokens
		WHERE token_hash = $1
		`,
		tokenHash,
	).Scan(
		&verification.ID,
		&verification.UserID,
		&verification.ExpiresAt,
		&verification.UsedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EmailVerification{}, ErrVerificationTokenNotFound
		}

		return EmailVerification{}, fmt.Errorf("get email verification by token hash: %w", err)
	}

	return verification, nil
}

func (r *PostgresRepository) VerifyEmail(
	ctx context.Context,
	verificationID uuid.UUID,
	userID uuid.UUID,
	verifiedAt time.Time,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin email verification transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	tag, err := tx.Exec(
		ctx,
		`
		UPDATE email_verification_tokens
		SET used_at = $2
		WHERE id = $1 AND used_at IS NULL
		`,
		verificationID,
		verifiedAt,
	)
	if err != nil {
		return fmt.Errorf("mark email verification token used: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return ErrVerificationTokenUsed
	}

	tag, err = tx.Exec(
		ctx,
		`
		UPDATE users
		SET email_verified_at = $2, updated_at = $2
		WHERE id = $1 AND email_verified_at IS NULL
		`,
		userID,
		verifiedAt,
	)
	if err != nil {
		return fmt.Errorf("mark user email verified: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return ErrEmailAlreadyVerified
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit email verification transaction: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetUserVerificationStatusByEmail(
	ctx context.Context,
	email string,
) (UserVerificationStatus, error) {
	var status UserVerificationStatus

	err := r.pool.QueryRow(
		ctx,
		`
		SELECT id, email, email_verified_at
		FROM users
		WHERE email = $1
		`,
		email,
	).Scan(&status.ID, &status.Email, &status.EmailVerifiedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserVerificationStatus{}, ErrUserNotFound
		}

		return UserVerificationStatus{}, fmt.Errorf("get user verification status by email: %w", err)
	}

	return status, nil
}

func (r *PostgresRepository) ReplaceEmailVerificationToken(
	ctx context.Context,
	userID uuid.UUID,
	tokenHash string,
	expiresAt time.Time,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin replace verification token transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(
		ctx,
		`DELETE FROM email_verification_tokens WHERE user_id = $1 AND used_at IS NULL`,
		userID,
	); err != nil {
		return fmt.Errorf("delete active verification tokens: %w", err)
	}

	tag, err := tx.Exec(
		ctx,
		`
		INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		`,
		userID,
		tokenHash,
		expiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert replacement verification token: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return errors.New("insert replacement verification token: unexpected row count")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit replace verification token transaction: %w", err)
	}

	return nil
}
