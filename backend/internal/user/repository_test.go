package user

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapDeleteUserResult(t *testing.T) {
	t.Parallel()

	if err := mapDeleteUserResult(pgconn.NewCommandTag("DELETE 1"), nil); err != nil {
		t.Fatalf("mapDeleteUserResult() error = %v", err)
	}
}

func TestMapDeleteUserResultReturnsNotFound(t *testing.T) {
	t.Parallel()

	err := mapDeleteUserResult(pgconn.NewCommandTag("DELETE 0"), nil)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("mapDeleteUserResult() error = %v, want %v", err, ErrUserNotFound)
	}
}

func TestMapDeleteUserResultWrapsDatabaseFailure(t *testing.T) {
	t.Parallel()

	databaseError := errors.New("database unavailable")
	err := mapDeleteUserResult(pgconn.CommandTag{}, databaseError)
	if !errors.Is(err, databaseError) {
		t.Fatalf("mapDeleteUserResult() error = %v, want wrapped %v", err, databaseError)
	}
}
