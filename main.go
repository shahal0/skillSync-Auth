package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	logger "skillsync-authservice/Logger"
	"skillsync-authservice/config"
	models "skillsync-authservice/domain/models"
	repository "skillsync-authservice/internal/repository/postgres"
	"skillsync-authservice/internal/usecase"
	"skillsync-authservice/pkg"
	"skillsync-authservice/internal/grpcServer"

	"cloud.google.com/go/storage"
	"github.com/shahal0/skillsync-protos/gen/authpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)


// JWT secret for token generation and validation
const jwtSecretKey = "your_jwt_secret"

func main() {
	// Initialize logger early for better debugging
	logger.InitLogger()
	log.Println("Starting Auth Service...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Configure GORM with better logging and connection settings
	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	}

	// Connect to database
	log.Println("Connecting to database...")
	db, err := gorm.Open(postgres.Open(cfg.DB.Config().ConnString()), gormConfig)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Run database migrations
	log.Println("Running database migrations...")
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
	log.Println("Database migrations completed successfully")

	// Initialize JWT maker with secret key
	jwtMaker := pkg.NewJWTMaker(jwtSecretKey)

	// Initialize repositories and usecases
	log.Println("Initializing repositories and usecases...")
	employerRepo := repository.NewEmployerRepository(db)
	employerUsecase := usecase.NewEmployerUsecase(employerRepo, jwtMaker, db)

	// Initialize GCS client with timeout context
	log.Println("Initializing GCS client...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create GCS client: %v", err)
	}

	// Create GCS uploader
	bucketName := "skillsync-resume"
	gcsUploader := pkg.NewGcsClient(client, bucketName)
	if gcsUploader == nil {
		log.Fatalf("Failed to initialize GCS client")
	}
	log.Println("GCS client initialized successfully")

	// Initialize candidate repository and usecase
	candidaterepo := repository.NewCandidateRepository(db, jwtMaker)
	candidateUsecase := usecase.NewCandidateUsecase(candidaterepo, jwtMaker, gcsUploader)

	// Configure gRPC server with keepalive options for better connection management
	keepaliveOpts := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Minute,
		MaxConnectionAge:      30 * time.Minute,
		MaxConnectionAgeGrace: 5 * time.Minute,
		Time:                  5 * time.Minute,
		Timeout:               20 * time.Second,
	})

	// Start gRPC server
	log.Println("Starting gRPC server...")
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create gRPC server with keepalive options
	grpcServer := grpc.NewServer(keepaliveOpts)

	// Register services
	authpb.RegisterAuthServiceServer(grpcServer, grpcserver.NewAuthGRPCServer(*candidateUsecase, *employerUsecase))
	
	// Enable reflection for easier debugging with tools like grpcurl
	reflection.Register(grpcServer)

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down server...")
		grpcServer.GracefulStop()
		log.Println("Server shutdown complete")
		os.Exit(0)
	}()

	// Start serving
	log.Println("Auth Service gRPC server is running on port 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
