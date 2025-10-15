package utils

import (
	"fmt"
	"net/smtp"
	"oracle_engine/internal/logging"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

type EmailService struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

func NewEmailService() *EmailService {
	return &EmailService{
		SMTPHost:     getEnvOrDefault("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvOrDefault("SMTP_PORT", "587"),
		SMTPUser:     getEnvOrDefault("SMTP_USER", ""),
		SMTPPassword: getEnvOrDefault("SMTP_PASSWORD", ""),
		FromEmail:    getEnvOrDefault("SMTP_FROM_EMAIL", "noreply@ifalabs.com"),
		FromName:     getEnvOrDefault("SMTP_FROM_NAME", "IFA Labs"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (e *EmailService) SendVerificationEmail(toEmail, verificationToken string) error {
	// Get the frontend URL from environment or use default
	frontendURL := getEnvOrDefault("FRONTEND_URL", "http://localhost:3000")
	verificationLink := fmt.Sprintf("%s/complete-registration?token=%s", frontendURL, verificationToken)

	subject := "Verify your email address - IFA Labs"
	body := e.buildVerificationEmailHTML(toEmail, verificationLink)

	return e.sendEmail(toEmail, subject, body)
}

func (e *EmailService) SendWelcomeEmail(toEmail, name string) error {
	subject := "Welcome to IFA Labs!"
	body := e.buildWelcomeEmailHTML(name)

	return e.sendEmail(toEmail, subject, body)
}

func (e *EmailService) SendPasswordResetEmail(toEmail, resetToken string) error {
	// Get the frontend URL from environment or use default
	frontendURL := getEnvOrDefault("FRONTEND_URL", "http://localhost:3000")
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, resetToken)

	subject := "Reset your password - IFA Labs"
	body := e.buildPasswordResetEmailHTML(toEmail, resetLink)

	return e.sendEmail(toEmail, subject, body)
}

func (e *EmailService) SendPasswordChangedEmail(toEmail, name string) error {
	subject := "Password Changed Successfully - IFA Labs"
	body := e.buildPasswordChangedEmailHTML(toEmail, name)

	return e.sendEmail(toEmail, subject, body)
}

func (e *EmailService) SendAPIKeyCreatedEmail(toEmail, name, keyName, apiKeyPreview string) error {
	subject := "New API Key Created - IFA Labs"
	body := e.buildAPIKeyCreatedEmailHTML(toEmail, name, keyName, apiKeyPreview)

	return e.sendEmail(toEmail, subject, body)
}

func (e *EmailService) SendSubscriptionActivatedEmail(toEmail, name, planID, billingCycle string, expiresAt *time.Time) error {
	subject := "Subscription Activated - IFA Labs"
	body := e.buildSubscriptionActivatedEmailHTML(toEmail, name, planID, billingCycle, expiresAt)

	return e.sendEmail(toEmail, subject, body)
}

func (e *EmailService) sendEmail(to, subject, body string) error {
	// If SMTP credentials are not configured, log and return (development mode)
	if e.SMTPUser == "" || e.SMTPPassword == "" {
		logging.Logger.Warn("SMTP not configured, skipping email send",
			zap.String("to", to),
			zap.String("subject", subject))
		logging.Logger.Info("Email would have been sent:",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.String("preview", body[:100]+"..."))
		return nil
	}

	// Set up authentication
	auth := smtp.PlainAuth("", e.SMTPUser, e.SMTPPassword, e.SMTPHost)

	// Compose message
	from := fmt.Sprintf("%s <%s>", e.FromName, e.FromEmail)
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", from, to, subject, body))

	// Send email
	addr := fmt.Sprintf("%s:%s", e.SMTPHost, e.SMTPPort)
	err := smtp.SendMail(addr, auth, e.FromEmail, []string{to}, msg)
	if err != nil {
		logging.Logger.Error("Failed to send email",
			zap.Error(err),
			zap.String("to", to),
			zap.String("subject", subject))
		return err
	}

	logging.Logger.Info("Email sent successfully",
		zap.String("to", to),
		zap.String("subject", subject))
	return nil
}

func (e *EmailService) buildVerificationEmailHTML(email, verificationLink string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Verify Your Email - IFA Labs</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="text-align: center; margin-bottom: 30px;">
    <h1 style="color: #4F46E5; margin: 0;">IFA Labs</h1>
  </div>
  
  <h2 style="color: #1a1a1a; margin-bottom: 20px;">Verify Your Email Address</h2>
  
  <p style="margin-bottom: 20px;">Thank you for signing up for IFA Labs! Please verify your email address to complete your registration.</p>
  
  <div style="text-align: center; margin: 30px 0;">
    <a href="%s" 
       style="background-color: #4F46E5; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: bold;">
      Verify Email Address
    </a>
  </div>
  
  <p style="margin-bottom: 10px;">Or copy and paste this link into your browser:</p>
  <p style="background-color: #f5f5f5; padding: 10px; word-break: break-all; border-radius: 4px; font-size: 14px;">
    %s
  </p>
  
  <p style="color: #666; font-size: 14px; margin-top: 30px;">
    This link will expire in 24 hours. If you didn't request this email, you can safely ignore it.
  </p>
  
  <hr style="border: none; border-top: 1px solid #e5e5e5; margin: 30px 0;">
  
  <p style="color: #999; font-size: 12px; text-align: center;">
    © 2024 IFA Labs. All rights reserved.
  </p>
</body>
</html>
`, verificationLink, verificationLink)
}

func (e *EmailService) buildWelcomeEmailHTML(name string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Welcome to IFA Labs</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="text-align: center; margin-bottom: 30px;">
    <h1 style="color: #4F46E5; margin: 0;">IFA Labs</h1>
  </div>
  
  <h2 style="color: #1a1a1a; margin-bottom: 20px;">Welcome %s!</h2>
  
  <p style="margin-bottom: 20px;">Your account has been successfully created. You can now access all the features of IFA Labs Oracle Engine.</p>
  
  <div style="background-color: #f9fafb; border-left: 4px solid #4F46E5; padding: 16px; margin: 20px 0;">
    <h3 style="margin-top: 0; color: #4F46E5;">Getting Started</h3>
    <ul style="margin: 10px 0; padding-left: 20px;">
      <li>Create your first API key from your dashboard</li>
      <li>Explore our comprehensive API documentation</li>
      <li>Start accessing real-time oracle price feeds</li>
    </ul>
  </div>
  
  <p style="margin-bottom: 20px;">If you have any questions, feel free to reach out to our support team.</p>
  
  <div style="text-align: center; margin: 30px 0;">
    <a href="%s" 
       style="background-color: #4F46E5; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: bold;">
      Go to Dashboard
    </a>
  </div>
  
  <hr style="border: none; border-top: 1px solid #e5e5e5; margin: 30px 0;">
  
  <p style="color: #999; font-size: 12px; text-align: center;">
    © 2024 IFA Labs. All rights reserved.
  </p>
</body>
</html>
`, name, getEnvOrDefault("FRONTEND_URL", "http://localhost:3000")+"/dashboard")
}

func (e *EmailService) buildPasswordResetEmailHTML(email, resetLink string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Reset Your Password - IFA Labs</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="text-align: center; margin-bottom: 30px;">
    <h1 style="color: #4F46E5; margin: 0;">IFA Labs</h1>
  </div>
  
  <h2 style="color: #1a1a1a; margin-bottom: 20px;">Reset Your Password</h2>
  
  <p style="margin-bottom: 20px;">We received a request to reset your password for your IFA Labs account.</p>
  
  <div style="text-align: center; margin: 30px 0;">
    <a href="%s" 
       style="background-color: #4F46E5; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: bold;">
      Reset Password
    </a>
  </div>
  
  <p style="margin-bottom: 10px;">Or copy and paste this link into your browser:</p>
  <p style="background-color: #f5f5f5; padding: 10px; word-break: break-all; border-radius: 4px; font-size: 14px;">
    %s
  </p>
  
  <div style="background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 16px; margin: 20px 0;">
    <p style="margin: 0; color: #92400e;">
      <strong>Security Notice:</strong> If you didn't request this password reset, please ignore this email. Your password will remain unchanged.
    </p>
  </div>
  
  <p style="color: #666; font-size: 14px; margin-top: 30px;">
    This link will expire in 24 hours for security reasons.
  </p>
  
  <hr style="border: none; border-top: 1px solid #e5e5e5; margin: 30px 0;">
  
  <p style="color: #999; font-size: 12px; text-align: center;">
    © 2024 IFA Labs. All rights reserved.
  </p>
</body>
</html>
`, resetLink, resetLink)
}

func (e *EmailService) buildPasswordChangedEmailHTML(email, name string) string {
	frontendURL := getEnvOrDefault("FRONTEND_URL", "http://localhost:3000")
	currentTime := time.Now().Format("Monday, January 2, 2006 at 3:04 PM MST")
	
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Password Changed - IFA Labs</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="text-align: center; margin-bottom: 30px;">
    <h1 style="color: #4F46E5; margin: 0;">IFA Labs</h1>
  </div>
  
  <h2 style="color: #1a1a1a; margin-bottom: 20px;">Password Changed Successfully</h2>
  
  <p style="margin-bottom: 20px;">Hello %s,</p>
  
  <p style="margin-bottom: 20px;">This email confirms that your IFA Labs account password was successfully changed.</p>
  
  <div style="background-color: #f0fdf4; border-left: 4px solid #10b981; padding: 16px; margin: 20px 0;">
    <p style="margin: 0; color: #065f46;">
      <strong>✓ Password Changed</strong><br>
      Time: %s<br>
      Account: %s
    </p>
  </div>
  
  <p style="margin-bottom: 20px;">Your account is now secured with your new password. You can use it to sign in to your account.</p>
  
  <div style="text-align: center; margin: 30px 0;">
    <a href="%s/login" 
       style="background-color: #4F46E5; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: bold;">
      Sign In to Your Account
    </a>
  </div>
  
  <div style="background-color: #fee2e2; border-left: 4px solid #ef4444; padding: 16px; margin: 20px 0;">
    <p style="margin: 0; color: #991b1b;">
      <strong>⚠ Didn't change your password?</strong><br>
      If you did not make this change, your account may be compromised. Please contact our security team immediately and reset your password.
    </p>
  </div>
  
  <div style="background-color: #f9fafb; padding: 16px; margin: 20px 0; border-radius: 4px;">
    <h3 style="margin-top: 0; color: #4F46E5; font-size: 16px;">Security Tips</h3>
    <ul style="margin: 10px 0; padding-left: 20px; color: #666;">
      <li>Use a unique password for your IFA Labs account</li>
      <li>Never share your password with anyone</li>
      <li>Enable two-factor authentication when available</li>
      <li>Review your account activity regularly</li>
    </ul>
  </div>
  
  <p style="color: #666; font-size: 14px; margin-top: 30px;">
    If you have any questions or concerns, please contact our support team.
  </p>
  
  <hr style="border: none; border-top: 1px solid #e5e5e5; margin: 30px 0;">
  
  <p style="color: #999; font-size: 12px; text-align: center;">
    © 2024 IFA Labs. All rights reserved.
  </p>
</body>
</html>
`, name, currentTime, email, frontendURL)
}

func (e *EmailService) buildAPIKeyCreatedEmailHTML(email, name, keyName, apiKeyPreview string) string {
	frontendURL := getEnvOrDefault("FRONTEND_URL", "http://localhost:3000")
	currentTime := time.Now().Format("Monday, January 2, 2006 at 3:04 PM MST")
	
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>New API Key Created - IFA Labs</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="text-align: center; margin-bottom: 30px;">
    <h1 style="color: #4F46E5; margin: 0;">IFA Labs</h1>
  </div>
  
  <h2 style="color: #1a1a1a; margin-bottom: 20px;">New API Key Created</h2>
  
  <p style="margin-bottom: 20px;">Hello %s,</p>
  
  <p style="margin-bottom: 20px;">A new API key has been created for your IFA Labs account.</p>
  
  <div style="background-color: #eff6ff; border-left: 4px solid #3b82f6; padding: 16px; margin: 20px 0;">
    <p style="margin: 0; color: #1e40af;">
      <strong>🔑 API Key Details</strong><br>
      Key Name: <strong>%s</strong><br>
      Created: %s<br>
      Key Preview: <code style="background-color: #dbeafe; padding: 2px 6px; border-radius: 3px;">%s...</code>
    </p>
  </div>
  
  <div style="background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 16px; margin: 20px 0;">
    <p style="margin: 0; color: #92400e;">
      <strong>⚠️ Important Security Information</strong><br>
      • Store your API key securely - it won't be shown again<br>
      • Never share your API key with anyone<br>
      • Use environment variables to store the key<br>
      • Rotate keys regularly for better security
    </p>
  </div>
  
  <p style="margin-bottom: 20px;">You can manage your API keys from your dashboard at any time.</p>
  
  <div style="text-align: center; margin: 30px 0;">
    <a href="%s/api-keys" 
       style="background-color: #4F46E5; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: bold;">
      Manage API Keys
    </a>
  </div>
  
  <div style="background-color: #fee2e2; border-left: 4px solid #ef4444; padding: 16px; margin: 20px 0;">
    <p style="margin: 0; color: #991b1b;">
      <strong>⚠ Didn't create this API key?</strong><br>
      If you did not create this API key, your account may be compromised. Please delete this key immediately and change your password.
    </p>
  </div>
  
  <div style="background-color: #f9fafb; padding: 16px; margin: 20px 0; border-radius: 4px;">
    <h3 style="margin-top: 0; color: #4F46E5; font-size: 16px;">API Key Best Practices</h3>
    <ul style="margin: 10px 0; padding-left: 20px; color: #666;">
      <li>Keep your API keys private and secure</li>
      <li>Use different keys for different environments</li>
      <li>Set appropriate rate limits for each key</li>
      <li>Delete unused API keys immediately</li>
      <li>Monitor API key usage regularly</li>
      <li>Never commit API keys to version control</li>
    </ul>
  </div>
  
  <p style="color: #666; font-size: 14px; margin-top: 30px;">
    If you have any questions about API keys, please refer to our documentation or contact support.
  </p>
  
  <hr style="border: none; border-top: 1px solid #e5e5e5; margin: 30px 0;">
  
  <p style="color: #999; font-size: 12px; text-align: center;">
    © 2024 IFA Labs. All rights reserved.
  </p>
</body>
</html>
`, name, keyName, currentTime, apiKeyPreview, frontendURL)
}

func (e *EmailService) buildSubscriptionActivatedEmailHTML(email, name, planID, billingCycle string, expiresAt *time.Time) string {
	frontendURL := getEnvOrDefault("FRONTEND_URL", "http://localhost:3000")
	currentTime := time.Now().Format("Monday, January 2, 2006 at 3:04 PM MST")
	
	planName := planID
	switch planID {
	case "developer":
		planName = "Developer"
	case "professional":
		planName = "Professional"
	case "enterprise":
		planName = "Enterprise"
	case "free":
		planName = "Free"
	}
	
	billingText := billingCycle
	if billingCycle == "monthly" {
		billingText = "Monthly"
	} else if billingCycle == "annual" {
		billingText = "Annual"
	} else {
		billingText = "Lifetime"
	}
	
	expiryText := "Never expires"
	if expiresAt != nil {
		expiryText = expiresAt.Format("Monday, January 2, 2006")
	}
	
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Subscription Activated - IFA Labs</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="text-align: center; margin-bottom: 30px;">
    <h1 style="color: #4F46E5; margin: 0;">IFA Labs</h1>
  </div>
  
  <h2 style="color: #1a1a1a; margin-bottom: 20px;">🎉 Subscription Activated!</h2>
  
  <p style="margin-bottom: 20px;">Hello %s,</p>
  
  <p style="margin-bottom: 20px;">Great news! Your subscription payment has been confirmed and your account has been upgraded.</p>
  
  <div style="background-color: #f0fdf4; border-left: 4px solid #10b981; padding: 16px; margin: 20px 0;">
    <p style="margin: 0; color: #065f46;">
      <strong>✓ Subscription Details</strong><br>
      Plan: <strong>%s Tier</strong><br>
      Billing Cycle: <strong>%s</strong><br>
      Activated: %s<br>
      Expires: <strong>%s</strong>
    </p>
  </div>
  
  <div style="background-color: #eff6ff; border-left: 4px solid #3b82f6; padding: 16px; margin: 20px 0;">
    <h3 style="margin-top: 0; color: #1e40af; font-size: 16px;">What's Included</h3>
    <ul style="margin: 10px 0; padding-left: 20px; color: #1e3a8a;">
      <li>Full access to Oracle Engine API</li>
      <li>Real-time price feeds</li>
      <li>Enhanced rate limits</li>
      <li>24/7 technical support</li>
      <li>Priority feature requests</li>
    </ul>
  </div>
  
  <div style="text-align: center; margin: 30px 0;">
    <a href="%s/dashboard" 
       style="background-color: #4F46E5; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: bold; margin-right: 10px;">
      Go to Dashboard
    </a>
    <a href="%s/api-keys" 
       style="background-color: #10b981; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: bold;">
      Create API Keys
    </a>
  </div>
  
  <div style="background-color: #f9fafb; padding: 16px; margin: 20px 0; border-radius: 4px;">
    <h3 style="margin-top: 0; color: #4F46E5; font-size: 16px;">Next Steps</h3>
    <ol style="margin: 10px 0; padding-left: 20px; color: #666;">
      <li>Create your first API key from the dashboard</li>
      <li>Review our API documentation</li>
      <li>Start integrating Oracle Engine into your application</li>
      <li>Monitor your usage and API limits</li>
    </ol>
  </div>
  
  <div style="background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 16px; margin: 20px 0;">
    <p style="margin: 0; color: #92400e;">
      <strong>💡 Tip:</strong> Your subscription will automatically renew on %s. You can manage your subscription settings from your account page.
    </p>
  </div>
  
  <p style="color: #666; font-size: 14px; margin-top: 30px;">
    If you have any questions or need assistance, please don't hesitate to contact our support team.
  </p>
  
  <hr style="border: none; border-top: 1px solid #e5e5e5; margin: 30px 0;">
  
  <p style="color: #999; font-size: 12px; text-align: center;">
    © 2024 IFA Labs. All rights reserved.
  </p>
</body>
</html>
`, name, planName, billingText, currentTime, expiryText, frontendURL, frontendURL, expiryText)
}

// ValidateEmail performs basic email validation
func ValidateEmail(email string) bool {
	// Basic email validation
	if email == "" {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}
	if !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

