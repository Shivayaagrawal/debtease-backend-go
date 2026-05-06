package notifications

import (
	"context"
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
	"DebtEase/internal/logger"
)


// NewEmailService creates a new email service instance
func NewEmailService() (*EmailService, error) {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("RESEND_API_KEY not set in environment")
	}

	fromEmail := os.Getenv("FROM_EMAIL")
	if fromEmail == "" {
		return nil, fmt.Errorf("FROM_EMAIL not set in environment")
	}

	client := resend.NewClient(apiKey)

	return &EmailService{
		client:    client,
		fromEmail: fromEmail,
	}, nil
}

// SendWelcomeEmail sends a welcome email to a newly registered user
func (es *EmailService) SendWelcomeEmail(ctx context.Context, recipient, name string) error {
	subject := "Welcome to DebtEase!"
	
	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
				.content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 5px 5px; }
				.button { display: inline-block; padding: 12px 24px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px; margin-top: 20px; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Welcome to DebtEase!</h1>
				</div>
				<div class="content">
					<p>Hi %s,</p>
					<p>Thank you for joining DebtEase! We're excited to help you take control of your finances and manage your debts effectively.</p>
					<p>With DebtEase, you can:</p>
					<ul>
						<li>Track all your debts in one place</li>
						<li>Get personalized repayment strategies</li>
						<li>Monitor your progress and stay on track</li>
						<li>Receive timely reminders and insights</li>
					</ul>
					<p>Get started by adding your first debt to your dashboard.</p>
					<a href="#" class="button">Go to Dashboard</a>
					<p>If you have any questions, feel free to reach out to our support team.</p>
					<p>Best regards,<br>The DebtEase Team</p>
				</div>
			</div>
		</body>
		</html>
	`, name)

	plainTextBody := fmt.Sprintf(`
		Welcome to DebtEase!
		
		Hi %s,
		
		Thank you for joining DebtEase! We're excited to help you take control of your finances and manage your debts effectively.
		
		With DebtEase, you can:
		- Track all your debts in one place
		- Get personalized repayment strategies
		- Monitor your progress and stay on track
		- Receive timely reminders and insights
		
		Get started by adding your first debt to your dashboard.
		
		If you have any questions, feel free to reach out to our support team.
		
		Best regards,
		The DebtEase Team
	`, name)

	params := &resend.SendEmailRequest{
		From:    es.fromEmail,
		To:      []string{recipient},
		Subject: subject,
		Html:    htmlBody,
		Text:    plainTextBody,
	}

	sent, err := es.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		logger.Logger.Errorw("Failed to send welcome email", "error", err, "recipient", recipient)
		return fmt.Errorf("failed to send email: %w", err)
	}

	logger.Logger.Infow("Welcome email sent successfully", "email_id", sent.Id, "recipient", recipient)
	return nil
}

// SendPaymentReminder sends a payment reminder email
func (es *EmailService) SendPaymentReminder(ctx context.Context, recipient, name, debtName string, amount string, dueDate string) error {
	subject := fmt.Sprintf("Payment Reminder: %s", debtName)
	
	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #FF9800; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
				.content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 5px 5px; }
				.alert { background-color: #fff3cd; border-left: 4px solid #FF9800; padding: 15px; margin: 20px 0; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Payment Reminder</h1>
				</div>
				<div class="content">
					<p>Hi %s,</p>
					<div class="alert">
						<p><strong>Upcoming Payment Due</strong></p>
						<p>Debt: %s</p>
						<p>Amount: %s</p>
						<p>Due Date: %s</p>
					</div>
					<p>This is a friendly reminder that you have a payment due soon. Please make sure to submit your payment on time to avoid any late fees.</p>
					<p>Best regards,<br>The DebtEase Team</p>
				</div>
			</div>
		</body>
		</html>
	`, name, debtName, amount, dueDate)

	plainTextBody := fmt.Sprintf(`
		Payment Reminder
		
		Hi %s,
		
		Upcoming Payment Due:
		Debt: %s
		Amount: %s
		Due Date: %s
		
		This is a friendly reminder that you have a payment due soon. Please make sure to submit your payment on time to avoid any late fees.
		
		Best regards,
		The DebtEase Team
	`, name, debtName, amount, dueDate)

	params := &resend.SendEmailRequest{
		From:    es.fromEmail,
		To:      []string{recipient},
		Subject: subject,
		Html:    htmlBody,
		Text:    plainTextBody,
	}

	sent, err := es.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		logger.Logger.Errorw("Failed to send payment reminder", "error", err, "recipient", recipient)
		return fmt.Errorf("failed to send email: %w", err)
	}

	logger.Logger.Infow("Payment reminder sent successfully", "email_id", sent.Id, "recipient", recipient)
	return nil
}