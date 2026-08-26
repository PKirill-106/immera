package auth

type RegisterParams struct {
	Name         *string
	Email        string
	PhoneNumber  *string
	PasswordHash string
}
