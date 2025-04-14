package main

import (
	"log"

	"skillsync-authservice/config"
	models "skillsync-authservice/domain/models"
	handler "skillsync-authservice/internal/delivery/handler"
	repository "skillsync-authservice/internal/repository/postgres"
	"skillsync-authservice/internal/usecase"
	"skillsync-authservice/pkg"

	"gorm.io/driver/postgres" // PostgreSQL driver for GORM
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize the database connection
	// Initialize the configuration
	cfg,err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	db, err := gorm.Open(postgres.Open(cfg.DB.Config().ConnString()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Run migrations
	err = db.AutoMigrate(
		&models.Employer{},
		&models.Candidate{},
		&models.User{},
		&models.Skills{},
		&models.Education{},
		&models.VerificationTable{},
	)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize the employer repository and usecase
	employerRepo := repository.NewEmployerRepository(db)
	jwtMaker := pkg.NewJWTMaker("your-secret-key")
	employerUsecase := usecase.NewEmployerUsecase(employerRepo, jwtMaker, db)

	// Initialize the employer handler and routes
	router := gin.Default()
	handler.NewEmployerHandler(router.Group(""), employerUsecase)
	candidaterepo:= repository.NewCandidateRepository(db, jwtMaker)
	candidateUsecase := usecase.NewCandidateUsecase(candidaterepo, jwtMaker, db)
	handler.NewCandidateHandler(router.Group(""), candidateUsecase)


	// Start the server
	router.Run(":8060")
}
