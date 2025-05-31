package grpcserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	//"go/token"
	"log"
	"mime/multipart"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	models "skillsync-authservice/domain/models"
	"skillsync-authservice/internal/usecase"

	authpb "skillsync-authservice/skillsync-protos/gen/authpb"
)

// --- gRPC Server Implementation ---
type authGRPCServer struct {
	authpb.UnimplementedAuthServiceServer
	candidateUsecase usecase.CandidateUsecase
	employerUsecase  usecase.EmployerUsecase
}

func NewAuthGRPCServer(candidateUsecase usecase.CandidateUsecase, employerUsecase usecase.EmployerUsecase) authpb.AuthServiceServer {
	return &authGRPCServer{
		candidateUsecase: candidateUsecase,
		employerUsecase:  employerUsecase,
	}
}

// --- Candidate Endpoints ---

func (s *authGRPCServer) CandidateSignup(ctx context.Context, req *authpb.CandidateSignupRequest) (*authpb.CandidateSignupResponse, error) {
	domainReq := models.CandidateSignupRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		Name:     req.GetName(),
	}
	resp, err := s.candidateUsecase.Signup(models.SignupRequest{
		Email:    domainReq.Email,
		Password: domainReq.Password,
		Name:     domainReq.Name,
	})
	if err != nil {
		return nil, err
	}
	return &authpb.CandidateSignupResponse{
		Id:      resp.ID,
		Message: resp.Message,
	}, nil
}

func (s *authGRPCServer) CandidateLogin(ctx context.Context, req *authpb.CandidateLoginRequest) (*authpb.CandidateLoginResponse, error) {
	domainReq := models.CandidateLoginRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}
	resp, err := s.candidateUsecase.Login(models.LoginRequest{
		Email:    domainReq.Email,
		Password: domainReq.Password,
	})
	if err != nil {
		return nil, err
	}
	// Include token in the response
	return &authpb.CandidateLoginResponse{
		Id:      resp.ID,
		Token:   resp.Token,
		Message: "Candidate logged in successfully",
	}, nil
}

func (s *authGRPCServer) CandidateVerifyEmail(ctx context.Context, req *authpb.VerifyEmailRequest) (*authpb.GenericResponse, error) {
	domainReq := models.VerifyEmailRequest{
		Email: req.GetEmail(),
		Otp:   req.GetOtp(),
	}
	otp, err := strconv.ParseUint(domainReq.Otp, 10, 64)
	if err != nil {
		return nil, err
	}
	err = s.candidateUsecase.VerifyEmail(ctx, domainReq.Email, otp)
	if err != nil {
		return nil, err
	}
	return &authpb.GenericResponse{
		Message: "Email verified successfully",
	}, nil
}

func (s *authGRPCServer) CandidateResendOtp(ctx context.Context, req *authpb.ResendOtpRequest) (*authpb.GenericResponse, error) {
	domainReq := models.ResendOtpRequest{
		Email: req.GetEmail(),
	}
	err := s.candidateUsecase.ResendOtp(ctx, domainReq.Email)
	if err != nil {
		return nil, err
	}
	return &authpb.GenericResponse{
		Message: "OTP sent successfully",
		Success: true,
	}, nil
}

func (s *authGRPCServer) CandidateForgotPassword(ctx context.Context, req *authpb.ForgotPasswordRequest) (*authpb.GenericResponse, error) {
	email := req.GetEmail()
	err := s.candidateUsecase.ForgotPassword(ctx, email)
	if err != nil {
		return nil, err
	}
	return &authpb.GenericResponse{
		Message: "Password reset OTP sent successfully",
		Success: true,
	}, nil
}

func (s *authGRPCServer) CandidateResetPassword(ctx context.Context, req *authpb.ResetPasswordRequest) (*authpb.GenericResponse, error) {
	// Convert string OTP to uint64
	otp, err := strconv.ParseUint(req.GetOtp(), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid OTP format: %v", err)
	}

	// Call the usecase with the correct parameters
	err = s.candidateUsecase.ResetPassword(ctx, req.GetEmail(), otp, req.GetNewPassword())
	if err != nil {
		return nil, err
	}

	return &authpb.GenericResponse{
		Message: "Password reset successful",
		Success: true,
	}, nil
}

