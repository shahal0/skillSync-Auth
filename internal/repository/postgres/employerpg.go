package postgres

import (
	//"go/token"

	model "skillsync-authservice/domain/models"
	"skillsync-authservice/domain/repository"
	"skillsync-authservice/pkg"
	"strconv"

	//"github.com/golang-jwt/jwt"
	"gorm.io/gorm"

	//"context"
	"errors"
)

type employerPG struct {
	db       *gorm.DB
	jwtMaker *pkg.JWTMaker
}

func NewEmployerRepository(db *gorm.DB) repository.EmployerRepository {
	return &employerPG{db: db, jwtMaker: pkg.NewJWTMaker("your-secret-key")}
}

func (e *employerPG) CreateEmployer(profile *model.Employer) (int, error) {
	if err := e.db.Create(profile).Error; err != nil {
		return 0, err
	}
	return profile.ID, nil
}

func (e *employerPG) UpdateEmployer(input *model.UpdateEmployerInput) error {
	var profile model.Employer
	if err := e.db.First(&profile, "id = ?", input.ID).Error; err != nil {
		return err
	}
	profile.CompanyName = input.CompanyName
	profile.Email = input.Email
	profile.Industry = input.Industry
	profile.Phone = input.Phone
	profile.Location = input.Location
	profile.Website = input.Website
	return e.db.Save(&profile).Error
}

func (e *employerPG) GetEmployerByUserID(userID string) (*model.Employer, error) {
	var emp model.Employer
	if err := e.db.Where("id= ?", userID).First(&emp).Error; err != nil {
		return nil, err
	}
	return &emp, nil
}
func (r *employerPG) GetEmployerByEmail(email string) (*model.Employer, error) {
	query := `SELECT id, name, email, password, company_name, role FROM employers WHERE email = $1`

	var emp model.Employer
	err := r.db.Raw(query, email).Scan(&emp).Error
	if err != nil {
		return nil, errors.New("employer not found")
	}

	return &emp, nil
}

func (r *employerPG) Login(request model.LoginRequest) (*model.LoginResponse, error) {
	var emp model.Employer
	if err := r.db.Where("email = ? AND password = ?", request.Email, request.Password).First(&emp).Error; err != nil {
		return nil, errors.New("invalid email or password")
	}
	id:= strconv.Itoa(emp.ID)
	token, err := r.jwtMaker.GenerateToken(id, "employer")
	if err != nil {
		return nil, err
	}
	response := &model.LoginResponse{
		ID:    id,
		Role:  "employer",
		Token: token,
	}
	return response, nil
}
func (e *employerPG) Signup(request model.SignupRequest) (*model.AuthResponse, error) {
	profile := &model.Employer{
		CompanyName: request.Name,
		Email:       request.Email,
		Password:    request.Password,
	}
	if err := e.db.Create(profile).Error; err != nil {
		return nil, err
	}
	response := &model.AuthResponse{
		ID:      strconv.Itoa(profile.ID),
		Message: "Employer created successfully",
	}
	return response, nil
}
