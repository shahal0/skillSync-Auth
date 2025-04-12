package domain

type Employer struct {
    ID          int    `gorm:"primaryKey;autoIncrement"` // Auto-incrementing integer
    Email       string
    Password    string
    CompanyName string
    Phone       int64
    Industry    string
    Location    string
    Website     string
    IsTrusted   bool
}

type UpdateEmployerInput struct {
	ID          string `json:"id"`
	CompanyName string `json:"company_name"`
	Phone       int64  `json:"phone"`
	Email       string `json:"email"`
	Industry    string `json:"industry"`
	Location    string `json:"location"`
	Website     string `json:"website"`
}
