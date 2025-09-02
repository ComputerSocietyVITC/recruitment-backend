package utils

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

type Mailer struct {
	config mailerConfig
}

type mailerConfig struct {
	Host     string
	From     string
	Username string
	Password string
	Port     int
}

var MailerInstance *Mailer

func GetMailerInstance() *Mailer {
	if MailerInstance == nil {
		MailerInstance = CreateMailerInstance()
	}
	return MailerInstance
}

func CreateMailerInstance() *Mailer {
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		log.Printf("Warning: SMTP_PORT is not a valid integer, defaulting to 587. Error: %v", err)
		port = 587
	}

	config := mailerConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     port,
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
	}

	return &Mailer{config: config}
}

func (m *Mailer) Send(toList []string, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", m.config.Host, m.config.Port)
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Quit()

	tlsConfig := &tls.Config{ServerName: m.config.Host}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	auth := smtp.PlainAuth("", m.config.Username, m.config.Password, m.config.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	if err := client.Mail(m.config.From); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	for _, recipient := range toList {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("RCPT TO for %s failed: %w", recipient, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}
	defer w.Close()

	message := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n"+
			"%s",
		m.config.From, strings.Join(toList, ","), subject, body,
	)

	if _, err := w.Write([]byte(message)); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	return nil
}
