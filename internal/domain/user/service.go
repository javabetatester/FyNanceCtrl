package user

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Repository Repository
}

func (s *Service) Create(user *User) error {
	user.Id = uuid.New()

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	return s.Repository.Create(user)
}

func (s *Service) Update(user *User) error {
	return s.Repository.Update(user)
}

func (s *Service) Delete(id string) error {
	return s.Repository.Delete(id)
}

func (s *Service) GetByID(id string) (*User, error) {
	return s.Repository.GetById(id)
}

func (s *Service) GetByEmail(email string) (*User, error) {
	return s.Repository.GetByEmail(email)
}
