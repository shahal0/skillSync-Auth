package postgres

import (
	model "skillsync-authservice/domain/models"
	"skillsync-authservice/domain/repository"
	"skillsync-authservice/pkg"
	"strconv"

	"gorm.io/gorm"
)

type candidatePG struct {
	db       *gorm.DB
	jwtMaker *pkg.JWTMaker
}

func NewCandidateRepository(db *gorm.DB, jwtMaker *pkg.JWTMaker) repository.CandidateRepository {
	return &candidatePG{db: db, jwtMaker: jwtMaker}
}

func (c *candidatePG) CreateCandidate(profile *model.Candidate) (int, error) {
	if err := c.db.Create(profile).Error; err != nil {
		return 0, err
	}
	res, _ := strconv.Atoi(profile.ID)
	return res, nil
}

func (c *candidatePG) UpdateCandidate(input *model.UpdateCandidateInput) error {
	var skills []model.Skills
	var education []model.Education
	if err := c.db.Where("candidate_id = ?", input.ID).Find(&education).Error; err != nil {
	}
	if err := c.db.Where("candidate_id = ?", input.ID).Find(&skills).Error; err != nil {


	}

	candidate := model.Candidate{
		Email:     input.Email,
		Name:      input.Name,
		Phone: input.Phone,
		Experience: input.Experience,
		Skills:    skills,
		Resume:   input.Resume,
		Education: education,
		CurrentLocation: input.CurrentLocation,
		PreferredLocation: input.PreferredLocation,
	}
	return c.db.Save(&candidate).Error
}

func (c *candidatePG) GetCandidateByUserID(userID string) (*model.Candidate, error) {
	var cand model.Candidate
	if err := c.db.Where("user_id = ?", userID).First(&cand).Error; err != nil {
		return nil, err
	}
	return &cand, nil
}

func (c *candidatePG) AddEducation(education model.Education) error {
	return c.db.Model(&model.Candidate{}).Where("candidate_id = ?", education.CandidateID).Association("Education").Append(&education)
}
func (c *candidatePG) AddSkills(skills model.Skills) error {
	return c.db.Model(&model.Candidate{}).Where("candidate_id = ?", skills.CandidateID).Association("skills").Append(&skills)
}

func (c *candidatePG) GetCandidateByEmail(email string) (*model.Employer, error) {
	var emp model.Employer
	if err := c.db.Where("email = ?", email).First(&emp).Error; err != nil {
		return nil, err
	}
	return &emp, nil
}

func (c *candidatePG) Login(request model.LoginRequest) (*model.LoginResponse, error) {
	var cand model.Candidate
	if err := c.db.Where("email = ? AND password = ?", request.Email, request.Password).First(&cand).Error; err != nil {
		return nil, err
	}
	response := &model.LoginResponse{
		ID:    cand.ID,
		Role:  "candidate",
		Token: func() string {
			token, _ := c.jwtMaker.GenerateToken(cand.ID, "candidate")
			return token
		}(),
	}
	return response, nil
}

func (c *candidatePG) Signup(request model.SignupRequest) (*model.AuthResponse, error) {
	candidate := model.Candidate{
		Email:    request.Email,
		Name:     request.Name,
		Password: request.Password,
	}
	if err := c.db.Create(&candidate).Error; err != nil {
		return nil, err
	}

	response := &model.AuthResponse{
		ID:    candidate.ID,
		Message: "Candidate created successfully",
	}
	return response, nil
}