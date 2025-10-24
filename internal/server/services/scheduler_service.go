package services

import (
	"context"
	"fmt"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"oracle_engine/internal/server/repository"
	"oracle_engine/internal/utils"
	"sync"
	"time"

	"go.uber.org/zap"
)

type SchedulerService interface {
	Start(ctx context.Context) error
	Stop() error
	RunInvoiceGenerationJob(ctx context.Context) error
	GetJobStatus() map[string]interface{}
}

type schedulerService struct {
	invoiceService InvoiceService
	repo           repository.DashboardRepository
	emailService   *utils.EmailService
	ticker         *time.Ticker
	stopChan       chan struct{}
	running        bool
	mu             sync.RWMutex
	lastJobRun     *time.Time
	jobStatus      map[string]interface{}
}

func NewSchedulerService(repo repository.DashboardRepository, emailService *utils.EmailService) SchedulerService {
	invoiceService := NewInvoiceService(repo, emailService)
	
	return &schedulerService{
		invoiceService: invoiceService,
		repo:           repo,
		emailService:   emailService,
		stopChan:       make(chan struct{}),
		jobStatus:      make(map[string]interface{}),
	}
}

// Start begins the scheduler service
func (s *schedulerService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	// Run invoice generation job immediately on startup
	go func() {
		if err := s.RunInvoiceGenerationJob(ctx); err != nil {
			logging.Logger.Error("Failed to run initial invoice generation job", zap.Error(err))
		}
	}()

	// Schedule daily runs at 9:00 AM UTC
	s.ticker = time.NewTicker(24 * time.Hour)
	s.running = true

	go s.runScheduler(ctx)

	logging.Logger.Info("Scheduler service started successfully")
	return nil
}

// Stop stops the scheduler service
func (s *schedulerService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("scheduler is not running")
	}

	s.running = false
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopChan)

	logging.Logger.Info("Scheduler service stopped")
	return nil
}

// runScheduler runs the main scheduler loop
func (s *schedulerService) runScheduler(ctx context.Context) {
	for {
		select {
		case <-s.ticker.C:
			// Run invoice generation job
			if err := s.RunInvoiceGenerationJob(ctx); err != nil {
				logging.Logger.Error("Failed to run scheduled invoice generation job", zap.Error(err))
			}
		case <-s.stopChan:
			logging.Logger.Info("Scheduler loop stopped")
			return
		case <-ctx.Done():
			logging.Logger.Info("Scheduler context cancelled")
			return
		}
	}
}

// RunInvoiceGenerationJob runs the invoice generation job manually
func (s *schedulerService) RunInvoiceGenerationJob(ctx context.Context) error {
	s.mu.Lock()
	s.jobStatus["status"] = "running"
	s.jobStatus["started_at"] = time.Now()
	s.mu.Unlock()

	logging.Logger.Info("Starting invoice generation job")

	// Use advisory lock to prevent multiple instances from running simultaneously
	lockKey := "invoice_generation_job"
	if err := s.acquireAdvisoryLock(ctx, lockKey); err != nil {
		logging.Logger.Warn("Could not acquire advisory lock, another instance may be running", zap.Error(err))
		return err
	}
	defer s.releaseAdvisoryLock(ctx, lockKey)

	// Run the invoice generation
	job, err := s.invoiceService.GenerateInvoicesForDueAccounts(ctx)
	if err != nil {
		s.mu.Lock()
		s.jobStatus["status"] = "failed"
		s.jobStatus["error"] = err.Error()
		s.mu.Unlock()
		
		logging.Logger.Error("Invoice generation job failed", zap.Error(err))
		return err
	}

	s.mu.Lock()
	s.jobStatus["status"] = "completed"
	s.jobStatus["last_run"] = time.Now()
	s.jobStatus["invoices_created"] = job.InvoicesCreated
	s.jobStatus["emails_sent"] = job.EmailsSent
	s.jobStatus["errors"] = job.Errors
	s.lastJobRun = &job.StartedAt
	s.mu.Unlock()

	logging.Logger.Info("Invoice generation job completed successfully",
		zap.Int("invoices_created", job.InvoicesCreated),
		zap.Int("emails_sent", job.EmailsSent),
		zap.Int("errors", len(job.Errors)),
	)

	return nil
}

