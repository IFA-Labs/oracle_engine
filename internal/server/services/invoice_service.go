package services

import (
	"context"
	"fmt"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"oracle_engine/internal/server/repository"
	"oracle_engine/internal/utils"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type InvoiceService interface {
	// Invoice CRUD operations
	CreateInvoice(ctx context.Context, req *models.CreateInvoiceRequest) (*models.Invoice, error)
	GetInvoiceByID(ctx context.Context, invoiceID string) (*models.Invoice, error)
	GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*models.Invoice, error)
	GetInvoicesByAccount(ctx context.Context, accountID string, limit int, offset int) (*models.InvoiceListResponse, error)
	GetInvoicesByStatus(ctx context.Context, status string, limit int, offset int) (*models.InvoiceListResponse, error)
	UpdateInvoiceStatus(ctx context.Context, invoiceID string, req *models.UpdateInvoiceStatusRequest) error

	// Invoice generation and automation
	GenerateInvoicesForDueAccounts(ctx context.Context) (*models.InvoiceGenerationJob, error)
	SendInvoiceNotification(ctx context.Context, invoice *models.Invoice) error
	ProcessPaymentForInvoice(ctx context.Context, accountID string, paymentID string, amount float64, currency string) error

	// Utility methods
	GenerateInvoiceNumber() string
	CalculateNextPaymentDate(subscriptionExpiresAt *time.Time, billingCycle string) *time.Time
	GetSubscriptionAmount(subscriptionPlan string, billingCycle string) int64
}

type invoiceService struct {
	repo        repository.DashboardRepository
	emailService *utils.EmailService
}

func NewInvoiceService(repo repository.DashboardRepository, emailService *utils.EmailService) InvoiceService {
	return &invoiceService{
		repo:         repo,
		emailService: emailService,
	}
}

