package domain

type Candidate struct {
    ID                string        `gorm:"primaryKey;default:gen_random_uuid()"` // Auto-generate UUID
    Email             string        `json:"email"`
    Password          string        `json:"password"`
    Name              string        `json:"name"`
    Phone             int64         `json:"phone"` 
    Experience        int64         `json:"experience"` 
    Skills            []Skills      `json:"skills"`
    Resume            string        `json:"resume"`
    Education         []Education  `json:"education"`
    CurrentLocation   string        `json:"current_location"`
    PreferredLocation string        `json:"preferred_location"`
    Linkedin          string        `json:"linkedin"` 
    Github            string        `json:"github"`   
    ProfilePicture    string        `json:"profile_picture"` 
    IsVerified        bool          `json:"is_verified"`
}
type Skills struct {
	CandidateID string `json:"candidate_id"`
	Skill string `json:"skill"`
	Level string `json:"level"`
}
type Education struct {
	CandidateID string `json:"candidate_id"`
	University string `json:"university"`
	Location string `json:"location"`
	Major string `json:"major"`
	StartDate string `json:"start_date"`
	EndDate string `json:"end_date"`
	Grade string `json:"grade"`
}
type UpdateCandidateInput struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone int64  `json:"phone"`
	Experience int64  `json:"experience"`
	Skills []Skills `json:"skills"`
	Resume string `json:"resume"`
	Education []Education `json:"education"`
	CurrentLocation string `json:"current_location"`
	PreferredLocation string `json:"preferred_location"`
}
