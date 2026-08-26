package auth

type RegisterUserDTO struct {
	Name        *string `json:"name"`
	Email       string  `json:"email"`
	PhoneNumber *string `json:"phone_number"`
	Password    string  `json:"password"`
}
