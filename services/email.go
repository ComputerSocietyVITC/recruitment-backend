package services

import (
	"time"

	"github.com/ComputerSocietyVITC/recruitment-backend/utils"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

var Mailer chan *gomail.Message

func InitMailer(logger *zap.Logger) {
	Mailer = make(chan *gomail.Message)

	host := utils.GetEnvWithDefault("SMTP_HOST", "smtp.example.com")
	port := utils.GetEnvAsInt("SMTP_PORT", 587)
	user := utils.GetEnvWithDefault("SMTP_USER", "user")
	password := utils.GetEnvWithDefault("SMTP_PASSWORD", "123456")

	d := gomail.NewDialer(host, port, user, password)

	var s gomail.SendCloser
	var err error
	open := false

	logger.Info("Starting mailer service",
		zap.String("host", host),
		zap.Int("port", port),
		zap.String("user", user))

	for {
		select {
		case m, ok := <-Mailer:
			if !ok {
				logger.Info("Mailer channel closed, shutting down")
				if open {
					if err := s.Close(); err != nil {
						logger.Error("Failed to close SMTP connection during shutdown", zap.Error(err))
					}
				}
				return
			}
			if !open {
				logger.Debug("Attempting to connect to SMTP server")
				if s, err = d.Dial(); err != nil {
					logger.Error("Failed to dial SMTP server", zap.Error(err))
					// Continue to process next message - maybe connection will work later
					continue
				}
				logger.Debug("Successfully connected to SMTP server")
				open = true
			}
			if err := gomail.Send(s, m); err != nil {
				logger.Error("Failed to send email", zap.Error(err))
				// Close connection on send error - it might be stale
				if open {
					if closeErr := s.Close(); closeErr != nil {
						logger.Error("Failed to close SMTP connection after send error", zap.Error(closeErr))
					}
					open = false
				}
			} else {
				logger.Debug("Email sent successfully")
			}
		// Close the connection to the SMTP server if no email was sent in
		// the last 30 seconds.
		case <-time.After(30 * time.Second):
			if open {
				logger.Debug("Closing idle SMTP connection")
				if err := s.Close(); err != nil {
					logger.Error("Failed to close SMTP connection", zap.Error(err))
				}
				open = false
			}
		}
	}
}

func CloseMailer(logger *zap.Logger) {
	close(Mailer)
	logger.Info("Mailer closed")
}
