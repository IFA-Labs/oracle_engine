package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"oracle_engine/internal/database/timescale"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DashboardRepository interface {
	// User management
	CreateUser(ctx context.Context, req *models.SignUpRequest) (*models.CompanyProfile, error)
	CreateUserWithHashedPassword(ctx context.Context, req *models.SignUpRequest) (*models.CompanyProfile, error)
	GetUserByEmail(ctx context.Context, email string) (*models.CompanyProfile, error)
	GetUserByID(ctx context.Context, id string) (*models.CompanyProfile, error)
	UpdateProfile(ctx context.Context, id string, req *models.UpdateProfileRequest) (*models.CompanyProfile, error)
	UpdateSubscription(ctx context.Context, id string, subscriptionPlan string) (*models.CompanyProfile, error)
	DeleteUser(ctx context.Context, id string) error

	// Email verification
	CreateVerificationToken(ctx context.Context, token, email string, expiresAt time.Time) error
	GetVerificationToken(ctx context.Context, token string) (*timescale.VerificationToken, error)
	MarkTokenAsUsed(ctx context.Context, token string) error
	
	// Password reset
	CreatePasswordResetToken(ctx context.Context, token, email string, expiresAt time.Time) error
	UpdateUserPassword(ctx context.Context, email, hashedPassword string) error
	
	// Password change (with current password verification)
	ChangeUserPassword(ctx context.Context, userID, currentPasswordHash, newPasswordHash string) error
	
	// Subscription activation
	UpdateUserSubscription(ctx context.Context, userID, planID, billingCycle string, expiresAt *time.Time) error
	
	// Payment storage
	StoreNOWPayment(ctx context.Context, payment *models.Payment) error
	UpdatePaymentStatus(ctx context.Context, paymentID, status string) error
	GetPaymentByID(ctx context.Context, paymentID string) (*models.Payment, error)

	// API Key management
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

	// Invoice management
	CreateInvoice(ctx context.Context, invoice *models.Invoice) error
	GetInvoiceByID(ctx context.Context, invoiceID string) (*models.Invoice, error)
	GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*models.Invoice, error)
	GetInvoicesByAccount(ctx context.Context, accountID string, limit int, offset int) ([]models.Invoice, int64, error)
	GetInvoicesByStatus(ctx context.Context, status string, limit int, offset int) ([]models.Invoice, int64, error)
	GetInvoicesDueSoon(ctx context.Context, daysAhead int) ([]models.Invoice, error)
	UpdateInvoiceStatus(ctx context.Context, invoiceID string, status string, paymentID *string, paidAt *time.Time) error
	CheckInvoiceExists(ctx context.Context, accountID string, dueDate time.Time) (bool, error)
	GetInvoicesForPayment(ctx context.Context, accountID string, dueDate time.Time) ([]models.Invoice, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{
		db: db,
	}
}


func (r *dashboardRepository) CreateUser(ctx context.Context, req *models.SignUpRequest) (*models.CompanyProfile, error) {
	// Hash password for security
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
		Password:         string(hashedPassword), // Store hashed password
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

func (r *dashboardRepository) DeleteUser(ctx context.Context, id string) error {
	// Start a transaction to delete user and all related data
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete all API keys for this user
		if err := tx.Where("profile_id = ?", id).Delete(&timescale.DashboardAPIKey{}).Error; err != nil {
			logging.Logger.Error("Failed to delete API keys", zap.String("user_id", id), zap.Error(err))
			return fmt.Errorf("failed to delete API keys: %w", err)
		}

		// Delete all API usage records
		if err := tx.Where("profile_id = ?", id).Delete(&timescale.DashboardAPIKeyUsage{}).Error; err != nil {
			logging.Logger.Error("Failed to delete API usage records", zap.String("user_id", id), zap.Error(err))
			return fmt.Errorf("failed to delete API usage records: %w", err)
		}

		// Delete all payments
		if err := tx.Where("profile_id = ?", id).Delete(&timescale.DashboardPayment{}).Error; err != nil {
			logging.Logger.Error("Failed to delete payments", zap.String("user_id", id), zap.Error(err))
			return fmt.Errorf("failed to delete payments: %w", err)
		}

		// Finally, delete the user profile
		if err := tx.Where("id = ?", id).Delete(&timescale.CompanyProfile{}).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("user not found")
			}
			logging.Logger.Error("Failed to delete user profile", zap.String("user_id", id), zap.Error(err))
			return fmt.Errorf("failed to delete user: %w", err)
		}

		logging.Logger.Info("User account deleted successfully", zap.String("user_id", id))
		return nil
	})
}

