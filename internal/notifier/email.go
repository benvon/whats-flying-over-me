package notifier

import (
	"fmt"
	"net/smtp"

	"github.com/example/whats-flying-over-me/internal/config"
)

// Email implements Notifier using SMTP.
type Email struct {
	cfg config.EmailConfig
}

// NewEmail creates a new Email notifier.
func NewEmail(cfg config.EmailConfig) *Email {
	return &Email{cfg: cfg}
}

// Notify sends an email notification.
func (e *Email) Notify(subject, body string) error {
	auth := smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.SMTPServer)
	addr := fmt.Sprintf("%s:%d", e.cfg.SMTPServer, e.cfg.SMTPPort)

	msg := []byte("To: " + e.cfg.To + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + body + "\r\n")

	return smtp.SendMail(addr, auth, e.cfg.From, []string{e.cfg.To}, msg)
}