func (s *authGRPCServer) CandidateChangePassword(ctx context.Context, req *authpb.ChangePasswordRequest) (*authpb.GenericResponse, error) {
	// Get request parameters
	email := req.GetEmail()
	oldPassword := req.GetOldPassword()
	newPassword := req.GetNewPassword()

	// Call the usecase method with the correct parameters
	err := s.candidateUsecase.ChangePassword(ctx, email, oldPassword, newPassword)
	if err != nil {
		return nil, err
	}

	// Return the GenericResponse as defined in the proto file
	return &authpb.GenericResponse{
		Message: "Password changed successfully",
		Success: true,
	}, nil
}

func (s *authGRPCServer) CandidateProfile(ctx context.Context, req *authpb.CandidateProfileRequest) (*authpb.CandidateProfileResponse, error) {
	// Extract token from the request
	token := req.GetToken()

	// Call GetProfile with the token
	resp, err := s.candidateUsecase.GetProfile(ctx, token)
	if err != nil {
		return nil, err
	}
	// Convert skills to proto format
	skills := make([]*authpb.Skill, 0)
	for _, skill := range resp.Skills {
		skills = append(skills, &authpb.Skill{
			CandidateId: skill.CandidateID,
			Skill:       skill.Skill,
			Level:       skill.Level,
		})
	}

	// Convert education to proto format
	education := make([]*authpb.Education, 0)
	for _, edu := range resp.Education {
		education = append(education, &authpb.Education{
			CandidateId: edu.CandidateID,
			University:  edu.University,
			Location:    edu.Location,
			Major:       edu.Major,
			StartDate:   edu.StartDate,
			EndDate:     edu.EndDate,
			Grade:       edu.Grade,
		})
	}

	return &authpb.CandidateProfileResponse{
		Id:                resp.ID,
		Email:             resp.Email,
		Name:              resp.Name,
		Phone:             resp.Phone,
		Experience:        resp.Experience,
		Skills:            skills,
		Resume:            resp.Resume,
		Education:         education,
		CurrentLocation:   resp.CurrentLocation,
		PreferredLocation: resp.PreferredLocation,
		Linkedin:          resp.Linkedin,
		Github:            resp.Github,
		ProfilePicture:    resp.ProfilePicture,
		IsVerified:        resp.IsVerified,
	}, nil
}

func (s *authGRPCServer) CandidateProfileUpdate(ctx context.Context, req *authpb.CandidateProfileUpdateRequest) (*authpb.GenericResponse, error) {
	// Create update input from request
	updateInput := &models.UpdateCandidateInput{
		Email: req.GetEmail(),
		Name:  req.GetName(),
		// Add more fields as needed
	}

	// Use the existing UpdateCandidateProfile method
	err := s.candidateUsecase.UpdateCandidateProfile(ctx, updateInput, req.GetToken())
	if err != nil {
		return nil, err
	}

	// Return success response directly
	return &authpb.GenericResponse{
		Message: "Profile updated successfully",
		Success: true,
	}, nil
}

func (s *authGRPCServer) CandidateSkillsUpdate(ctx context.Context, req *authpb.SkillsUpdateRequest) (*authpb.GenericResponse, error) {
	// Extract token from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}
	token := tokens[0]
	var skillNames []string
	for _, skill := range req.GetSkills() {
		skillNames = append(skillNames, skill.GetSkill())
	}

	skills := models.Skills{
		Skill: strings.Join(skillNames, ","),
	}

	// Call the usecase with correct parameters
	err := s.candidateUsecase.AddSkills(ctx, skills, token)
	if err != nil {
		return nil, err
	}

	return &authpb.GenericResponse{
		Message: "Skills updated successfully",
		Success: true,
	}, nil
}