// CreateUserWithHashedPassword creates a user with an already hashed password (used for email verification flow)
func (r *dashboardRepository) CreateUserWithHashedPassword(ctx context.Context, req *models.SignUpRequest) (*models.CompanyProfile, error) {
	now := time.Now()
	profile := timescale.CompanyProfile{
		ID:               uuid.New().String(),
		Name:             req.Name,
		Description:      req.Description,
		Website:          req.Website,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Email:            req.Email,
		Password:         req.Password, // Already hashed
		EmailVerified:    true,         // Email is verified through the token flow
		SubscriptionPlan: "free",        // Default to free tier
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

// CreateVerificationToken creates a new email verification token
func (r *dashboardRepository) CreateVerificationToken(ctx context.Context, token, email string, expiresAt time.Time) error {
	verificationToken := timescale.VerificationToken{
		Token:     token,
		Email:     email,
		Type:      "email_verification",
		Used:      false,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	if err := r.db.WithContext(ctx).Create(&verificationToken).Error; err != nil {
		return fmt.Errorf("failed to create verification token: %w", err)
	}

	return nil
}

// GetVerificationToken retrieves a verification token
func (r *dashboardRepository) GetVerificationToken(ctx context.Context, token string) (*timescale.VerificationToken, error) {
	var verificationToken timescale.VerificationToken
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&verificationToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("failed to get verification token: %w", err)
	}

	return &verificationToken, nil
}

// MarkTokenAsUsed marks a verification token as used
func (r *dashboardRepository) MarkTokenAsUsed(ctx context.Context, token string) error {
	if err := r.db.WithContext(ctx).Model(&timescale.VerificationToken{}).
		Where("token = ?", token).
		Update("used", true).Error; err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// CreatePasswordResetToken creates a new password reset token
func (r *dashboardRepository) CreatePasswordResetToken(ctx context.Context, token, email string, expiresAt time.Time) error {
	resetToken := timescale.VerificationToken{
		Token:     token,
		Email:     email,
		Type:      "password_reset",
		Used:      false,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	if err := r.db.WithContext(ctx).Create(&resetToken).Error; err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	return nil
}

// UpdateUserPassword updates a user's password by email
func (r *dashboardRepository) UpdateUserPassword(ctx context.Context, email, hashedPassword string) error {
	if err := r.db.WithContext(ctx).Model(&timescale.CompanyProfile{}).
		Where("email = ?", email).
		Update("password", hashedPassword).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// ChangeUserPassword updates a user's password by ID (used for authenticated password change)
func (r *dashboardRepository) ChangeUserPassword(ctx context.Context, userID, currentPasswordHash, newPasswordHash string) error {
	// First verify the current password
	var profile timescale.CompanyProfile
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// The current password verification happens in the service layer
	// Here we just update the password
	if err := r.db.WithContext(ctx).Model(&timescale.CompanyProfile{}).
		Where("id = ?", userID).
		Update("password", newPasswordHash).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// UpdateUserSubscription updates user's subscription plan, billing cycle, and expiry
func (r *dashboardRepository) UpdateUserSubscription(ctx context.Context, userID, planID, billingCycle string, expiresAt *time.Time) error {
	updates := map[string]interface{}{
		"subscription_plan":      planID,
		"billing_cycle":          billingCycle,
		"subscription_expires_at": expiresAt,
		"updated_at":             time.Now(),
	}

	if err := r.db.WithContext(ctx).Model(&timescale.CompanyProfile{}).
		Where("id = ?", userID).
		Updates(updates).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

// StoreNOWPayment stores a NOWPayments transaction
func (r *dashboardRepository) StoreNOWPayment(ctx context.Context, payment *models.Payment) error {
	dbPayment := timescale.DashboardPayment{
		ID:               payment.ID,
		ProfileID:        payment.ProfileID,
		Amount:           payment.Amount,
		Currency:         payment.Currency,
		SubscriptionType: payment.SubscriptionType,
		PaymentMethod:    payment.PaymentMethod,
		Status:           payment.Status,
		PaymentIntentID:  payment.ID, // Use payment ID as intent ID for NOWPayments
		CreatedAt:        payment.CreatedAt,
		UpdatedAt:        payment.UpdatedAt,
	}

	if err := r.db.WithContext(ctx).Create(&dbPayment).Error; err != nil {
		return fmt.Errorf("failed to store payment: %w", err)
	}

	return nil
}

// UpdatePaymentStatus updates payment status
func (r *dashboardRepository) UpdatePaymentStatus(ctx context.Context, paymentID, status string) error {
	if err := r.db.WithContext(ctx).Model(&timescale.DashboardPayment{}).
		Where("id = ?", paymentID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

// GetPaymentByID retrieves a payment by ID
func (r *dashboardRepository) GetPaymentByID(ctx context.Context, paymentID string) (*models.Payment, error) {
	var payment timescale.DashboardPayment
	if err := r.db.WithContext(ctx).Where("id = ?", paymentID).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return &models.Payment{
		ID:               payment.ID,
		ProfileID:        payment.ProfileID,
		Amount:           payment.Amount,
		Currency:         payment.Currency,
		SubscriptionType: payment.SubscriptionType,
		PaymentMethod:    payment.PaymentMethod,
		Status:           payment.Status,
		CreatedAt:        payment.CreatedAt,
		UpdatedAt:        payment.UpdatedAt,
	}, nil
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

	now := time.Now()
	dbAPIKey := timescale.DashboardAPIKey{
		ID:        uuid.New().String(),
		ProfileID: profileID,
		Name:      req.Name,
		KeyPrefix: keyPrefix,
		KeyHash:   apiKey, // Store plain text as hash for compatibility
		KeyPlain:  apiKey, // Store plain text API key
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
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
		keys[i] = models.APIKey{
			ID:        dbKey.ID,
			Key:       dbKey.KeyPlain, // Return plain text API key
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
		Key:       dbKey.KeyPlain, // Return plain text API key
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

	// Compare the provided key with each stored plain text key from the matching prefix
	for _, dbKey := range dbKeys {
		if dbKey.KeyPlain == apiKey {
			// Found matching key
			return &models.APIKey{
				ID:        dbKey.ID,
				ProfileID: dbKey.ProfileID,
				Name:      dbKey.Name,
				IsActive:  dbKey.IsActive,
				CreatedAt: dbKey.CreatedAt,
				UpdatedAt: dbKey.UpdatedAt,
				LastUsed:  dbKey.LastUsed,
				Key:       dbKey.KeyPlain,
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
		Key:       dbKey.KeyPlain, // Return plain text API key
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

// Invoice management methods

// CreateInvoice creates a new invoice
func (r *dashboardRepository) CreateInvoice(ctx context.Context, invoice *models.Invoice) error {
	dbInvoice := timescale.Invoice{
		ID:            uuid.New(),
		InvoiceNumber: invoice.InvoiceNumber,
		AccountID:     invoice.AccountID,
		Amount:        invoice.Amount,
		Currency:      invoice.Currency,
		DueDate:       invoice.DueDate,
		IssuedAt:      invoice.IssuedAt,
		Status:        invoice.Status,
		Metadata:      convertToDatatypesJSON(invoice.Metadata),
		PaidAt:        invoice.PaidAt,
		PaymentID:     invoice.PaymentID,
		CreatedAt:     invoice.CreatedAt,
		UpdatedAt:     invoice.UpdatedAt,
	}

	if err := r.db.WithContext(ctx).Create(&dbInvoice).Error; err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	// Update the invoice ID in the model
	invoice.ID = dbInvoice.ID.String()
	return nil
}

// GetInvoiceByID retrieves an invoice by ID
func (r *dashboardRepository) GetInvoiceByID(ctx context.Context, invoiceID string) (*models.Invoice, error) {
	var dbInvoice timescale.Invoice
	if err := r.db.WithContext(ctx).Where("id = ?", invoiceID).First(&dbInvoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return r.invoiceGormToModel(dbInvoice), nil
}

// GetInvoiceByNumber retrieves an invoice by invoice number
func (r *dashboardRepository) GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*models.Invoice, error) {
	var dbInvoice timescale.Invoice
	if err := r.db.WithContext(ctx).Where("invoice_number = ?", invoiceNumber).First(&dbInvoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return r.invoiceGormToModel(dbInvoice), nil
}

// GetInvoicesByAccount retrieves invoices for a specific account
func (r *dashboardRepository) GetInvoicesByAccount(ctx context.Context, accountID string, limit int, offset int) ([]models.Invoice, int64, error) {
	var dbInvoices []timescale.Invoice
	var totalCount int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&timescale.Invoice{}).Where("account_id = ?", accountID).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Get invoices with pagination
	query := r.db.WithContext(ctx).Where("account_id = ?", accountID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&dbInvoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get invoices: %w", err)
	}

	invoices := make([]models.Invoice, len(dbInvoices))
	for i, dbInvoice := range dbInvoices {
		invoices[i] = *r.invoiceGormToModel(dbInvoice)
	}

	return invoices, totalCount, nil
}

// GetInvoicesByStatus retrieves invoices by status
func (r *dashboardRepository) GetInvoicesByStatus(ctx context.Context, status string, limit int, offset int) ([]models.Invoice, int64, error) {
	var dbInvoices []timescale.Invoice
	var totalCount int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&timescale.Invoice{}).Where("status = ?", status).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Get invoices with pagination
	query := r.db.WithContext(ctx).Where("status = ?", status).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&dbInvoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get invoices: %w", err)
	}

	invoices := make([]models.Invoice, len(dbInvoices))
	for i, dbInvoice := range dbInvoices {
		invoices[i] = *r.invoiceGormToModel(dbInvoice)
	}

	return invoices, totalCount, nil
}

// GetInvoicesDueSoon retrieves invoices due within the specified number of days
func (r *dashboardRepository) GetInvoicesDueSoon(ctx context.Context, daysAhead int) ([]models.Invoice, error) {
	var dbInvoices []timescale.Invoice
	
	// Calculate the date range
	now := time.Now()
	endDate := now.AddDate(0, 0, daysAhead)
	
	// Get invoices due within the date range
	if err := r.db.WithContext(ctx).
		Where("due_date BETWEEN ? AND ? AND status = ?", now, endDate, "pending").
		Order("due_date ASC").
		Find(&dbInvoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoices due soon: %w", err)
	}

	invoices := make([]models.Invoice, len(dbInvoices))
	for i, dbInvoice := range dbInvoices {
		invoices[i] = *r.invoiceGormToModel(dbInvoice)
	}

	return invoices, nil
}

// UpdateInvoiceStatus updates an invoice's status and payment information
func (r *dashboardRepository) UpdateInvoiceStatus(ctx context.Context, invoiceID string, status string, paymentID *string, paidAt *time.Time) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if paymentID != nil {
		updates["payment_id"] = *paymentID
	}
	if paidAt != nil {
		updates["paid_at"] = *paidAt
	}

	if err := r.db.WithContext(ctx).Model(&timescale.Invoice{}).
		Where("id = ?", invoiceID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	return nil
}

// CheckInvoiceExists checks if an invoice already exists for an account and due date
func (r *dashboardRepository) CheckInvoiceExists(ctx context.Context, accountID string, dueDate time.Time) (bool, error) {
	var count int64
	
	// Normalize the due date to start of day for comparison
	startOfDay := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	if err := r.db.WithContext(ctx).Model(&timescale.Invoice{}).
		Where("account_id = ? AND due_date BETWEEN ? AND ? AND status != ?", accountID, startOfDay, endOfDay, "void").
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check invoice existence: %w", err)
	}

	return count > 0, nil
}

// GetInvoicesForPayment retrieves invoices for a specific account and due date (for payment processing)
func (r *dashboardRepository) GetInvoicesForPayment(ctx context.Context, accountID string, dueDate time.Time) ([]models.Invoice, error) {
	var dbInvoices []timescale.Invoice
	
	// Normalize the due date to start of day for comparison
	startOfDay := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	if err := r.db.WithContext(ctx).
		Where("account_id = ? AND due_date BETWEEN ? AND ? AND status = ?", accountID, startOfDay, endOfDay, "pending").
		Find(&dbInvoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoices for payment: %w", err)
	}

	invoices := make([]models.Invoice, len(dbInvoices))
	for i, dbInvoice := range dbInvoices {
		invoices[i] = *r.invoiceGormToModel(dbInvoice)
	}

	return invoices, nil
}

// Helper function to convert GORM invoice model to domain model
func (r *dashboardRepository) invoiceGormToModel(dbInvoice timescale.Invoice) *models.Invoice {
	return &models.Invoice{
		ID:            dbInvoice.ID.String(),
		InvoiceNumber: dbInvoice.InvoiceNumber,
		AccountID:     dbInvoice.AccountID,
		Amount:        dbInvoice.Amount,
		Currency:      dbInvoice.Currency,
		DueDate:       dbInvoice.DueDate,
		IssuedAt:      dbInvoice.IssuedAt,
		Status:        dbInvoice.Status,
		Metadata:      convertFromDatatypesJSON(dbInvoice.Metadata),
		PaidAt:        dbInvoice.PaidAt,
		PaymentID:     dbInvoice.PaymentID,
		CreatedAt:     dbInvoice.CreatedAt,
		UpdatedAt:     dbInvoice.UpdatedAt,
	}
}

// Helper function to convert GORM model to domain model
func (r *dashboardRepository) gormToModel(profile timescale.CompanyProfile) *models.CompanyProfile {
	return &models.CompanyProfile{
		ID:                    profile.ID,
		Name:                  profile.Name,
		Description:           profile.Description,
		Website:               profile.Website,
		LogoURL:               profile.LogoURL,
		FirstName:             profile.FirstName,
		LastName:              profile.LastName,
		Email:                 profile.Email,
		Password:              profile.Password,
		SubscriptionPlan:      profile.SubscriptionPlan,
		BillingCycle:          profile.BillingCycle,
		SubscriptionExpiresAt: profile.SubscriptionExpiresAt,
		CreatedAt:             profile.CreatedAt,
		UpdatedAt:             profile.UpdatedAt,
	}
}

// Helper function to convert map[string]interface{} to datatypes.JSON
func convertToDatatypesJSON(metadata map[string]interface{}) datatypes.JSON {
	if metadata == nil {
		return datatypes.JSON("{}")
	}
	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		logging.Logger.Error("Failed to marshal metadata to JSON", zap.Error(err))
		return datatypes.JSON("{}")
	}
	return datatypes.JSON(jsonBytes)
}

// Helper function to convert datatypes.JSON to map[string]interface{}
func convertFromDatatypesJSON(metadata datatypes.JSON) map[string]interface{} {
	if len(metadata) == 0 {
		return make(map[string]interface{})
	}
	var result map[string]interface{}
	err := json.Unmarshal(metadata, &result)
	if err != nil {
		logging.Logger.Error("Failed to unmarshal metadata from JSON", zap.Error(err))
		return make(map[string]interface{})
	}
	return result
}
