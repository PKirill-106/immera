package user

func toUserByIDResponse(user User) userByIDResponse {
	return userByIDResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
	}
}
