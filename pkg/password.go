package pkg

import (
	"encoding/base64"
	"errors"
	"log"
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
	log.Println("Generating OTP for:", to)

	// Random OTP generation
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	otp := r.Intn(900000) + 100000 // Generates a 6-digit OTP
	log.Println("Generated OTP:", otp)

	now := time.Now().Unix()
	expiryTime := now + int64(otpexpiry) // OTP expiry time

	// Create or update the OTP record in the database
	verification := models.VerificationTable{
		Email:              to,
		OTP:                uint64(otp),
		OTPExpiry:          uint64(expiryTime),
		VerificationStatus: false,
	}

	if err := db.Where("email = ?", verification.Email).
		Assign(verification).
		FirstOrCreate(&verification).Error; err != nil {
		log.Println("Failed to store OTP in database:", err)
		return errors.New("failed to store OTP information in the database: " + err.Error())
	}

	// Send email
	auth := smtp.PlainAuth("", "petplate0@gmail.com", "fsjazfcjcllfnxqu", "smtp.gmail.com")
	message := []byte("Subject: Your OTP Code\n\nYour OTP is: " + strconv.Itoa(otp))
	err := smtp.SendMail("smtp.gmail.com:587", auth, "petplate0@gmail.com", []string{to}, message)
	if err != nil {
		log.Println("Failed to send email:", err)
		return errors.New("failed to send email: " + err.Error())
	}

	log.Println("OTP sent successfully to:", to)
	return nil
}

func GenerateStateToken() string {
	b := make([]byte, 16) // 16 bytes = 128 bits
	_, err := rand.Read(b)
	if err != nil {
		log.Println("Error generating state token:", err)
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
