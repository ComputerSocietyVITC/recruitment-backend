package services

import (
	"github.com/ComputerSocietyVITC/recruitment-backend/utils"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
	"time"
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
	for {
		select {
		case m, ok := <-Mailer:
			if !ok {
				return
			}
			if !open {
				if s, err = d.Dial(); err != nil {
					panic(err)
				}
				open = true
			}
			if err := gomail.Send(s, m); err != nil {
				logger.Error("Failed to send email", zap.Error(err))
			}
		// Close the connection to the SMTP server if no email was sent in
		// the last 30 seconds.
		case <-time.After(30 * time.Second):
			if open {
				if err := s.Close(); err != nil {
					panic(err)
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
