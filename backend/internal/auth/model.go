package auth

type RegisterUserParams struct {
	Name         *string
	Email        string
	PhoneNumber  *string
	PasswordHash string
}
