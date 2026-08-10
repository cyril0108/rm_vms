package email

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Sender holds the SMTP connection parameters.
// Built from DB settings at send-time so configuration changes take effect immediately.
type Sender struct {
	Host        string
	Port        int
	Username    string
	Password    string
	SenderEmail string
	SenderName  string
	UseTLS      bool
}

// Message represents an outgoing email.
// If HTML is non-empty, the email is sent as multipart/alternative (plain + HTML).
type Message struct {
	To      []string
	Subject string
	Body    string // plain text
	HTML    string // html body (optional)
}

// Send delivers an email message via the configured SMTP server.
// It automatically selects the appropriate transport:
//   - Port 465 with TLS: direct TLS (implicit)
//   - Port 587 with TLS: STARTTLS
//   - Otherwise: plain SMTP
func (s *Sender) Send(msg *Message) error {
	body := s.buildMessage(msg)
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)

	// Port 465 = implicit TLS (direct TLS connection)
	if s.UseTLS && s.Port == 465 {
		return s.sendDirectTLS(addr, auth, msg.To, []byte(body))
	}

	// Port 587 (or others) with TLS = STARTTLS upgrade
	if s.UseTLS {
		return s.sendSTARTTLS(addr, auth, msg.To, []byte(body))
	}

	// Plain SMTP (no encryption)
	return smtp.SendMail(addr, auth, s.SenderEmail, msg.To, []byte(body))
}

// SendTest sends a predefined test email to verify SMTP configuration.
func (s *Sender) SendTest(to string) error {
	return s.Send(&Message{
		To:      []string{to},
		Subject: "NVR 測試郵件 — Test Email",
		Body:    "這是一封測試郵件，表示您的 SMTP 設定正確。\n\nThis is a test email confirming your SMTP settings are working correctly.",
		HTML: `<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
	<div style="background: #1976d2; color: white; padding: 16px 24px; border-radius: 8px 8px 0 0;">
		<h2 style="margin: 0;">✅ NVR 測試郵件</h2>
	</div>
	<div style="background: #f5f5f5; padding: 24px; border-radius: 0 0 8px 8px;">
		<p>這是一封測試郵件，表示您的 SMTP 設定正確。</p>
		<p>This is a test email confirming your SMTP settings are working correctly.</p>
	</div>
</div>`,
	})
}

// ──────────────────────────────────────────────
// Internal helpers
// ──────────────────────────────────────────────

// buildMessage constructs the raw RFC 2822 email body.
func (s *Sender) buildMessage(msg *Message) string {
	from := s.formatFrom()
	to := strings.Join(msg.To, ", ")
	subject := mime.QEncoding.Encode("UTF-8", msg.Subject)

	if msg.HTML != "" {
		return s.buildMultipartMessage(from, to, subject, msg.Body, msg.HTML)
	}

	return fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=\"UTF-8\"\r\n"+
			"\r\n"+
			"%s",
		from, to, subject, msg.Body,
	)
}

// buildMultipartMessage constructs a multipart/alternative email with both plain text and HTML parts.
func (s *Sender) buildMultipartMessage(from, to, subject, plainBody, htmlBody string) string {
	boundary := fmt.Sprintf("==NVR_BOUNDARY_%d==", time.Now().UnixNano())

	return fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: multipart/alternative; boundary=\"%s\"\r\n"+
			"\r\n"+
			"--%s\r\n"+
			"Content-Type: text/plain; charset=\"UTF-8\"\r\n"+
			"\r\n"+
			"%s\r\n"+
			"--%s\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
			"\r\n"+
			"%s\r\n"+
			"--%s--\r\n",
		from, to, subject,
		boundary,
		boundary, plainBody,
		boundary, htmlBody,
		boundary,
	)
}

// formatFrom returns the RFC 5322 formatted sender address.
func (s *Sender) formatFrom() string {
	if s.SenderName != "" {
		return fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("UTF-8", s.SenderName), s.SenderEmail)
	}
	return s.SenderEmail
}

// sendDirectTLS connects via implicit TLS (port 465).
func (s *Sender) sendDirectTLS(addr string, auth smtp.Auth, to []string, body []byte) error {
	tlsConfig := &tls.Config{ServerName: s.Host}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer client.Close()

	return s.sendViaClient(client, auth, to, body)
}

// sendSTARTTLS connects via plain TCP then upgrades to TLS (port 587).
func (s *Sender) sendSTARTTLS(addr string, auth smtp.Auth, to []string, body []byte) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer client.Close()

	tlsConfig := &tls.Config{ServerName: s.Host}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("STARTTLS failed: %w", err)
	}

	return s.sendViaClient(client, auth, to, body)
}

// sendViaClient performs the SMTP dialogue (AUTH → MAIL → RCPT → DATA → QUIT).
func (s *Sender) sendViaClient(client *smtp.Client, auth smtp.Auth, to []string, body []byte) error {
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}

	if err := client.Mail(s.SenderEmail); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("RCPT TO (%s) failed: %w", addr, err)
		}
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}

	if _, err := wc.Write(body); err != nil {
		return fmt.Errorf("write body failed: %w", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("close data failed: %w", err)
	}

	return client.Quit()
}