// CreateInvoice creates a new invoice
func (s *invoiceService) CreateInvoice(ctx context.Context, req *models.CreateInvoiceRequest) (*models.Invoice, error) {
	// Generate unique invoice number
	invoiceNumber := s.GenerateInvoiceNumber()

	// Check if invoice already exists for this account and due date
	exists, err := s.repo.CheckInvoiceExists(ctx, req.AccountID, req.DueDate)
	if err != nil {
		return nil, fmt.Errorf("failed to check invoice existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("invoice already exists for account %s on due date %s", req.AccountID, req.DueDate.Format("2006-01-02"))
	}

	now := time.Now()
	invoice := &models.Invoice{
		ID:            uuid.New().String(),
		InvoiceNumber: invoiceNumber,
		AccountID:     req.AccountID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		DueDate:       req.DueDate,
		IssuedAt:      now,
		Status:        "pending",
		Metadata:      req.Metadata,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	logging.Logger.Info("Invoice created successfully",
		zap.String("invoice_id", invoice.ID),
		zap.String("invoice_number", invoice.InvoiceNumber),
		zap.String("account_id", invoice.AccountID),
		zap.Int64("amount", invoice.Amount),
		zap.String("due_date", invoice.DueDate.Format("2006-01-02")),
	)

	return invoice, nil
}

// GetInvoiceByID retrieves an invoice by ID
func (s *invoiceService) GetInvoiceByID(ctx context.Context, invoiceID string) (*models.Invoice, error) {
	return s.repo.GetInvoiceByID(ctx, invoiceID)
}

// GetInvoiceByNumber retrieves an invoice by invoice number
func (s *invoiceService) GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*models.Invoice, error) {
	return s.repo.GetInvoiceByNumber(ctx, invoiceNumber)
}

// GetInvoicesByAccount retrieves invoices for a specific account
func (s *invoiceService) GetInvoicesByAccount(ctx context.Context, accountID string, limit int, offset int) (*models.InvoiceListResponse, error) {
	invoices, totalCount, err := s.repo.GetInvoicesByAccount(ctx, accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &models.InvoiceListResponse{
		Invoices:   invoices,
		TotalCount: totalCount,
		Page:       (offset / limit) + 1,
		PageSize:   limit,
	}, nil
}

// GetInvoicesByStatus retrieves invoices by status
func (s *invoiceService) GetInvoicesByStatus(ctx context.Context, status string, limit int, offset int) (*models.InvoiceListResponse, error) {
	invoices, totalCount, err := s.repo.GetInvoicesByStatus(ctx, status, limit, offset)
	if err != nil {
		return nil, err
	}

	return &models.InvoiceListResponse{
		Invoices:   invoices,
		TotalCount: totalCount,
		Page:       (offset / limit) + 1,
		PageSize:   limit,
	}, nil
}

// UpdateInvoiceStatus updates an invoice's status
func (s *invoiceService) UpdateInvoiceStatus(ctx context.Context, invoiceID string, req *models.UpdateInvoiceStatusRequest) error {
	return s.repo.UpdateInvoiceStatus(ctx, invoiceID, req.Status, req.PaymentID, req.PaidAt)
}

// GenerateInvoicesForDueAccounts generates invoices for accounts with payments due in 7 days
func (s *invoiceService) GenerateInvoicesForDueAccounts(ctx context.Context) (*models.InvoiceGenerationJob, error) {
	job := &models.InvoiceGenerationJob{
		ID:        uuid.New().String(),
		Status:    "running",
		StartedAt: time.Now(),
	}

	logging.Logger.Info("Starting invoice generation job", zap.String("job_id", job.ID))

	// Get all accounts with subscriptions that expire in 7 days
	sevenDaysFromNow := time.Now().AddDate(0, 0, 7)
	
	// We need to query accounts where subscription_expires_at is around 7 days from now
	// Since we don't have a direct method for this, we'll get all accounts and filter
	// In a real implementation, you might want to add a specific repository method for this
	
	// For now, let's implement a simplified version that gets accounts due soon
	// This would need to be enhanced based on your specific requirements
	
	var invoicesCreated int
	var emailsSent int
	var errors []string

	// Get accounts that need invoices (this is a simplified approach)
	// In practice, you'd want to query accounts where subscription_expires_at is around 7 days from now
	accountsToProcess := s.getAccountsDueForInvoices(ctx, sevenDaysFromNow)

	for _, account := range accountsToProcess {
		// Calculate next payment date
		nextPaymentDate := s.CalculateNextPaymentDate(account.SubscriptionExpiresAt, account.BillingCycle)
		if nextPaymentDate == nil {
			errors = append(errors, fmt.Sprintf("Could not calculate next payment date for account %s", account.ID))
			continue
		}

		// Check if invoice already exists
		exists, err := s.repo.CheckInvoiceExists(ctx, account.ID, *nextPaymentDate)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to check invoice existence for account %s: %v", account.ID, err))
			continue
		}
		if exists {
			logging.Logger.Info("Invoice already exists for account", 
				zap.String("account_id", account.ID),
				zap.String("due_date", nextPaymentDate.Format("2006-01-02")),
			)
			continue
		}

		// Calculate amount based on subscription plan
		amount := s.GetSubscriptionAmount(account.SubscriptionPlan, account.BillingCycle)
		if amount == 0 {
			logging.Logger.Warn("No amount calculated for account", 
				zap.String("account_id", account.ID),
				zap.String("subscription_plan", account.SubscriptionPlan),
				zap.String("billing_cycle", account.BillingCycle),
			)
			continue
		}

		// Create invoice
		invoiceReq := &models.CreateInvoiceRequest{
			AccountID: account.ID,
			Amount:    amount,
			Currency:  "USD",
			DueDate:   *nextPaymentDate,
			Metadata: map[string]interface{}{
				"subscription_plan": account.SubscriptionPlan,
				"billing_cycle":     account.BillingCycle,
				"account_name":      account.Name,
				"account_email":    account.Email,
			},
		}

		invoice, err := s.CreateInvoice(ctx, invoiceReq)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to create invoice for account %s: %v", account.ID, err))
			continue
		}

		invoicesCreated++

		// Send email notification
		if err := s.SendInvoiceNotification(ctx, invoice); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to send email for invoice %s: %v", invoice.InvoiceNumber, err))
			continue
		}

		emailsSent++
	}

	job.Status = "completed"
	job.EndedAt = &[]time.Time{time.Now()}[0]
	job.InvoicesCreated = invoicesCreated
	job.EmailsSent = emailsSent
	job.Errors = errors

	logging.Logger.Info("Invoice generation job completed",
		zap.String("job_id", job.ID),
		zap.Int("invoices_created", invoicesCreated),
		zap.Int("emails_sent", emailsSent),
		zap.Int("errors", len(errors)),
	)

	return job, nil
}

