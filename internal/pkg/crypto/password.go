package crypto

import "golang.org/x/crypto/bcrypt"



type bcryptHasher struct {
	cost int
}

func NewBcryptHasher(c int) *bcryptHasher {
	return &bcryptHasher{cost: c}
}

func (b *bcryptHasher) Generate(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil{
		return "", err
	}
	return string(hash), nil
}

func (b *bcryptHasher) Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}