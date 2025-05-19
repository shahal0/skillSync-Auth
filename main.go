package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	logger "skillsync-authservice/Logger"
	"skillsync-authservice/config"
	models "skillsync-authservice/domain/models"
	repository "skillsync-authservice/internal/repository/postgres"
	"skillsync-authservice/internal/usecase"
	"skillsync-authservice/pkg"
	"skillsync-authservice/internal/grpcServer"
	
	"context"

	"cloud.google.com/go/storage"
	"github.com/shahal0/skillsync-protos/gen/authpb"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	db, err := gorm.Open(postgres.Open(cfg.DB.Config().ConnString()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	err = db.AutoMigrate(
		&models.Employer{},
		&models.Candidate{},
		&models.User{},
		&models.Skills{},
		&models.Education{},
		&models.VerificationTable{},
		&models.CandidateResume{},
	)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	logger.InitLogger()

	employerRepo := repository.NewEmployerRepository(db)
	jwtMaker := pkg.NewJWTMaker("your_jwt_secret")
	token, err := jwtMaker.GenerateToken("3", "employer")
	if err != nil {
		log.Println("Error generating token:", err)
	}
	log.Println("Generated Token:", token)
	employerUsecase := usecase.NewEmployerUsecase(employerRepo, jwtMaker, db)

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create GCS client: %v", err)
	}

	bucketName := "skillsync-resume"
	gcsUploader := pkg.NewGcsClient(client, bucketName)
	if gcsUploader == nil {
		log.Fatalf("GCS client initialized successfully")
	}
	log.Println("GCS client initialized ", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))

	candidaterepo := repository.NewCandidateRepository(db, jwtMaker)
	candidateUsecase := usecase.NewCandidateUsecase(candidaterepo, jwtMaker, gcsUploader)

	// Start only the gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	authpb.RegisterAuthServiceServer(grpcServer, grpcserver.NewAuthGRPCServer(*candidateUsecase, *employerUsecase))
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down server...")
		grpcServer.GracefulStop()
		os.Exit(0)
	}()

	log.Println("gRPC server is running on port 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
