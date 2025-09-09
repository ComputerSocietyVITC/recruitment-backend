package utils

import (
	"fmt"
	"strings"
	"time"
)

// EmailTemplate contains subject and body for an email
type EmailTemplate struct {
	Subject string
	Body    string
}

// GetEmailVerificationTemplate returns the email verification template with OTP and duration
func GetEmailVerificationTemplate(otp string, duration time.Duration) EmailTemplate {
	subject := GetEnvWithDefault(
		"EMAIL_VERIFICATION_SUBJECT",
		"Thank you for applying to IEEE Computer Society VITC. Please verify your email address",
	)

	bodyTemplate := GetEnvWithDefault(
		"EMAIL_VERIFICATION_BODY",
		"Your OTP is: <strong>{{.OTP}}</strong>. It is valid for {{.DURATION}}.",
	)

	// Replace placeholders
	body := strings.ReplaceAll(bodyTemplate, "{{.OTP}}", otp)
	body = strings.ReplaceAll(body, "{{.DURATION}}", formatDuration(duration))

	return EmailTemplate{
		Subject: subject,
		Body:    body,
	}
}

// GetResendVerificationTemplate returns the resend verification template with OTP and duration
func GetResendVerificationTemplate(otp string, duration time.Duration) EmailTemplate {
	subject := GetEnvWithDefault(
		"EMAIL_RESEND_VERIFICATION_SUBJECT",
		"IEEE Computer Society VITC - New Verification Code",
	)

	bodyTemplate := GetEnvWithDefault(
		"EMAIL_RESEND_VERIFICATION_BODY",
		"Your new OTP is: <strong>{{.OTP}}</strong>. It is valid for {{.DURATION}}.",
	)

	// Replace placeholders
	body := strings.ReplaceAll(bodyTemplate, "{{.OTP}}", otp)
	body = strings.ReplaceAll(body, "{{.DURATION}}", formatDuration(duration))

	return EmailTemplate{
		Subject: subject,
		Body:    body,
	}
}

// GetPasswordResetTemplate returns the password reset template with token and duration
func GetPasswordResetTemplate(resetToken string, duration time.Duration) EmailTemplate {
	subject := GetEnvWithDefault(
		"EMAIL_PASSWORD_RESET_SUBJECT",
		"IEEE Computer Society VITC - Password Reset Request",
	)

	bodyTemplate := GetEnvWithDefault(
		"EMAIL_PASSWORD_RESET_BODY",
		"You have requested to reset your password. Your reset token is: <strong>{{.TOKEN}}</strong>. This token is valid for {{.DURATION}}. If you did not request this reset, please ignore this email.",
	)

	// Replace placeholders
	body := strings.ReplaceAll(bodyTemplate, "{{.TOKEN}}", resetToken)
	body = strings.ReplaceAll(body, "{{.DURATION}}", formatDuration(duration))

	return EmailTemplate{
		Subject: subject,
		Body:    body,
	}
}

// GetPasswordResetSuccessTemplate returns the password reset success template
func GetPasswordResetSuccessTemplate() EmailTemplate {
	subject := GetEnvWithDefault(
		"EMAIL_PASSWORD_RESET_SUCCESS_SUBJECT",
		"IEEE Computer Society VITC - Password Reset Successful",
	)

	body := GetEnvWithDefault(
		"EMAIL_PASSWORD_RESET_SUCCESS_BODY",
		"Your password has been successfully reset. If you did not perform this action, please contact support immediately.",
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
	}
}

// formatDuration converts time.Duration to a human-readable string
func formatDuration(d time.Duration) string {
	if d >= time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}

	minutes := int(d.Minutes())
	if minutes == 1 {
		return "1 minute"
	}
	return fmt.Sprintf("%d minutes", minutes)
}