func (s *authGRPCServer) CandidateEducationUpdate(ctx context.Context, req *authpb.EducationUpdateRequest) (*authpb.GenericResponse, error) {
	// Extract token from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}
	token := tokens[0]

	// Convert proto education to domain model education
	var educations []models.Education
	for _, edu := range req.GetEducation() {
		educations = append(educations, models.Education{
			University: edu.GetUniversity(),
			Location:   edu.GetLocation(),
			Major:      edu.GetMajor(),
			StartDate:  edu.GetStartDate(),
			EndDate:    edu.GetEndDate(),
			Grade:      edu.GetGrade(),
		})
	}

	// Call the usecase with correct parameters
	for _, education := range educations {
		err := s.candidateUsecase.AddEducation(ctx, education, token)
		if err != nil {
			return nil, err
		}
	}

	return &authpb.GenericResponse{
		Message: "Education updated successfully",
		Success: true,
	}, nil
}
func (s *authGRPCServer) CandidateUploadResume(ctx context.Context, req *authpb.UploadResumeRequest) (*authpb.GenericResponse, error) {
	// Extract user ID from token
	userID, err := s.candidateUsecase.ExtractUserIDFromToken(req.GetToken())
	if err != nil {
		log.Printf("Failed to extract user ID from token: %v", err)
		return &authpb.GenericResponse{
			Message: "Invalid or expired token",
			Success: false,
		}, nil
	}

	// Create a reader from the resume bytes
	file := bytes.NewReader(req.GetResume())

	// Create a multipart file header with appropriate metadata
	fileHeader := &multipart.FileHeader{
		Filename: fmt.Sprintf("%s_resume.pdf", userID),
		Size:     int64(len(req.GetResume())),
	}

	// Call the usecase with the correct parameters
	resumePath, err := s.candidateUsecase.AddResume(ctx, file, fileHeader, userID)
	if err != nil {
		// Log the error for debugging
		log.Printf("UploadResume error: %v", err)
		return &authpb.GenericResponse{
			Message: "Failed to upload resume: " + err.Error(),
			Success: false,
		}, nil
	}
	return &authpb.GenericResponse{
		Message: "Resume uploaded successfully: " + resumePath,
		Success: true,
	}, nil
}

func (s *authGRPCServer) CandidateGoogleLogin(ctx context.Context, req *authpb.GoogleLoginRequest) (*authpb.AuthResponse, error) {
	// Extract the redirect URL from the request
	redirectURL := req.GetRedirectUrl()

	// Call the usecase with just the redirectURL parameter
	authURL, err := s.candidateUsecase.GoogleLogin(redirectURL)
	if err != nil {
		return nil, err
	}

	return &authpb.AuthResponse{
		Token:   "",
		Message: authURL,
	}, nil
}

func (s *authGRPCServer) CandidateGoogleCallback(ctx context.Context, req *authpb.GoogleCallbackRequest) (*authpb.AuthResponse, error) {
	// Extract the code from the request
	code := req.GetCode()

	// Call the usecase with just the code parameter
	resp, err := s.candidateUsecase.GoogleCallback(code)
	if err != nil {
		return nil, err
	}
	return &authpb.AuthResponse{
		Token:   resp.Token,
		Message: "Google callback successful",
	}, nil
}

// --- Employer Endpoints ---

func (s *authGRPCServer) EmployerSignup(ctx context.Context, req *authpb.EmployerSignupRequest) (*authpb.EmployerSignupResponse, error) {
	// Create a context map to store additional employer details
	contextMap := map[string]string{
		"phone":    strconv.FormatInt(req.GetPhone(), 10),
		"industry": req.GetIndustry(),
		"location": req.GetLocation(),
		"website":  req.GetWebsite(),
	}

	domainReq := models.SignupRequest{
		Name:     req.GetCompanyName(),
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		Role:     "employer",
		Context:  contextMap,
	}

	resp, err := s.employerUsecase.Signup(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	idInt, err := strconv.ParseInt(resp.ID, 10, 64)
	if err != nil {
		return nil, err
	}
	return &authpb.EmployerSignupResponse{
		Id:      idInt,
		Message: resp.Message,
	}, nil
}

func (s *authGRPCServer) EmployerLogin(ctx context.Context, req *authpb.EmployerLoginRequest) (*authpb.EmployerLoginResponse, error) {
	domainReq := models.LoginRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		Role:     "employer",
	}
	resp, err := s.employerUsecase.Login(domainReq)
	if err != nil {
		return nil, err
	}
	// Convert string ID to int64
	idInt, err := strconv.ParseInt(resp.ID, 10, 64)
	if err != nil {
		return nil, errors.New("failed to parse ID: " + err.Error())
	}
	// Include all fields in the response (ID, token, and message)
	return &authpb.EmployerLoginResponse{
		Id:      idInt,
		Token:   resp.Token,
		Message: resp.Message,
	}, nil
}

