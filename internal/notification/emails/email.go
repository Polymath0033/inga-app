package emails

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/mail"
	"net/smtp"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"example.com/inga/internal/config"
)

type Message struct {
	To       []string
	Cc       []string
	Bcc      []string
	Subject  string
	Template string
	Data     any
}

type Sender interface {
	Send(context.Context, Message, string) error
	SendTemplate(context.Context, Message) error
}

type Service struct {
	cfg config.EmailConfig
}

func New(cfg config.EmailConfig) (*Service, error) {
	if cfg.SMTPHost == "" {
		return nil, errors.New("smtp host is required")
	}
	if cfg.SMTPPort == 0 {
		cfg.SMTPPort = 465
	}
	if cfg.FromEmail == "" {
		cfg.FromEmail = cfg.SMTPUsername
	}
	if cfg.FromEmail == "" {
		return nil, errors.New("smtp from email is required")
	}
	if cfg.SMTPUsername == "" {
		return nil, errors.New("smtp username is required")
	}
	if cfg.SMTPPassword == "" {
		return nil, errors.New("smtp password is required")
	}
	if cfg.TemplatesDir == "" {
		cfg.TemplatesDir = "internal/notification/emails/templates"
	}

	return &Service{cfg: cfg}, nil
}

func (s *Service) Send(ctx context.Context, msg Message, htmlBody string) error {
	return s.send(ctx, msg, htmlBody)
}

func (s *Service) SendTemplate(ctx context.Context, msg Message) error {
	body, err := s.renderTemplate(msg.Template, msg.Data)
	if err != nil {
		return err
	}
	return s.send(ctx, msg, body)
}

func (s *Service) renderTemplate(name string, data any) (string, error) {
	if name == "" {
		return "", errors.New("template name is required")
	}

	if data == nil {
		data = map[string]any{}
	}

	path := name
	if !filepath.IsAbs(name) {
		path = filepath.Join(s.cfg.TemplatesDir, name)
	}

	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return "", err
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
		return "", err
	}

	return rendered.String(), nil
}

func (s *Service) send(ctx context.Context, msg Message, htmlBody string) error {
	if len(msg.To) == 0 {
		return errors.New("at least one recipient is required")
	}
	if msg.Subject == "" {
		return errors.New("email subject is required")
	}

	from := s.cfg.FromEmail
	if from == "" {
		from = s.cfg.SMTPUsername
	}

	addr := net.JoinHostPort(s.cfg.SMTPHost, strconv.Itoa(s.cfg.SMTPPort))
	timeout := s.cfg.Timeout
	if timeout == 0 {
		timeout = 15 * time.Second
	}
	dialer := &net.Dialer{Timeout: timeout}
	rawConn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("dial smtp %s: %w", addr, err)
	}
	defer rawConn.Close()

	var client *smtp.Client
	// Use implicit TLS on port 465 (SMTPS), otherwise plain then STARTTLS
	if s.cfg.SMTPPort == 465 {
		tlsConn := tls.Client(rawConn, &tls.Config{ServerName: s.cfg.SMTPHost})
		if err := tlsConn.Handshake(); err != nil {
			return fmt.Errorf("tls handshake: %w", err)
		}
		client, err = smtp.NewClient(tlsConn, s.cfg.SMTPHost)
		if err != nil {
			return fmt.Errorf("create smtp client (implicit tls): %w", err)
		}
	} else {
		client, err = smtp.NewClient(rawConn, s.cfg.SMTPHost)
		if err != nil {
			return fmt.Errorf("create smtp client: %w", err)
		}
		// If configured to use STARTTLS, upgrade the connection
		if s.cfg.UseTLS {
			hasTLS, _ := client.Extension("STARTTLS")
			if !hasTLS {
				client.Close()
				return errors.New("smtp server does not support STARTTLS")
			}
			tlsConfig := &tls.Config{ServerName: s.cfg.SMTPHost}
			if err := client.StartTLS(tlsConfig); err != nil {
				client.Close()
				return fmt.Errorf("starttls failed: %w", err)
			}
		}
	}
	defer client.Close()

	auth := smtp.PlainAuth("", s.cfg.SMTPUsername, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth failed: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return err
	}

	recipients := uniqueRecipients(msg.To, msg.Cc, msg.Bcc)
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	// ensure DATA writer is closed to complete the transaction
	defer func() {
		_ = w.Close()
	}()

	headers := make([]string, 0, 8)
	fromAddress := mail.Address{Name: s.cfg.FromName, Address: from}
	headers = append(headers,
		fmt.Sprintf("From: %s", fromAddress.String()),
		fmt.Sprintf("To: %s", strings.Join(msg.To, ", ")),
		fmt.Sprintf("Subject: %s", msg.Subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"UTF-8\"",
	)
	if len(msg.Cc) > 0 {
		headers = append(headers, fmt.Sprintf("Cc: %s", strings.Join(msg.Cc, ", ")))
	}

	if _, err := fmt.Fprint(w, strings.Join(headers, "\r\n")+"\r\n\r\n"+htmlBody); err != nil {
		return fmt.Errorf("write email body: %w", err)
	}

	return nil
}

func uniqueRecipients(groups ...[]string) []string {
	seen := make(map[string]struct{})
	recipients := make([]string, 0)
	for _, group := range groups {
		for _, address := range group {
			if _, ok := seen[address]; ok || address == "" {
				continue
			}
			seen[address] = struct{}{}
			recipients = append(recipients, address)
		}
	}
	return recipients
}