// SendInvoiceNotification sends an email notification for a new invoice
func (s *invoiceService) SendInvoiceNotification(ctx context.Context, invoice *models.Invoice) error {
	// Get account details
	account, err := s.repo.GetUserByID(ctx, invoice.AccountID)
	if err != nil {
		return fmt.Errorf("failed to get account details: %w", err)
	}

	// Prepare email data
	amountInDollars := float64(invoice.Amount) / 100.0
	
	emailData := map[string]interface{}{
		"AccountName":    account.Name,
		"InvoiceNumber":  invoice.InvoiceNumber,
		"Amount":         fmt.Sprintf("$%.2f", amountInDollars),
		"Currency":       invoice.Currency,
		"DueDate":        invoice.DueDate.Format("January 2, 2006"),
		"IssuedDate":     invoice.IssuedAt.Format("January 2, 2006"),
		"SubscriptionPlan": invoice.Metadata["subscription_plan"],
		"BillingCycle":   invoice.Metadata["billing_cycle"],
	}

	// Send email
	if err := s.emailService.SendInvoiceNotification(account.Email, emailData); err != nil {
		return fmt.Errorf("failed to send invoice email: %w", err)
	}

	logging.Logger.Info("Invoice notification email sent",
		zap.String("invoice_number", invoice.InvoiceNumber),
		zap.String("account_email", account.Email),
	)

	return nil
}

// ProcessPaymentForInvoice processes a payment and marks the corresponding invoice as paid
func (s *invoiceService) ProcessPaymentForInvoice(ctx context.Context, accountID string, paymentID string, amount float64, currency string) error {
	// Find pending invoices for this account
	// We'll look for invoices due around now (within a reasonable time window)
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	invoices, err := s.repo.GetInvoicesForPayment(ctx, accountID, startOfDay)
	if err != nil {
		return fmt.Errorf("failed to get invoices for payment: %w", err)
	}

	if len(invoices) == 0 {
		logging.Logger.Warn("No pending invoices found for payment",
			zap.String("account_id", accountID),
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

// GenerateInvoiceNumber generates a unique invoice number
func (s *invoiceService) GenerateInvoiceNumber() string {
	year := time.Now().Year()
	month := int(time.Now().Month())
	
	// Format: INV-YYYY-MM-XXXXXX (where XXXXXX is a random number)
	randomPart := uuid.New().String()[:6]
	return fmt.Sprintf("INV-%d-%02d-%s", year, month, randomPart)
}

// CalculateNextPaymentDate calculates the next payment date based on subscription expiry and billing cycle
func (s *invoiceService) CalculateNextPaymentDate(subscriptionExpiresAt *time.Time, billingCycle string) *time.Time {
	if subscriptionExpiresAt == nil {
		return nil
	}

	switch billingCycle {
	case "monthly":
		return subscriptionExpiresAt
	case "annual":
		return subscriptionExpiresAt
	case "lifetime":
		return nil // No recurring payments for lifetime
	default:
		return subscriptionExpiresAt
	}
}

// GetSubscriptionAmount returns the amount in cents for a subscription plan and billing cycle
func (s *invoiceService) GetSubscriptionAmount(subscriptionPlan string, billingCycle string) int64 {
	// Define subscription pricing (in cents)
	pricing := map[string]map[string]int64{
		"developer": {
			"monthly": 5000,  // $50.00
			"annual":  50000, // $500.00 (assuming some discount)
		},
		"professional": {
			"monthly": 10000, // $100.00
			"annual":  100000, // $1000.00
		},
		"enterprise": {
			"monthly": 20000, // $200.00
			"annual":  200000, // $2000.00
		},
	}

	if planPricing, exists := pricing[subscriptionPlan]; exists {
		if amount, exists := planPricing[billingCycle]; exists {
			return amount
		}
	}

	return 0
}

// getAccountsDueForInvoices is a helper method to get accounts that need invoices
// This is a simplified implementation - in practice, you'd want a more sophisticated query
func (s *invoiceService) getAccountsDueForInvoices(ctx context.Context, targetDate time.Time) []*models.CompanyProfile {
	// This is a placeholder implementation
	// In a real system, you'd want to query accounts where:
	// 1. subscription_expires_at is around the target date
	// 2. subscription_plan is not 'free'
	// 3. billing_cycle is not 'lifetime'
	
	// For now, return empty slice - this would need to be implemented based on your specific needs
	return []*models.CompanyProfile{}
}