func (s *authGRPCServer) EmployerVerifyEmail(ctx context.Context, req *authpb.VerifyEmailRequest) (*authpb.GenericResponse, error) {
	// Convert string OTP to uint64
	otp, err := strconv.ParseUint(req.GetOtp(), 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid OTP format: %v", err)
	}

	// Call the usecase with individual parameters
	err = s.employerUsecase.VerifyEmail(ctx, req.GetEmail(), otp)
	if err != nil {
		return nil, err
	}

	// Return success response
	return &authpb.GenericResponse{
		Message: "Email verified successfully",
		Success: true,
	}, nil
}

func (s *authGRPCServer) EmployerForgotPassword(ctx context.Context, req *authpb.ForgotPasswordRequest) (*authpb.GenericResponse, error) {
	domainReq := models.ForgotPasswordRequest{
		Email: req.GetEmail(),
	}
	err := s.employerUsecase.ForgotPassword(ctx, domainReq.Email)
	if err != nil {
		return nil, err
	}
	return &authpb.GenericResponse{
		Message: "Password reset OTP sent successfully",
		Success: true,
	}, nil
}

func (s *authGRPCServer) EmployerResetPassword(ctx context.Context, req *authpb.ResetPasswordRequest) (*authpb.GenericResponse, error) {
	// Convert string OTP to uint64
	otp, err := strconv.ParseUint(req.GetOtp(), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid OTP format: %v", err)
	}

	// Call the usecase with the correct parameters
	err = s.employerUsecase.ResetPassword(ctx, req.GetEmail(), otp, req.GetNewPassword())
	if err != nil {
		return nil, err
	}

	return &authpb.GenericResponse{
		Message: "Password reset successful",
		Success: true,
	}, nil
}

func (s *authGRPCServer) EmployerChangePassword(ctx context.Context, req *authpb.ChangePasswordRequest) (*authpb.GenericResponse, error) {
	// Get request parameters
	email := req.GetEmail()
	oldPassword := req.GetOldPassword()
	newPassword := req.GetNewPassword()

	// Call the usecase method with the correct parameters
	err := s.employerUsecase.ChangePassword(ctx, email, oldPassword, newPassword)
	if err != nil {
		return nil, err
	}
	return &authpb.GenericResponse{
		Message: "Password changed successfully",
		Success: true,
	}, nil
}

func (s *authGRPCServer) EmployerProfile(ctx context.Context, req *authpb.EmployerProfileRequest) (*authpb.EmployerProfileResponse, error) {
	// Extract token from the request
	token := req.GetToken()

	// Call GetProfile with the token
	resp, err := s.employerUsecase.GetProfile(ctx, token)
	if err != nil {
		return nil, err
	}
	return &authpb.EmployerProfileResponse{
		Id:          int64(resp.ID),
		Email:       resp.Email,
		CompanyName: resp.CompanyName,
		Phone:       resp.Phone,
		Industry:    resp.Industry,
		Location:    resp.Location,
		Website:     resp.Website,
		IsVerified:  resp.IsVerified,
		IsTrusted:   resp.IsTrusted,
	}, nil
}

func (s *authGRPCServer) EmployerProfileById(ctx context.Context, req *authpb.EmployerProfileByIdRequest) (*authpb.EmployerProfileResponse, error) {
	// Extract employer ID from the request
	employerId := req.GetEmployerId()

	// Call GetProfileById with the employer ID
	// We need to add this method to the employer usecase
	resp, err := s.employerUsecase.GetProfileById(ctx, employerId)
	if err != nil {
		return nil, err
	}
	return &authpb.EmployerProfileResponse{
		Id:          int64(resp.ID),
		Email:       resp.Email,
		CompanyName: resp.CompanyName,
		Phone:       resp.Phone,
		Industry:    resp.Industry,
		Location:    resp.Location,
		Website:     resp.Website,
		IsVerified:  resp.IsVerified,
		IsTrusted:   resp.IsTrusted,
	}, nil
}

