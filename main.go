package main

import (
	"log"

	"skillsync-authservice/config"
	http "skillsync-authservice/internal/delivery/handler"
	postgress "skillsync-authservice/internal/repository/postgres"
	"skillsync-authservice/internal/usecase"
	"skillsync-authservice/pkg"
	models "skillsync-authservice/domain/models"

	"gorm.io/driver/postgres" // PostgreSQL driver for GORM
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load config and establish DB connection
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	defer cfg.DB.Close()

	gormDB, err := gorm.Open(postgres.Open(cfg.DB.Config().ConnString()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to initialize GORM DB: %v", err)
	}
	err = gormDB.AutoMigrate(
		&models.Employer{},
		&models.Candidate{},
		&models.User{},
		&models.Skills{},
		&models.Education{},
	)
	employerRepo := postgress.NewEmployerRepository(gormDB)
	candidateRepo := postgress.NewCandidateRepository(gormDB,&pkg.JWTMaker{})

	// Initialize use cases
	employerUC := usecase.NewEmployerUsecase(employerRepo)
	candidateUC := usecase.NewCandidateUsecase(candidateRepo)

	// Setup Gin router
	router := gin.Default()
	api := router.Group("")

	// Register handlers
	http.NewEmployerHandler(api, employerUC)
	http.NewCandidateHandler(api, candidateUC)

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8060"
	}
	log.Printf("Server is running on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
