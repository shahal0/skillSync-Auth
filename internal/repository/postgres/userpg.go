package postgres

import (
	model "skillsync-authservice/domain/models"
	"skillsync-authservice/domain/repository"

	"gorm.io/gorm"
)

type userPG struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userPG{db}
}

func (u *userPG) Create(user *model.User) (*model.User, error) {
	if err := u.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (u *userPG) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := u.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *userPG) FindByID(id string) (*model.User, error) {
	var user model.User
	if err := u.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *userPG) UpdatePassword(userID, hashedPassword string) error {
	return u.db.Model(&model.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

func (u *userPG) VerifyEmail(userID string) error {
	return u.db.Model(&model.User{}).Where("id = ?", userID).Update("is_verified", true).Error
}