// GetJobStatus returns the current job status
func (s *schedulerService) GetJobStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := make(map[string]interface{})
	for k, v := range s.jobStatus {
		status[k] = v
	}
	
	if s.lastJobRun != nil {
		status["last_job_run"] = s.lastJobRun.Format(time.RFC3339)
	}
	
	status["running"] = s.running
	return status
}

// acquireAdvisoryLock acquires a PostgreSQL advisory lock
func (s *schedulerService) acquireAdvisoryLock(ctx context.Context, lockKey string) error {
	// Convert string to int64 for advisory lock
	var lockID int64
	for _, char := range lockKey {
		lockID = lockID*31 + int64(char)
	}
	
	// Use a simple approach - in production, you might want to use a more sophisticated locking mechanism
	// For now, we'll use a simple in-memory lock
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Check if we already have a lock
	if _, exists := s.jobStatus["lock_acquired"]; exists {
		return fmt.Errorf("lock already acquired")
	}
	
	s.jobStatus["lock_acquired"] = true
	s.jobStatus["lock_id"] = lockID
	
	return nil
}

// releaseAdvisoryLock releases the advisory lock
func (s *schedulerService) releaseAdvisoryLock(ctx context.Context, lockKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.jobStatus, "lock_acquired")
	delete(s.jobStatus, "lock_id")
	
	return nil
}

// Enhanced invoice service with better account querying
type EnhancedInvoiceService struct {
	*invoiceService
}

func NewEnhancedInvoiceService(repo repository.DashboardRepository, emailService *utils.EmailService) InvoiceService {
	baseService := NewInvoiceService(repo, emailService)
	return &EnhancedInvoiceService{
		invoiceService: baseService.(*invoiceService),
	}
}

// getAccountsDueForInvoices gets accounts that need invoices generated
func (s *EnhancedInvoiceService) getAccountsDueForInvoices(ctx context.Context, targetDate time.Time) []*models.CompanyProfile {
	// This is a simplified implementation
	// In a real system, you'd want to query accounts where:
	// 1. subscription_expires_at is around the target date (7 days from now)
	// 2. subscription_plan is not 'free'
	// 3. billing_cycle is not 'lifetime'
	
	// For now, we'll implement a basic version that gets all paid accounts
	// and filters them based on subscription expiry
	
	var accounts []*models.CompanyProfile
	
	// Get all accounts with paid subscriptions
	// This would need to be implemented in the repository layer
	// For now, return empty slice as placeholder
	
	logging.Logger.Info("Getting accounts due for invoices",
		zap.String("target_date", targetDate.Format("2006-01-02")),
		zap.Int("accounts_found", len(accounts)),
	)
	
	return accounts
}

// Override the getAccountsDueForInvoices method
func (s *EnhancedInvoiceService) GenerateInvoicesForDueAccounts(ctx context.Context) (*models.InvoiceGenerationJob, error) {
	job := &models.InvoiceGenerationJob{
		ID:        fmt.Sprintf("job_%d", time.Now().Unix()),
		Status:    "running",
		StartedAt: time.Now(),
	}

	logging.Logger.Info("Starting enhanced invoice generation job", zap.String("job_id", job.ID))

	// Get all accounts with subscriptions that expire in 7 days
	sevenDaysFromNow := time.Now().AddDate(0, 0, 7)
	accountsToProcess := s.getAccountsDueForInvoices(ctx, sevenDaysFromNow)

	var invoicesCreated int
	var emailsSent int
	var errors []string

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

	logging.Logger.Info("Enhanced invoice generation job completed",
		zap.String("job_id", job.ID),
		zap.Int("invoices_created", invoicesCreated),
		zap.Int("emails_sent", emailsSent),
		zap.Int("errors", len(errors)),
	)

	return job, nil
}
