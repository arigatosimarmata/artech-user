package email

import (
	"fmt"
	"net/smtp"
)

// EmailSender handles email sending operations
type EmailSender struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	emailFrom    string
}

// NewEmailSender creates a new email sender
func NewEmailSender(smtpHost, smtpPort, smtpUsername, smtpPassword, emailFrom string) *EmailSender {
	return &EmailSender{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		emailFrom:    emailFrom,
	}
}

// SendForgotPasswordEmail sends a forgot password email
func (s *EmailSender) SendForgotPasswordEmail(to, resetToken string) error {
	subject := "Password Reset Request"
	resetURL := fmt.Sprintf("http://localhost:8080/api/v1/auth/reset-password?token=%s", resetToken)
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Password Reset Request</h2>
			<p>You have requested to reset your password. Click the link below to reset your password:</p>
			<p><a href="%s">Reset Password</a></p>
			<p>This link will expire in 1 hour.</p>
			<p>If you did not request this, please ignore this email.</p>
		</body>
		</html>
	`, resetURL)

	return s.sendEmail(to, subject, body)
}

// SendWelcomeEmail sends a welcome email to new users
func (s *EmailSender) SendWelcomeEmail(to, fullName string) error {
	subject := "Welcome to Artech"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Welcome to Artech, %s!</h2>
			<p>Thank you for registering. We're excited to have you on board.</p>
			<p>You can now login and start using our services.</p>
		</body>
		</html>
	`, fullName)

	return s.sendEmail(to, subject, body)
}

// sendEmail sends an email using SMTP
func (s *EmailSender) sendEmail(to, subject, body string) error {
	// For development, just log the email instead of actually sending it
	// In production, implement actual SMTP sending
	fmt.Printf("Sending email to %s\nSubject: %s\nBody: %s\n", to, subject, body)
	
	// Actual SMTP implementation (commented out for development):
	/*
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
	
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body))
	
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	return smtp.SendMail(addr, auth, s.emailFrom, []string{to}, msg)
	*/
	
	return nil
}
