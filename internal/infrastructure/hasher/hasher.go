package hasher

import "golang.org/x/crypto/bcrypt"

// Bcrypt — реализация через bcrypt.
type bcryptHasher struct{}

func NewBcryptHasher() *bcryptHasher { return &bcryptHasher{} }

func (bcryptHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (bcryptHasher) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
