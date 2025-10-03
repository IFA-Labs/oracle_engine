package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"oracle_engine/internal/database/timescale"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type DashboardRepository interface {
	// User management
	CreateUser(ctx context.Context, req *models.SignUpRequest) (*models.CompanyProfile, error)
	GetUserByEmail(ctx context.Context, email string) (*models.CompanyProfile, error)
	GetUserByID(ctx context.Context, id string) (*models.CompanyProfile, error)
	UpdateProfile(ctx context.Context, id string, req *models.UpdateProfileRequest) (*models.CompanyProfile, error)
	UpdateSubscription(ctx context.Context, id string, subscriptionPlan string) (*models.CompanyProfile, error)

	// API Key management
	HashAPIKey(apiKey string) (string, error)
	CreateAPIKey(ctx context.Context, profileID string, req *models.CreateAPIKeyRequest) (*models.APIKey, error)
	GetAPIKeys(ctx context.Context, profileID string) ([]models.APIKey, error)
	GetAPIKeyByID(ctx context.Context, profileID, keyID string) (*models.APIKey, error)
	GetAPIKeyByHash(ctx context.Context, keyHash string) (*models.APIKey, error)
	GetAPIKeyByPlainKey(ctx context.Context, apiKey string) (*models.APIKey, error)
	DeleteAPIKey(ctx context.Context, profileID, keyID string) error
	UpdateAPIKeyLastUsed(ctx context.Context, keyID string) error

	// API Usage tracking
	RecordAPIUsage(ctx context.Context, usage *models.APIKeyUsage) error
	GetAPIUsage(ctx context.Context, profileID string, limit int, offset int) ([]models.APIKeyUsage, error)
	GetMonthlyUsage(ctx context.Context, keyID string) (int64, error)
	GetDailyUsage(ctx context.Context, keyID string) (int64, error)
	GetHourlyUsage(ctx context.Context, keyID string) (int64, error)
	CheckRateLimit(ctx context.Context, keyID string, rateLimitPerHour, rateLimitPerDay int) (bool, error)

	// Payment management (basic structure for future implementation)
	CreatePayment(ctx context.Context, payment *models.Payment) error
	GetPaymentHistory(ctx context.Context, profileID string, limit int, offset int) ([]models.Payment, int64, error)
	UpdatePaymentStatus(ctx context.Context, paymentID, status string) error
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB, encryptionKey string) DashboardRepository {
	return &dashboardRepository{
		db: db,
	}
}


