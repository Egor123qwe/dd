package hasher

import (
	"golang.org/x/crypto/bcrypt"
)

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

type hasher struct {
	cost int
}

func New() Hasher {
	return hasher{
		cost: bcrypt.DefaultCost,
	}
}

func (h hasher) Hash(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

func (h hasher) Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
