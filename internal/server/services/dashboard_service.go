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
	"time"

	"go.uber.org/zap"

	"github.com/golang-jwt/jwt/v5"
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

	// Profile management
	GetProfile(ctx context.Context, id string) (*models.CompanyProfile, error)
	UpdateProfile(ctx context.Context, id string, req *models.UpdateProfileRequest) (*models.CompanyProfile, error)

	// API Key management
	CreateAPIKey(ctx context.Context, profileID string, req *models.CreateAPIKeyRequest) (*models.CreateAPIKeyResponse, error)
	GetAPIKeys(ctx context.Context, profileID string) ([]models.APIKey, error)
	DeleteAPIKey(ctx context.Context, profileID, keyID string) error
	ValidateAPIKey(ctx context.Context, apiKey string) (*models.APIKey, error)
	CheckAPILimits(ctx context.Context, keyData *models.APIKey) (rateLimited, usageLimitExceeded bool, err error)
	RecordAPIUsage(ctx context.Context, keyData *models.APIKey, endpoint, method, ipAddress, userAgent string) error
	GetAPIUsage(ctx context.Context, profileID string, limit int, offset int) ([]models.APIKeyUsage, error)

	// Payment management (placeholder for future implementation)
	CreatePayment(ctx context.Context, profileID string, req *models.CreatePaymentRequest) (*models.Payment, error)
	GetPaymentHistory(ctx context.Context, profileID string, page, pageSize int) (*models.PaymentHistoryResponse, error)
}

func NewDashboardService(repo repository.DashboardRepository, jwtSecret string, cfg *config.Config) DashboardService {
	return &dashboardService{
		repo:      repo,
		jwtSecret: jwtSecret,
		config:    cfg,
	}
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

	// Verify password
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

func (s *dashboardService) CreateAPIKey(ctx context.Context, profileID string, req *models.CreateAPIKeyRequest) (*models.CreateAPIKeyResponse, error) {
	apiKey, err := s.repo.CreateAPIKey(ctx, profileID, req)
	if err != nil {
		return nil, err
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
	// TODO: default allow users for now - remove this when ready to enable limits
	_ = ctx
	_ = keyData
	return false, false, nil

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