func (r *dashboardRepository) CreateUser(ctx context.Context, req *models.SignUpRequest) (*models.CompanyProfile, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	profile := timescale.CompanyProfile{
		ID:               uuid.New().String(),
		Name:             req.Name,
		Description:      req.Description,
		Website:          req.Website,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Email:            req.Email,
		Password:         string(hashedPassword),
		SubscriptionPlan: "free", // Default to free tier
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := r.db.WithContext(ctx).Create(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, fmt.Errorf("email already exists")
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return r.gormToModel(profile), nil
}

func (r *dashboardRepository) GetUserByEmail(ctx context.Context, email string) (*models.CompanyProfile, error) {
	var profile timescale.CompanyProfile
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return r.gormToModel(profile), nil
}

func (r *dashboardRepository) GetUserByID(ctx context.Context, id string) (*models.CompanyProfile, error) {
	var profile timescale.CompanyProfile
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return r.gormToModel(profile), nil
}

func (r *dashboardRepository) UpdateProfile(ctx context.Context, id string, req *models.UpdateProfileRequest) (*models.CompanyProfile, error) {
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Website != nil {
		updates["website"] = *req.Website
	}
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.LogoURL != nil {
		updates["logo_url"] = *req.LogoURL
	}

	updates["updated_at"] = time.Now()

	if err := r.db.WithContext(ctx).Model(&timescale.CompanyProfile{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return r.GetUserByID(ctx, id)
}

func (r *dashboardRepository) UpdateSubscription(ctx context.Context, id string, subscriptionPlan string) (*models.CompanyProfile, error) {
	updates := map[string]interface{}{
		"subscription_plan": subscriptionPlan,
		"updated_at":        time.Now(),
	}

	if err := r.db.WithContext(ctx).Model(&timescale.CompanyProfile{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	// Log the subscription update
	logging.Logger.Info("Subscription plan updated", 
		zap.String("user_id", id), 
		zap.String("new_plan", subscriptionPlan))

	return r.GetUserByID(ctx, id)
}

func (r *dashboardRepository) HashAPIKey(apiKey string) (string, error) {
	hashedKey, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash API key: %w", err)
	}
	return string(hashedKey), nil
}

func (r *dashboardRepository) CreateAPIKey(ctx context.Context, profileID string, req *models.CreateAPIKeyRequest) (*models.APIKey, error) {
	// Generate API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	apiKey := fmt.Sprintf("ifa_%s", hex.EncodeToString(keyBytes))

	// Extract prefix for fast lookup (first 16 characters)
	keyPrefix := apiKey[:16]

	// Hash the key for storage (for security)
	hashedKey, err := r.HashAPIKey(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to hash API key: %w", err)
	}

	now := time.Now()
	dbAPIKey := timescale.DashboardAPIKey{
		ID:           uuid.New().String(),
		ProfileID:    profileID,
		Name:         req.Name,
		KeyPrefix:    keyPrefix,
		KeyHash:      string(hashedKey),
		KeyEncrypted: apiKey, // Store plain text instead of encrypted
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := r.db.WithContext(ctx).Create(&dbAPIKey).Error; err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return &models.APIKey{
		ID:        dbAPIKey.ID,
		Key:       apiKey, // Only returned on creation
		Name:      dbAPIKey.Name,
		IsActive:  dbAPIKey.IsActive,
		CreatedAt: dbAPIKey.CreatedAt,
		UpdatedAt: dbAPIKey.UpdatedAt,
		LastUsed:  dbAPIKey.LastUsed,
	}, nil
}

func (r *dashboardRepository) GetAPIKeys(ctx context.Context, profileID string) ([]models.APIKey, error) {
	var dbKeys []timescale.DashboardAPIKey
	if err := r.db.WithContext(ctx).Where("profile_id = ?", profileID).Find(&dbKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to get API keys: %w", err)
	}

	keys := make([]models.APIKey, len(dbKeys))
	for i, dbKey := range dbKeys {
		// Return the plain text key directly
		keys[i] = models.APIKey{
			ID:        dbKey.ID,
			Key:       dbKey.KeyEncrypted, // Now contains plain text
			Name:      dbKey.Name,
			IsActive:  dbKey.IsActive,
			CreatedAt: dbKey.CreatedAt,
			UpdatedAt: dbKey.UpdatedAt,
			LastUsed:  dbKey.LastUsed,
		}
	}

	return keys, nil
}

func (r *dashboardRepository) GetAPIKeyByID(ctx context.Context, profileID, keyID string) (*models.APIKey, error) {
	var dbKey timescale.DashboardAPIKey
	if err := r.db.WithContext(ctx).Where("id = ? AND profile_id = ?", keyID, profileID).First(&dbKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return &models.APIKey{
		ID:        dbKey.ID,
		Key:       dbKey.KeyEncrypted, // Now contains plain text
		Name:      dbKey.Name,
		IsActive:  dbKey.IsActive,
		CreatedAt: dbKey.CreatedAt,
		UpdatedAt: dbKey.UpdatedAt,
		LastUsed:  dbKey.LastUsed,
	}, nil
}

func (r *dashboardRepository) GetAPIKeyByPlainKey(ctx context.Context, apiKey string) (*models.APIKey, error) {
	// Check if the API key is long enough to have a prefix
	if len(apiKey) < 16 {
		return nil, fmt.Errorf("invalid API key format")
	}

	// Extract prefix for fast lookup (first 16 characters)
	keyPrefix := apiKey[:16]

	// First, find keys with matching prefix for efficient filtering
	var dbKeys []timescale.DashboardAPIKey
	if err := r.db.WithContext(ctx).Where("key_prefix = ? AND is_active = true", keyPrefix).Find(&dbKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to get API keys: %w", err)
	}

	// If no keys found with this prefix, return not found
	if len(dbKeys) == 0 {
		return nil, fmt.Errorf("API key not found")
	}

	// Compare the provided key with each stored hash from the matching prefix
	for _, dbKey := range dbKeys {
		if err := bcrypt.CompareHashAndPassword([]byte(dbKey.KeyHash), []byte(apiKey)); err == nil {
			// Found matching key
			return &models.APIKey{
				ID:        dbKey.ID,
				ProfileID: dbKey.ProfileID,
				Name:      dbKey.Name,
				IsActive:  dbKey.IsActive,
				CreatedAt: dbKey.CreatedAt,
				UpdatedAt: dbKey.UpdatedAt,
				LastUsed:  dbKey.LastUsed,
				KeyHash:   dbKey.KeyHash,
			}, nil
		}
	}

	return nil, fmt.Errorf("API key not found")
}

func (r *dashboardRepository) GetAPIKeyByHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	var dbKey timescale.DashboardAPIKey
	if err := r.db.WithContext(ctx).Where("key_hash = ? AND is_active = true", keyHash).First(&dbKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return &models.APIKey{
		ID:        dbKey.ID,
		Name:      dbKey.Name,
		IsActive:  dbKey.IsActive,
		CreatedAt: dbKey.CreatedAt,
		UpdatedAt: dbKey.UpdatedAt,
		LastUsed:  dbKey.LastUsed,
		KeyHash:   dbKey.KeyHash,
	}, nil
}

func (r *dashboardRepository) DeleteAPIKey(ctx context.Context, profileID, keyID string) error {
	result := r.db.WithContext(ctx).Where("id = ? AND profile_id = ?", keyID, profileID).Delete(&timescale.DashboardAPIKey{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete API key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}
	return nil
}

func (r *dashboardRepository) UpdateAPIKeyLastUsed(ctx context.Context, keyID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&timescale.DashboardAPIKey{}).
		Where("id = ?", keyID).
		Update("last_used", now).Error
}

func (r *dashboardRepository) RecordAPIUsage(ctx context.Context, usage *models.APIKeyUsage) error {
	logging.Logger.Info("Recording API usage", zap.String("key_id", usage.KeyID), zap.String("endpoint", usage.Endpoint), zap.String("method", usage.Method), zap.String("ip_address", usage.IPAddress), zap.String("user_agent", usage.UserAgent))

	dbUsage := timescale.DashboardAPIKeyUsage{
		ID:        uuid.New().String(),
		ProfileID: usage.ProfileID,
		KeyID:     usage.KeyID,
		Endpoint:  usage.Endpoint,
		Method:    usage.Method,
		IPAddress: usage.IPAddress,
		UserAgent: usage.UserAgent,
		CreatedAt: time.Now(),
	}

	err := r.db.WithContext(ctx).Create(&dbUsage).Error
	if err != nil {
		logging.Logger.Error("Failed to record API usage", zap.Error(err), zap.String("key_id", usage.KeyID))
		return err
	}

	logging.Logger.Info("Successfully recorded API usage", zap.String("usage_id", dbUsage.ID), zap.String("key_id", usage.KeyID))
	return nil
}

func (r *dashboardRepository) GetAPIUsage(ctx context.Context, profileID string, limit int, offset int) ([]models.APIKeyUsage, error) {
	var dbUsage []timescale.DashboardAPIKeyUsage
	query := r.db.WithContext(ctx).Where("profile_id = ?", profileID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&dbUsage).Error; err != nil {
		return nil, fmt.Errorf("failed to get API usage: %w", err)
	}

	usage := make([]models.APIKeyUsage, len(dbUsage))
	for i, u := range dbUsage {
		usage[i] = models.APIKeyUsage{
			ID:        u.ID,
			KeyID:     u.KeyID,
			ProfileID: u.ProfileID,
			Endpoint:  u.Endpoint,
			Method:    u.Method,
			IPAddress: u.IPAddress,
			UserAgent: u.UserAgent,
			CreatedAt: u.CreatedAt,
		}
	}

	return usage, nil
}

func (r *dashboardRepository) GetMonthlyUsage(ctx context.Context, keyID string) (int64, error) {
	var count int64
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	err := r.db.WithContext(ctx).Model(&timescale.DashboardAPIKeyUsage{}).
		Where("key_id = ? AND created_at >= ?", keyID, startOfMonth).
		Count(&count).Error

	return count, err
}

func (r *dashboardRepository) GetDailyUsage(ctx context.Context, keyID string) (int64, error) {
	var count int64
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	err := r.db.WithContext(ctx).Model(&timescale.DashboardAPIKeyUsage{}).
		Where("key_id = ? AND created_at >= ?", keyID, startOfDay).
		Count(&count).Error

	return count, err
}

func (r *dashboardRepository) GetHourlyUsage(ctx context.Context, keyID string) (int64, error) {
	var count int64
	now := time.Now()
	startOfHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())

	err := r.db.WithContext(ctx).Model(&timescale.DashboardAPIKeyUsage{}).
		Where("key_id = ? AND created_at >= ?", keyID, startOfHour).
		Count(&count).Error

	return count, err
}

func (r *dashboardRepository) CheckRateLimit(ctx context.Context, keyID string, rateLimitPerHour, rateLimitPerDay int) (bool, error) {
	// If both limits are 0, no rate limiting (enterprise plan)
	if rateLimitPerHour <= 0 && rateLimitPerDay <= 0 {
		return false, nil
	}

	// Check hourly limit if specified
	if rateLimitPerHour > 0 {
		hourlyUsage, err := r.GetHourlyUsage(ctx, keyID)
		if err != nil {
			return false, fmt.Errorf("failed to check hourly usage: %w", err)
		}
		if hourlyUsage >= int64(rateLimitPerHour) {
			return true, nil // Rate limited by hourly limit
		}
	}

	// Check daily limit if specified
	if rateLimitPerDay > 0 {
		dailyUsage, err := r.GetDailyUsage(ctx, keyID)
		if err != nil {
			return false, fmt.Errorf("failed to check daily usage: %w", err)
		}
		if dailyUsage >= int64(rateLimitPerDay) {
			return true, nil // Rate limited by daily limit
		}
	}

	return false, nil // Not rate limited
}

func (r *dashboardRepository) CreatePayment(ctx context.Context, payment *models.Payment) error {
	dbPayment := timescale.DashboardPayment{
		ID:               payment.ID,
		ProfileID:        payment.ProfileID, // Fixed - should be payment.ProfileID
		Amount:           payment.Amount,
		Currency:         payment.Currency,
		Status:           payment.Status,
		SubscriptionType: payment.SubscriptionType,
		PaymentMethod:    payment.PaymentMethod,
		CreatedAt:        payment.CreatedAt,
		UpdatedAt:        payment.UpdatedAt,
	}

	return r.db.WithContext(ctx).Create(&dbPayment).Error
}

func (r *dashboardRepository) GetPaymentHistory(ctx context.Context, profileID string, limit int, offset int) ([]models.Payment, int64, error) {
	var dbPayments []timescale.DashboardPayment
	var totalCount int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&timescale.DashboardPayment{}).Where("profile_id = ?", profileID).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payments: %w", err)
	}

	// Get payments with pagination
	query := r.db.WithContext(ctx).Where("profile_id = ?", profileID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&dbPayments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get payment history: %w", err)
	}

	payments := make([]models.Payment, len(dbPayments))
	for i, p := range dbPayments {
		payments[i] = models.Payment{
			ID:               p.ID,
			ProfileID:        p.ProfileID,
			Amount:           p.Amount,
			Currency:         p.Currency,
			Status:           p.Status,
			SubscriptionType: p.SubscriptionType,
			PaymentMethod:    p.PaymentMethod,
			CreatedAt:        p.CreatedAt,
			UpdatedAt:        p.UpdatedAt,
		}
	}

	return payments, totalCount, nil
}

func (r *dashboardRepository) UpdatePaymentStatus(ctx context.Context, paymentID, status string) error {
	return r.db.WithContext(ctx).Model(&timescale.DashboardPayment{}).
		Where("id = ?", paymentID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// Helper function to convert GORM model to domain model
func (r *dashboardRepository) gormToModel(profile timescale.CompanyProfile) *models.CompanyProfile {
	return &models.CompanyProfile{
		ID:               profile.ID,
		Name:             profile.Name,
		Description:      profile.Description,
		Website:          profile.Website,
		LogoURL:          profile.LogoURL,
		FirstName:        profile.FirstName,
		LastName:         profile.LastName,
		Email:            profile.Email,
		Password:         profile.Password,
		SubscriptionPlan: profile.SubscriptionPlan,
		CreatedAt:        profile.CreatedAt,
		UpdatedAt:        profile.UpdatedAt,
	}
}
