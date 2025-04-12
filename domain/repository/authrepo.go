package repository

import (
	models "skillsync-authservice/domain/models"
)

type AuthUsecase interface {
	Signup(input models.SignupRequest) (*models.AuthResponse, error)
	Login(input models.LoginRequest) (*models.LoginResponse, error)
	
}