func (s *authGRPCServer) EmployerProfileUpdate(ctx context.Context, req *authpb.EmployerProfileUpdateRequest) (*authpb.GenericResponse, error) {
	domainReq := models.UpdateEmployerInput{
		Email:       req.GetEmail(),
		CompanyName: req.GetCompanyName(),
		Phone:       req.GetPhone(),
		Industry:    req.GetIndustry(),
		Location:    req.GetLocation(),
		Website:     req.GetWebsite(),
	}
	err := s.employerUsecase.UpdateProfile(ctx, &domainReq)
	if err != nil {
		return nil, err
	}
	return &authpb.GenericResponse{
		Message: "profile updated successfully",
		Success: true,
	}, nil
}

func (s *authGRPCServer) EmployerGoogleLogin(ctx context.Context, req *authpb.GoogleLoginRequest) (*authpb.AuthResponse, error) {
	redirectURL := req.GetRedirectUrl()
	url, err := s.employerUsecase.GoogleLogin(redirectURL)
	if err != nil {
		return nil, err
	}
	return &authpb.AuthResponse{
		Token:   "",
		Message: url,
	}, nil
}

func (s *authGRPCServer) EmployerGoogleCallback(ctx context.Context, req *authpb.GoogleCallbackRequest) (*authpb.AuthResponse, error) {
	code := req.GetCode()
	// Create a models.GoogleCallbackRequest struct with the code
	callbackReq := models.GoogleCallbackRequest{
		Code: code,
	}
	response, err := s.employerUsecase.GoogleCallback(ctx, callbackReq)
	if err != nil {
		return nil, err
	}

	return &authpb.AuthResponse{
		Token:   response.Token,
		Message: response.Message,
		Id:      response.ID,
		Role:    response.Role,
	}, nil
}

func (s *authGRPCServer) GetCandidateSkills(ctx context.Context, req *authpb.GetCandidateSkillsRequest) (*authpb.GetCandidateSkillsResponse, error) {
	// Get candidate ID from request
	candidateID := req.GetCandidateId()

	// Get skills from candidate usecase
	skills, err := s.candidateUsecase.GetSkills(ctx, candidateID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get candidate skills: %v", err)
	}

	return &authpb.GetCandidateSkillsResponse{
		Skills: skills,
	}, nil
}

// VerifyToken verifies a JWT token and returns user ID and role
func (s *authGRPCServer) VerifyToken(ctx context.Context, req *authpb.VerifyTokenRequest) (*authpb.VerifyTokenResponse, error) {
	// Extract token from the request
	var token string

	// First try to get token from the request
	if req != nil {
		token = req.GetToken()
	}

	// If no token in request, try to extract from metadata (Authorization header)
	if token == "" {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			auth := md.Get("authorization")
			if len(auth) > 0 {
				token = auth[0]
			}
		}
	}

	// If still no token, return error
	if token == "" {
		return nil, status.Errorf(codes.Unauthenticated, "No token provided")
	}

	// Handle Bearer token format if present
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	// Verify the token
	claims, err := s.candidateUsecase.VerifyToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid token: %v", err)
	}

	// Return the response with user ID and role
	return &authpb.VerifyTokenResponse{
		UserId: claims.UserID,
		Role:   claims.Role,
	}, nil
}

