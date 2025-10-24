package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"oracle_engine/internal/config"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"oracle_engine/internal/server/repository"
	"oracle_engine/internal/utils"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type dashboardService struct {
	repo      repository.DashboardRepository
	jwtSecret string
	config    *config.Config
}

type DashboardService interface {
	// Authentication
	SignUp(ctx context.Context, req *models.SignUpRequest) (*models.SignUpResponse, error)
	Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)

	// Email verification registration
	InitiateRegistration(ctx context.Context, req *models.InitiateRegistrationRequest) (*models.InitiateRegistrationResponse, error)
	VerifyToken(ctx context.Context, token string) (*models.VerifyTokenResponse, error)
	CompleteRegistration(ctx context.Context, req *models.CompleteRegistrationRequest) (*models.CompleteRegistrationResponse, error)
	
	// Password reset
	InitiatePasswordReset(ctx context.Context, req *models.ForgotPasswordRequest) (*models.ForgotPasswordResponse, error)
	VerifyResetToken(ctx context.Context, token string) (*models.VerifyResetTokenResponse, error)
	ResetPassword(ctx context.Context, req *models.ResetPasswordRequest) (*models.ResetPasswordResponse, error)
	
	// Password change (authenticated)
	ChangePassword(ctx context.Context, userID string, req *models.ChangePasswordRequest) (*models.ChangePasswordResponse, error)
	
	// Subscription activation
	ActivateSubscription(ctx context.Context, req *models.ActivateSubscriptionRequest) (*models.ActivateSubscriptionResponse, error)
	
	// Payment storage
	StoreNOWPayment(ctx context.Context, req *models.StorePaymentRequest) (*models.PaymentStorageResponse, error)
	UpdatePaymentStatus(ctx context.Context, req *models.UpdatePaymentStatusRequest) error

	// Profile management
	GetProfile(ctx context.Context, id string) (*models.CompanyProfile, error)
	UpdateProfile(ctx context.Context, id string, req *models.UpdateProfileRequest) (*models.CompanyProfile, error)
	UpdateSubscription(ctx context.Context, id string, req *models.UpdateSubscriptionRequest) (*models.CompanyProfile, error)
	DeleteProfile(ctx context.Context, id string) error

	// API Key management
	CreateAPIKey(ctx context.Context, profileID string, req *models.CreateAPIKeyRequest) (*models.CreateAPIKeyResponse, error)
	GetAPIKeys(ctx context.Context, profileID string) ([]models.APIKey, error)
	GetAPIKeyByID(ctx context.Context, profileID, keyID string) (*models.APIKey, error)
	DeleteAPIKey(ctx context.Context, profileID, keyID string) error
	ValidateAPIKey(ctx context.Context, apiKey string) (*models.APIKey, error)
	CheckAPILimits(ctx context.Context, keyData *models.APIKey) (rateLimited, usageLimitExceeded bool, err error)
	RecordAPIUsage(ctx context.Context, keyData *models.APIKey, endpoint, method, ipAddress, userAgent string) error
	GetAPIUsage(ctx context.Context, profileID string, limit int, offset int) ([]models.APIKeyUsage, error)

	// Payment management (placeholder for future implementation)
	CreatePayment(ctx context.Context, profileID string, req *models.CreatePaymentRequest) (*models.Payment, error)
	GetPaymentHistory(ctx context.Context, profileID string, page, pageSize int) (*models.PaymentHistoryResponse, error)
	
	// Repository access for other services
	GetRepository() repository.DashboardRepository
}

func NewDashboardService(repo repository.DashboardRepository, jwtSecret string, cfg *config.Config) DashboardService {
	return &dashboardService{
		repo:      repo,
		jwtSecret: jwtSecret,
		config:    cfg,
	}
}

// GetRepository returns the repository instance (needed for invoice service)
func (s *dashboardService) GetRepository() repository.DashboardRepository {
	return s.repo
}

func (s *dashboardService) SignUp(ctx context.Context, req *models.SignUpRequest) (*models.SignUpResponse, error) {
	profile, err := s.repo.CreateUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &models.SignUpResponse{
		ID:      profile.ID,
		Message: "User created successfully",
	}, nil
}

