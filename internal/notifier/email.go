package notifier

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"value-sniffer-radar/internal/config"
)

type Email struct {
	host          string
	port          int
	username      string
	password      string
	from          string
	to            []string
	subjectPrefix string
}

func NewEmail(c config.NotifierConfig) (*Email, error) {
	if c.SMTPHost == "" || c.SMTPPort == 0 {
		return nil, errors.New("email.smtp_host and email.smtp_port required")
	}
	user := ""
	pass := ""
	if c.UsernameEnv != "" {
		user = mustEnvOrEmpty(c.UsernameEnv)
	}
	if c.PasswordEnv != "" {
		pass = mustEnvOrEmpty(c.PasswordEnv)
	}
	if user == "" || pass == "" {
		return nil, errors.New("email.username_env and email.password_env must be set and present in environment")
	}
	if c.From == "" || len(c.To) == 0 {
		return nil, errors.New("email.from and email.to required")
	}
	return &Email{
		host:          c.SMTPHost,
		port:          c.SMTPPort,
		username:      user,
		password:      pass,
		from:          c.From,
		to:            c.To,
		subjectPrefix: c.SubjectPrefix,
	}, nil
}

func (e *Email) Name() string { return "email" }

func (e *Email) Notify(_ context.Context, events []Event) error {
	subject := e.subjectPrefix
	if subject == "" {
		subject = "[ValueSniffer]"
	}
	subject = strings.TrimSpace(subject)
	if !strings.HasPrefix(subject, "[") {
		subject = "[ValueSniffer] " + subject
	}
	subject = subject + fmt.Sprintf(" %d alerts", len(events))

	body := JSON(events)
	msg := buildRFC822(e.from, e.to, subject, body)

	addr := fmt.Sprintf("%s:%d", e.host, e.port)

	// Support implicit TLS for common ports (465).
	if e.port == 465 {
		return e.sendSMTPTLS(addr, msg)
	}
	return e.sendSMTPPlain(addr, msg)
}

func (e *Email) sendSMTPPlain(addr, msg string) error {
	auth := smtp.PlainAuth("", e.username, e.password, e.host)
	return smtp.SendMail(addr, auth, e.from, e.to, []byte(msg))
}

func (e *Email) sendSMTPTLS(addr, msg string) error {
	d := net.Dialer{Timeout: 15 * time.Second}
	conn, err := tls.DialWithDialer(&d, "tcp", addr, &tls.Config{
		ServerName: e.host,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, e.host)
	if err != nil {
		return err
	}
	defer c.Quit()

	auth := smtp.PlainAuth("", e.username, e.password, e.host)
	if err := c.Auth(auth); err != nil {
		return err
	}
	if err := c.Mail(e.from); err != nil {
		return err
	}
	for _, rcpt := range e.to {
		if err := c.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

func buildRFC822(from string, to []string, subject string, body string) string {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + strings.Join(to, ",") + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: application/json; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	b.WriteString("\r\n")
	return b.String()
}

func mustEnvOrEmpty(k string) string {
	v, ok := LookupEnv(k)
	if !ok {
		return ""
	}
	return v
}
