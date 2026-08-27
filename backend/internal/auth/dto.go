package auth

type RegisterDTO struct {
	Name        *string `json:"name"`
	Email       string  `json:"email"`
	PhoneNumber *string `json:"phone_number"`
	Password    string  `json:"password"`
}
type loginRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponseDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