// GetCandidatesWithPagination implements the GetCandidatesWithPagination gRPC method
func (s *authGRPCServer) GetCandidatesWithPagination(ctx context.Context, req *authpb.GetCandidatesRequest) (*authpb.GetCandidatesResponse, error) {
	// Extract pagination parameters from the request
	paginationReq := req.GetPagination()
	page := int32(1)
	limit := int32(10)

	if paginationReq != nil {
		page = paginationReq.GetPage()
		limit = paginationReq.GetLimit()
	}

	// Create filters map from request parameters
	filters := make(map[string]interface{})
	if req.GetKeyword() != "" {
		filters["keyword"] = req.GetKeyword()
	}
	if req.GetSkill() != "" {
		filters["skill"] = req.GetSkill()
	}
	if req.GetLocation() != "" {
		filters["location"] = req.GetLocation()
	}
	if req.GetMinExperience() > 0 {
		filters["min_experience"] = req.GetMinExperience()
	}

	// Call the usecase to get paginated candidates
	candidates, totalCount, err := s.candidateUsecase.GetCandidatesWithPagination(ctx, page, limit, filters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch candidates: %v", err)
	}

	// Convert domain candidates to protobuf format
	pbCandidates := make([]*authpb.CandidateProfileResponse, 0, len(candidates))
	for _, candidate := range candidates {
		// Convert skills to proto format
		skills := make([]*authpb.Skill, 0, len(candidate.Skills))
		for _, skill := range candidate.Skills {
			skills = append(skills, &authpb.Skill{
				CandidateId: skill.CandidateID,
				Skill:       skill.Skill,
				Level:       skill.Level,
			})
		}

		// Convert education to proto format
		education := make([]*authpb.Education, 0, len(candidate.Education))
		for _, edu := range candidate.Education {
			education = append(education, &authpb.Education{
				CandidateId: edu.CandidateID,
				University:  edu.University,
				Location:    edu.Location,
				Major:       edu.Major,
				StartDate:   edu.StartDate,
				EndDate:     edu.EndDate,
				Grade:       edu.Grade,
			})
		}

		// Create the candidate profile response
		pbCandidates = append(pbCandidates, &authpb.CandidateProfileResponse{
			Id:                candidate.ID,
			Email:             candidate.Email,
			Name:              candidate.Name,
			Phone:             candidate.Phone,
			Experience:        candidate.Experience,
			Skills:            skills,
			Resume:            candidate.Resume,
			Education:         education,
			CurrentLocation:   candidate.CurrentLocation,
			PreferredLocation: candidate.PreferredLocation,
			Linkedin:          candidate.Linkedin,
			Github:            candidate.Github,
			ProfilePicture:    candidate.ProfilePicture,
			IsVerified:        candidate.IsVerified,
		})
	}

	// Calculate total pages
	totalPages := (totalCount + int64(limit) - 1) / int64(limit)
	if totalPages < 1 {
		totalPages = 1
	}

	// Create pagination response
	paginationResponse := &authpb.PaginationResponse{
		TotalCount: int32(totalCount),
		Page:       page,
		Limit:      limit,
		TotalPages: int32(totalPages),
	}

	// Return the response
	return &authpb.GetCandidatesResponse{
		Candidates: pbCandidates,
		Pagination: paginationResponse,
	}, nil
}

// GetEmployersWithPagination implements the GetEmployersWithPagination gRPC method
func (s *authGRPCServer) GetEmployersWithPagination(ctx context.Context, req *authpb.GetEmployersRequest) (*authpb.GetEmployersResponse, error) {
	// Extract pagination parameters from the request
	paginationReq := req.GetPagination()
	page := int32(1)
	limit := int32(10)

	if paginationReq != nil {
		page = paginationReq.GetPage()
		limit = paginationReq.GetLimit()
	}

	// Create filters map from request parameters
	filters := make(map[string]interface{})
	if req.GetKeyword() != "" {
		filters["keyword"] = req.GetKeyword()
	}
	if req.GetIndustry() != "" {
		filters["industry"] = req.GetIndustry()
	}
	if req.GetLocation() != "" {
		filters["location"] = req.GetLocation()
	}

	// Call the usecase to get paginated employers
	employers, totalCount, err := s.employerUsecase.GetEmployersWithPagination(ctx, page, limit, filters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch employers: %v", err)
	}

	// Convert domain employers to protobuf format
	pbEmployers := make([]*authpb.EmployerProfileResponse, 0, len(employers))
	for _, employer := range employers {
		pbEmployers = append(pbEmployers, &authpb.EmployerProfileResponse{
			Id:          int64(employer.ID),
			Email:       employer.Email,
			CompanyName: employer.CompanyName,
			Phone:       employer.Phone,
			Industry:    employer.Industry,
			Location:    employer.Location,
			Website:     employer.Website,
			IsVerified:  employer.IsVerified,
			IsTrusted:   employer.IsTrusted,
		})
	}

	// Calculate total pages
	totalPages := (totalCount + int64(limit) - 1) / int64(limit)
	if totalPages < 1 {
		totalPages = 1
	}

	// Create pagination response
	paginationResponse := &authpb.PaginationResponse{
		TotalCount: int32(totalCount),
		Page:       page,
		Limit:      limit,
		TotalPages: int32(totalPages),
	}

	// Return the response
	return &authpb.GetEmployersResponse{
		Employers:  pbEmployers,
		Pagination: paginationResponse,
	}, nil
}
