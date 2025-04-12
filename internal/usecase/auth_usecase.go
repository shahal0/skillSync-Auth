package usecase

import (
	"errors"
	models "skillsync-authservice/domain/models"
	domain "skillsync-authservice/domain/repository"
	"skillsync-authservice/pkg"
	"strconv"
	"strings"
)

type authUsecase struct {
	employerRepo domain.EmployerRepository
	candidateRepo domain.CandidateRepository
	jwtMaker      *pkg.JWTMaker
	hasher        *pkg.PasswordManager
}

// Login implements the login method for the authUsecase struct.
func (a *authUsecase) Login(input models.LoginRequest) (*models.LoginResponse, error) {
	role := strings.ToLower(input.Role)
	if role != "candidate" && role != "employer" {
		return nil, errors.New("invalid role")
	}

	var userID string
	var err error
	if role == "employer" {
		employer, err := a.employerRepo.GetEmployerByEmail(input.Email)
		if err != nil {
			return nil, errors.New("invalid credentials")
		}
		if err:= a.hasher.CheckPassword(input.Password, employer.Password);err!=nil {
			return nil, errors.New("invalid credentials")
		}
		userID = strconv.Itoa(employer.ID)
	} else {
		candidate, err := a.candidateRepo.GetCandidateByEmail(input.Email)
		if err != nil {
			return nil, errors.New("invalid credentials")
		}
		if err:=a.hasher.CheckPassword(input.Password, candidate.Password);err!=nil {
			return nil, errors.New("invalid credentials")
		}
		userID = strconv.Itoa(candidate.ID)
	}

	token, err := a.jwtMaker.GenerateToken(userID, role)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		ID:      userID,
		Role:    role,
		Token:   token,
	}, nil
}


func NewAuthUsecase(
	employerRepo domain.EmployerRepository,
	candidateRepo domain.CandidateRepository,
	jwtMaker *pkg.JWTMaker,
	hasher *pkg.PasswordManager,
) domain.AuthUsecase {
	return &authUsecase{
		employerRepo:  employerRepo,
		candidateRepo: candidateRepo,
		jwtMaker:      jwtMaker,
		hasher:        hasher,
	}
}

func (a *authUsecase) Signup(input models.SignupRequest) (*models.AuthResponse, error) {
	role := strings.ToLower(input.Role)
	if role != "candidate" && role != "employer" {
		return nil, errors.New("invalid role")
	}

	hashedPassword, err := a.hasher.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	if role == "employer" {
		employer := models.Employer{
			CompanyName:     input.Name,
			Email:    input.Email,
			Password: hashedPassword,
		}
		 id,err := a.employerRepo.CreateEmployer(&employer)
		if err != nil {
			return nil, err
		}

		
		return &models.AuthResponse{
			ID:    strconv.Itoa(id),
			Message: "Employer registered successfully",
		}, nil
	}

	candidate := models.Candidate{
		Name:     input.Name,
		Email:    input.Email,
		Password: hashedPassword,
	}
	id, err := a.candidateRepo.CreateCandidate(&candidate)
	if err != nil {
		return nil, err
	}
	return &models.AuthResponse{
		ID:    strconv.Itoa(id),
		Message: "Candidate registered successfully",
	}, nil
}