func (s *dashboardService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	profile, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password (bcrypt comparison)
	if err := bcrypt.CompareHashAndPassword([]byte(profile.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, expiresIn, err := s.generateJWT(profile.ID, profile.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.LoginResponse{
		ID:          profile.ID,
		Name:        profile.Name,
		Email:       profile.Email,
		AccessToken: token,
		ExpiresIn:   expiresIn,
	}, nil
}

func (s *dashboardService) GetProfile(ctx context.Context, id string) (*models.CompanyProfile, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *dashboardService) UpdateProfile(ctx context.Context, id string, req *models.UpdateProfileRequest) (*models.CompanyProfile, error) {
	return s.repo.UpdateProfile(ctx, id, req)
}

func (s *dashboardService) UpdateSubscription(ctx context.Context, id string, req *models.UpdateSubscriptionRequest) (*models.CompanyProfile, error) {
	// Create a profile update request with just the subscription plan
	// Update subscription plan by calling the repository
	// We'll need to add this method to the repository interface
	return s.repo.UpdateSubscription(ctx, id, req.SubscriptionPlan)
}

func (s *dashboardService) DeleteProfile(ctx context.Context, id string) error {
	// Delete user profile and all associated data
	return s.repo.DeleteUser(ctx, id)
}

func (s *dashboardService) CreateAPIKey(ctx context.Context, profileID string, req *models.CreateAPIKeyRequest) (*models.CreateAPIKeyResponse, error) {
	apiKey, err := s.repo.CreateAPIKey(ctx, profileID, req)
	if err != nil {
		return nil, err
	}

	// Get user profile for email notification
	profile, err := s.repo.GetUserByID(ctx, profileID)
	if err != nil {
		logging.Logger.Error("Failed to get user profile for API key email notification",
			zap.Error(err),
			zap.String("profile_id", profileID))
		// Don't fail API key creation if we can't get profile
	} else {
		// Send API key created notification email
		emailService := utils.NewEmailService()
		userName := profile.Name
		if userName == "" {
			userName = profile.FirstName + " " + profile.LastName
		}
		if userName == "" {
			userName = profile.Email
		}

		// Get first 8 characters of API key as preview
		apiKeyPreview := apiKey.Key
		if len(apiKey.Key) > 8 {
			apiKeyPreview = apiKey.Key[:8]
		}

		if err := emailService.SendAPIKeyCreatedEmail(profile.Email, userName, apiKey.Name, apiKeyPreview); err != nil {
			logging.Logger.Error("Failed to send API key created email",
				zap.Error(err),
				zap.String("email", profile.Email),
				zap.String("key_name", apiKey.Name))
			// Don't fail API key creation if email fails
		} else {
			logging.Logger.Info("API key created email sent successfully",
				zap.String("email", profile.Email),
				zap.String("key_name", apiKey.Name))
		}
	}

	return &models.CreateAPIKeyResponse{
		ID:      apiKey.ID,
		Key:     apiKey.Key,
		Name:    apiKey.Name,
		Message: "API key created successfully. Please store it securely as it will not be shown again.",
	}, nil
}

func (s *dashboardService) GetAPIKeys(ctx context.Context, profileID string) ([]models.APIKey, error) {
	return s.repo.GetAPIKeys(ctx, profileID)
}

func (s *dashboardService) GetAPIKeyByID(ctx context.Context, profileID, keyID string) (*models.APIKey, error) {
	return s.repo.GetAPIKeyByID(ctx, profileID, keyID)
}

func (s *dashboardService) DeleteAPIKey(ctx context.Context, profileID, keyID string) error {
	return s.repo.DeleteAPIKey(ctx, profileID, keyID)
}

func (s *dashboardService) ValidateAPIKey(ctx context.Context, apiKey string) (*models.APIKey, error) {
	logging.Logger.Info("Validating API key", zap.String("api_key_prefix", apiKey[:16]))

	// Get API key data
	key, err := s.repo.GetAPIKeyByPlainKey(ctx, apiKey)
	if err != nil {
		logging.Logger.Error("API key validation failed", zap.Error(err), zap.String("api_key_prefix", apiKey[:16]))
		return nil, fmt.Errorf("invalid API key")
	}

	// Get user profile to check subscription plan
	profile, err := s.repo.GetUserByID(ctx, key.ProfileID)
	if err != nil {
		logging.Logger.Error("Failed to get user profile for API key", zap.Error(err), zap.String("profile_id", key.ProfileID))
		return nil, fmt.Errorf("invalid API key")
	}

	// Add subscription plan to key data for rate limiting
	key.SubscriptionPlan = profile.SubscriptionPlan

	logging.Logger.Info("API key validated successfully", zap.String("key_id", key.ID), zap.String("profile_id", key.ProfileID), zap.String("subscription", profile.SubscriptionPlan))

	// Update last used timestamp
	if err := s.repo.UpdateAPIKeyLastUsed(ctx, key.ID); err != nil {
		// Log error but don't fail the request
		logging.Logger.Error("Failed to update API key last used timestamp", zap.Error(err))
	}

	return key, nil
}

func (s *dashboardService) CheckAPILimits(ctx context.Context, keyData *models.APIKey) (rateLimited, usageLimitExceeded bool, err error) {
	// Get subscription plan details from config
	plan, exists := s.config.SubscriptionPlans[keyData.SubscriptionPlan]
	if !exists {
		return false, false, fmt.Errorf("unknown subscription plan: %s", keyData.SubscriptionPlan)
	}

	// Check rate limit (hourly and daily limits)
	if plan.RateLimitPerHour > 0 || plan.RateLimitPerDay > 0 {
		isRateLimited, err := s.repo.CheckRateLimit(ctx, keyData.ID, plan.RateLimitPerHour, plan.RateLimitPerDay)
		if err != nil {
			return false, false, fmt.Errorf("failed to check rate limit: %w", err)
		}
		if isRateLimited {
			return true, false, nil
		}
	}

	// Check monthly usage limit
	if plan.APIRequests > 0 { // 0 means unlimited
		monthlyUsage, err := s.repo.GetMonthlyUsage(ctx, keyData.ID)
		if err != nil {
			return false, false, fmt.Errorf("failed to check monthly usage: %w", err)
		}
		if monthlyUsage >= plan.APIRequests {
			return false, true, nil
		}
	}

	return false, false, nil
}

func (s *dashboardService) RecordAPIUsage(ctx context.Context, keyData *models.APIKey, endpoint, method, ipAddress, userAgent string) error {
	// log here
	logging.Logger.Info("Recording API usage", zap.String("key_id", keyData.ID), zap.String("endpoint", endpoint), zap.String("method", method), zap.String("ip_address", ipAddress), zap.String("user_agent", userAgent))

	usage := &models.APIKeyUsage{
		KeyID:     keyData.ID,
		ProfileID: keyData.ProfileID,
		Endpoint:  endpoint,
		Method:    method,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	return s.repo.RecordAPIUsage(ctx, usage)
}

func (s *dashboardService) GetAPIUsage(ctx context.Context, profileID string, limit int, offset int) ([]models.APIKeyUsage, error) {
	return s.repo.GetAPIUsage(ctx, profileID, limit, offset)
}

func (s *dashboardService) CreatePayment(ctx context.Context, profileID string, req *models.CreatePaymentRequest) (*models.Payment, error) {
	// This is a placeholder implementation
	// In production, you would integrate with a payment processor like Stripe

	// Generate a unique payment ID
	paymentID, err := s.generatePaymentID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate payment ID: %w", err)
	}

	payment := &models.Payment{
		ID:               paymentID,
		ProfileID:        profileID,
		Amount:           req.Amount,
		Currency:         req.Currency,
		Status:           "pending",
		SubscriptionType: req.SubscriptionType,
		PaymentMethod:    req.PaymentMethod,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.CreatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}

func (s *dashboardService) GetPaymentHistory(ctx context.Context, profileID string, page, pageSize int) (*models.PaymentHistoryResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	payments, totalCount, err := s.repo.GetPaymentHistory(ctx, profileID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &models.PaymentHistoryResponse{
		Payments:   payments,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// Email verification registration methods

func (s *dashboardService) InitiateRegistration(ctx context.Context, req *models.InitiateRegistrationRequest) (*models.InitiateRegistrationResponse, error) {
	// Check if email already exists
	_, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, fmt.Errorf("email already registered")
	}

	// Generate verification token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// Store token with 24 hour expiration
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.repo.CreateVerificationToken(ctx, token, req.Email, expiresAt); err != nil {
		return nil, fmt.Errorf("failed to create verification token: %w", err)
	}

	// Send verification email
	emailService := utils.NewEmailService()
	if err := emailService.SendVerificationEmail(req.Email, token); err != nil {
		logging.Logger.Error("Failed to send verification email",
			zap.Error(err),
			zap.String("email", req.Email))
		// Don't fail the request if email sending fails
		// The token is still valid and can be used
	}

	return &models.InitiateRegistrationResponse{
		Message: "Verification email sent successfully",
		Email:   req.Email,
	}, nil
}

func (s *dashboardService) VerifyToken(ctx context.Context, token string) (*models.VerifyTokenResponse, error) {
	verificationToken, err := s.repo.GetVerificationToken(ctx, token)
	if err != nil {
		return &models.VerifyTokenResponse{
			Valid: false,
			Error: "Invalid or expired token",
		}, nil
	}

	// Check if token is expired
	if time.Now().After(verificationToken.ExpiresAt) {
		return &models.VerifyTokenResponse{
			Valid: false,
			Error: "Token has expired",
		}, nil
	}

	// Check if token is already used
	if verificationToken.Used {
		return &models.VerifyTokenResponse{
			Valid: false,
			Error: "Token has already been used",
		}, nil
	}

	return &models.VerifyTokenResponse{
		Valid: true,
		Email: verificationToken.Email,
	}, nil
}

func (s *dashboardService) CompleteRegistration(ctx context.Context, req *models.CompleteRegistrationRequest) (*models.CompleteRegistrationResponse, error) {
	// Verify token
	verificationToken, err := s.repo.GetVerificationToken(ctx, req.Token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token")
	}

	// Check if token is expired
	if time.Now().After(verificationToken.ExpiresAt) {
		return nil, fmt.Errorf("token has expired")
	}

	// Check if token is already used
	if verificationToken.Used {
		return nil, fmt.Errorf("token has already been used")
	}

	// Check if email is already registered
	_, err = s.repo.GetUserByEmail(ctx, verificationToken.Email)
	if err == nil {
		return nil, fmt.Errorf("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user with all required fields
	signupReq := &models.SignUpRequest{
		Name:        req.Name,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       verificationToken.Email,
		Password:    string(hashedPassword),
		Description: req.Description,
		Website:     req.Website,
	}

	// Create user profile (already hashed password)
	profile, err := s.repo.CreateUserWithHashedPassword(ctx, signupReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Mark token as used
	if err := s.repo.MarkTokenAsUsed(ctx, req.Token); err != nil {
		logging.Logger.Error("Failed to mark token as used",
			zap.Error(err),
			zap.String("token", req.Token))
		// Don't fail the request - user is created successfully
	}

	// Send welcome email
	emailService := utils.NewEmailService()
	if err := emailService.SendWelcomeEmail(profile.Email, profile.Name); err != nil {
		logging.Logger.Error("Failed to send welcome email",
			zap.Error(err),
			zap.String("email", profile.Email))
		// Don't fail the request if email sending fails
	}

	return &models.CompleteRegistrationResponse{
		ID:      profile.ID,
		Email:   profile.Email,
		Message: "Registration completed successfully",
	}, nil
}

// Password reset methods

func (s *dashboardService) InitiatePasswordReset(ctx context.Context, req *models.ForgotPasswordRequest) (*models.ForgotPasswordResponse, error) {
	// Check if user exists
	_, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// For security, don't reveal if email exists or not
		// Just return success message
		return &models.ForgotPasswordResponse{
			Message: "If an account exists with this email, a password reset link has been sent",
			Email:   req.Email,
		}, nil
	}

	// Generate reset token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// Store token with 24 hour expiration
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.repo.CreatePasswordResetToken(ctx, token, req.Email, expiresAt); err != nil {
		return nil, fmt.Errorf("failed to create reset token: %w", err)
	}

	// Send password reset email
	emailService := utils.NewEmailService()
	if err := emailService.SendPasswordResetEmail(req.Email, token); err != nil {
		logging.Logger.Error("Failed to send password reset email",
			zap.Error(err),
			zap.String("email", req.Email))
		// Don't fail the request if email sending fails
		// The token is still valid and can be used
	}

	return &models.ForgotPasswordResponse{
		Message: "If an account exists with this email, a password reset link has been sent",
		Email:   req.Email,
	}, nil
}

func (s *dashboardService) VerifyResetToken(ctx context.Context, token string) (*models.VerifyResetTokenResponse, error) {
	// Get token from database
	resetToken, err := s.repo.GetVerificationToken(ctx, token)
	if err != nil {
		return &models.VerifyResetTokenResponse{
			Valid: false,
			Error: "Invalid or expired reset token",
		}, nil
	}

	// Check if token is of the correct type
	if resetToken.Type != "password_reset" {
		return &models.VerifyResetTokenResponse{
			Valid: false,
			Error: "Invalid token type",
		}, nil
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		return &models.VerifyResetTokenResponse{
			Valid: false,
			Error: "Reset token has expired",
		}, nil
	}

	// Check if token has been used
	if resetToken.Used {
		return &models.VerifyResetTokenResponse{
			Valid: false,
			Error: "Reset token has already been used",
		}, nil
	}

	return &models.VerifyResetTokenResponse{
		Valid: true,
		Email: resetToken.Email,
	}, nil
}

func (s *dashboardService) ResetPassword(ctx context.Context, req *models.ResetPasswordRequest) (*models.ResetPasswordResponse, error) {
	// Verify token
	resetToken, err := s.repo.GetVerificationToken(ctx, req.Token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token")
	}

	// Check if token is of the correct type
	if resetToken.Type != "password_reset" {
		return nil, fmt.Errorf("invalid token type")
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		return nil, fmt.Errorf("token has expired")
	}

	// Check if token is already used
	if resetToken.Used {
		return nil, fmt.Errorf("token has already been used")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user's password
	if err := s.repo.UpdateUserPassword(ctx, resetToken.Email, string(hashedPassword)); err != nil {
		return nil, fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	if err := s.repo.MarkTokenAsUsed(ctx, req.Token); err != nil {
		logging.Logger.Error("Failed to mark reset token as used",
			zap.Error(err),
			zap.String("token", req.Token))
		// Don't fail the request - password is already updated
	}

	return &models.ResetPasswordResponse{
		Message: "Password reset successfully",
	}, nil
}

// Password change method

func (s *dashboardService) ChangePassword(ctx context.Context, userID string, req *models.ChangePasswordRequest) (*models.ChangePasswordResponse, error) {
	// Get user to verify current password
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		return nil, fmt.Errorf("current password is incorrect")
	}

	// Check if new password is different from current
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.NewPassword)); err == nil {
		return nil, fmt.Errorf("new password must be different from current password")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.repo.ChangeUserPassword(ctx, userID, user.Password, string(hashedPassword)); err != nil {
		return nil, fmt.Errorf("failed to update password: %w", err)
	}

	logging.Logger.Info("Password changed successfully",
		zap.String("user_id", userID),
		zap.String("email", user.Email))

	// Send password changed notification email
	emailService := utils.NewEmailService()
	userName := user.Name
	if userName == "" {
		userName = user.FirstName + " " + user.LastName
	}
	if userName == "" {
		userName = user.Email
	}
	
	if err := emailService.SendPasswordChangedEmail(user.Email, userName); err != nil {
		logging.Logger.Error("Failed to send password changed email",
			zap.Error(err),
			zap.String("email", user.Email))
		// Don't fail the request if email sending fails
		// The password was already changed successfully
	}

	return &models.ChangePasswordResponse{
		Message: "Password changed successfully",
	}, nil
}

// Subscription activation methods

func (s *dashboardService) ActivateSubscription(ctx context.Context, req *models.ActivateSubscriptionRequest) (*models.ActivateSubscriptionResponse, error) {
	// Determine the full plan ID based on billing cycle
	fullPlanID := req.PlanID
	if req.PlanID != "free" && req.PlanID != "enterprise" {
		fullPlanID = fmt.Sprintf("%s_%s", req.PlanID, req.BillingCycle)
	}

	// Get plan configuration
	plan, exists := s.config.SubscriptionPlans[fullPlanID]
	if !exists {
		return nil, fmt.Errorf("invalid plan: %s", fullPlanID)
	}

	// Calculate subscription expiry based on plan duration
	var expiresAt *time.Time
	if plan.SubscriptionDuration > 0 {
		expires := time.Now().Add(time.Duration(plan.SubscriptionDuration) * 24 * time.Hour)
		expiresAt = &expires
	}

	// Update user subscription
	if err := s.repo.UpdateUserSubscription(ctx, req.UserID, req.PlanID, req.BillingCycle, expiresAt); err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	// Store payment record
	payment := &models.Payment{
		ID:               req.PaymentID,
		ProfileID:        req.UserID,
		Amount:           req.AmountPaid,
		Currency:         "USD",
		SubscriptionType: req.PlanID,
		PaymentMethod:    fmt.Sprintf("nowpayments_%s", req.PayCurrency),
		Status:           "confirmed",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.StoreNOWPayment(ctx, payment); err != nil {
		logging.Logger.Error("Failed to store payment",
			zap.Error(err),
			zap.String("payment_id", req.PaymentID))
		// Don't fail subscription activation if payment storage fails
	}

	// Create invoice for this payment immediately
	var expiresAtTime time.Time
	if expiresAt != nil {
		expiresAtTime = *expiresAt
	} else {
		expiresAtTime = time.Now().Add(30 * 24 * time.Hour) // Default to 30 days
	}
	if err := s.createInvoiceForPayment(ctx, req.UserID, req.PlanID, req.BillingCycle, req.PaymentID, req.AmountPaid, "USD", expiresAtTime); err != nil {
		logging.Logger.Error("Failed to create invoice for payment",
			zap.Error(err),
			zap.String("user_id", req.UserID),
			zap.String("payment_id", req.PaymentID))
		// Don't fail subscription activation if invoice creation fails
	}

	// Mark any pending invoices as paid (for recurring payments)
	if err := s.markInvoicesAsPaid(ctx, req.UserID, req.PaymentID, req.AmountPaid, "USD"); err != nil {
		logging.Logger.Error("Failed to mark invoices as paid",
			zap.Error(err),
			zap.String("user_id", req.UserID),
			zap.String("payment_id", req.PaymentID))
		// Don't fail subscription activation if invoice marking fails
	}

	// Get user profile for email
	profile, err := s.repo.GetUserByID(ctx, req.UserID)
	if err != nil {
		logging.Logger.Error("Failed to get user profile for subscription email",
			zap.Error(err),
			zap.String("user_id", req.UserID))
	} else {
		// Send subscription activation email
		emailService := utils.NewEmailService()
		userName := profile.Name
		if userName == "" {
			userName = profile.FirstName + " " + profile.LastName
		}
		if userName == "" {
			userName = profile.Email
		}

		if err := emailService.SendSubscriptionActivatedEmail(profile.Email, userName, req.PlanID, req.BillingCycle, expiresAt); err != nil {
			logging.Logger.Error("Failed to send subscription activation email",
				zap.Error(err),
				zap.String("email", profile.Email))
		}
	}

	logging.Logger.Info("Subscription activated successfully",
		zap.String("user_id", req.UserID),
		zap.String("plan_id", req.PlanID),
		zap.String("billing_cycle", req.BillingCycle),
		zap.String("payment_id", req.PaymentID))

	return &models.ActivateSubscriptionResponse{
		Message:               "Subscription activated successfully",
		SubscriptionPlan:      req.PlanID,
		BillingCycle:          req.BillingCycle,
		SubscriptionExpiresAt: expiresAt,
	}, nil
}

// createInvoiceForPayment creates a paid invoice immediately when a payment is completed
func (s *dashboardService) createInvoiceForPayment(ctx context.Context, userID string, planID string, billingCycle string, paymentID string, amount float64, currency string, expiresAt time.Time) error {
	// Generate invoice number
	invoiceNumber := fmt.Sprintf("INV-%s-%d", userID[:8], time.Now().Unix())
	
	// Convert amount to cents
	amountCents := int64(amount * 100)
	
	// Create invoice metadata
	metadata := map[string]interface{}{
		"plan_id":         planID,
		"billing_cycle":   billingCycle,
		"payment_method":  "paystack",
		"subscription_id": paymentID,
		"created_from":    "payment_completion",
	}
	
	// Create the invoice model
	invoice := &models.Invoice{
		ID:            uuid.New().String(),
		InvoiceNumber: invoiceNumber,
		AccountID:     userID,
		Amount:        amountCents,
		Currency:      currency,
		DueDate:       time.Now(), // Due date is now since payment is completed
		IssuedAt:      time.Now(),
		Status:        "pending",
		Metadata:      metadata,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	// Create the invoice
	err := s.repo.CreateInvoice(ctx, invoice)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}
	
	// Mark invoice as paid immediately
	paidAt := time.Now()
	if err := s.repo.UpdateInvoiceStatus(ctx, invoice.ID, "paid", &paymentID, &paidAt); err != nil {
		logging.Logger.Error("Failed to mark invoice as paid",
			zap.String("invoice_id", invoice.ID),
			zap.String("payment_id", paymentID),
			zap.Error(err))
		return fmt.Errorf("failed to mark invoice as paid: %w", err)
	}
	
	logging.Logger.Info("Invoice created and marked as paid",
		zap.String("invoice_id", invoice.ID),
		zap.String("invoice_number", invoice.InvoiceNumber),
		zap.String("user_id", userID),
		zap.String("payment_id", paymentID),
		zap.Float64("amount", amount))
	
	return nil
}

// markInvoicesAsPaid marks pending invoices as paid when a payment is successful
func (s *dashboardService) markInvoicesAsPaid(ctx context.Context, userID string, paymentID string, amount float64, currency string) error {
	// Find pending invoices for this user
	// We'll look for invoices due around now (within a reasonable time window)
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	invoices, err := s.repo.GetInvoicesForPayment(ctx, userID, startOfDay)
	if err != nil {
		return fmt.Errorf("failed to get invoices for payment: %w", err)
	}

	if len(invoices) == 0 {
		logging.Logger.Info("No pending invoices found for payment",
			zap.String("user_id", userID),
			zap.String("payment_id", paymentID),
		)
		return nil
	}

	// Mark invoices as paid
	paidAt := time.Now()
	for _, invoice := range invoices {
		if err := s.repo.UpdateInvoiceStatus(ctx, invoice.ID, "paid", &paymentID, &paidAt); err != nil {
			logging.Logger.Error("Failed to update invoice status",
				zap.String("invoice_id", invoice.ID),
				zap.String("payment_id", paymentID),
				zap.Error(err),
			)
			continue
		}

		logging.Logger.Info("Invoice marked as paid",
			zap.String("invoice_id", invoice.ID),
			zap.String("invoice_number", invoice.InvoiceNumber),
			zap.String("payment_id", paymentID),
		)
	}

	return nil
}

func (s *dashboardService) StoreNOWPayment(ctx context.Context, req *models.StorePaymentRequest) (*models.PaymentStorageResponse, error) {
	// Extract user ID from order ID if needed
	// Order ID format: sub_{planId}_{billingFreq}_{userId}_{uuid}
	orderParts := strings.Split(req.OrderID, "_")
	userID := ""
	if len(orderParts) >= 4 {
		userID = orderParts[3]
	}

	payment := &models.Payment{
		ID:               req.PaymentID,
		ProfileID:        userID,
		Amount:           req.Amount,
		Currency:         req.Currency,
		SubscriptionType: orderParts[1], // Plan ID from order ID
		PaymentMethod:    fmt.Sprintf("nowpayments_%s", req.PayCurrency),
		Status:           req.Status,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.StoreNOWPayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to store payment: %w", err)
	}

	logging.Logger.Info("NOWPayment stored successfully",
		zap.String("payment_id", req.PaymentID),
		zap.String("status", req.Status))

	return &models.PaymentStorageResponse{
		Message:   "Payment stored successfully",
		PaymentID: req.PaymentID,
	}, nil
}

func (s *dashboardService) UpdatePaymentStatus(ctx context.Context, req *models.UpdatePaymentStatusRequest) error {
	if err := s.repo.UpdatePaymentStatus(ctx, req.PaymentID, req.Status); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	logging.Logger.Info("Payment status updated",
		zap.String("payment_id", req.PaymentID),
		zap.String("status", req.Status))

	return nil
}

// Helper methods

func (s *dashboardService) generateJWT(userID, email string) (string, int64, error) {
	if s.jwtSecret == "" {
		return "", 0, errors.New("JWT secret not configured")
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     expirationTime.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expirationTime.Unix(), nil
}

func (s *dashboardService) generatePaymentID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("pay_%s", hex.EncodeToString(bytes)), nil
}
