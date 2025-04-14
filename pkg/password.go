package pkg

import (
	"errors"
	"math/rand"
	"net/smtp"
	models "skillsync-authservice/domain/models"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type PasswordManager struct{}

type dbstruct struct {
	db *gorm.DB
}

func NewPasswordManager() *PasswordManager {
	return &PasswordManager{}
}

// HashPassword hashes a plain-text password
func (pm *PasswordManager) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPassword verifies a password against its hashed version
func (pm *PasswordManager) CheckPassword(plainPassword, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

func SendOtp(db *gorm.DB, to string, otpexpiry uint64) error {
	// Random OTP generation
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	otp := r.Intn(900000) + 100000 // Generates a 6-digit OTP
	now := time.Now().Unix()

	// Set OTP expiry time (5 minutes from now)
	expiryTime := now + 5*60 // 5 minutes in seconds

	// Create the VerificationTable record for storing the OTP
	verification := models.VerificationTable{
		Email:              to,
		OTP:                uint64(otp),
		OTPExpiry:          uint64(expiryTime),
		VerificationStatus: false,
	}

	// Store or update the OTP information in the database
	if err := db.Where("email = ?", verification.Email).
		Assign(verification).
		FirstOrCreate(&verification).Error; err != nil {
		return errors.New("failed to store OTP information in the database: " + err.Error())
	}

	// Send email
	auth := smtp.PlainAuth("", "petplate0@gmail.com", "fsjazfcjcllfnxqu", "smtp.gmail.com")
	message := []byte("Subject: Your OTP Code\n\nYour OTP is: " + strconv.Itoa(otp))
	err := smtp.SendMail("smtp.gmail.com:587", auth, "your-email@gmail.com", []string{to}, message)
	if err != nil {
		return errors.New("failed to send email: " + err.Error())
	}

	return nil
}
