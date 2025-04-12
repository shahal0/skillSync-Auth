package repository

import model "skillsync-authservice/domain/models"

type UserRepository interface {
	Create(user *model.User) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	FindByID(id string) (*model.User, error)
	UpdatePassword(userID, hashedPassword string) error
	VerifyEmail(userID string) error
}
