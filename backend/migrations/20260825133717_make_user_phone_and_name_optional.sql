-- +goose Up

ALTER TABLE users
    ALTER COLUMN name DROP NOT NULL,
    ALTER COLUMN phone_number DROP NOT NULL;

-- +goose Down

ALTER TABLE users
    ALTER COLUMN name SET NOT NULL,
    ALTER COLUMN phone_number SET NOT NULL;