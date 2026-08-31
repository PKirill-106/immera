package auth

import "golang.org/x/crypto/bcrypt"

const defaultCost = bcrypt.DefaultCost

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), defaultCost)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func comparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
}
