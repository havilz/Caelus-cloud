package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

// Config mendefinisikan parameter konfigurasi server SMTP.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// Client mengelola pengiriman email berbasis SMTP.
type Client struct {
	cfg Config
}

// NewClient membuat instance baru Client pengirim email SMTP.
// Parameter cfg memuat konfigurasi host, port, username, password, dan email pengirim.
// Mengembalikan pointer *Client.
func NewClient(cfg Config) *Client {
	if cfg.From == "" {
		cfg.From = "no-reply@caelus.cloud"
	}
	return &Client{cfg: cfg}
}

// EmailMessage merepresentasikan konten pesan email yang akan dikirimkan.
type EmailMessage struct {
	To      string
	Subject string
	Body    string // HTML atau Plaintext body
	IsHTML  bool
}

// SendEmail mengirimkan email ke alamat tujuan via SMTP.
// Parameter ctx merupakan context pemanggilan.
// Parameter msg memuat tujuan, subjek, dan isi email.
// Mengembalikan error jika otentikasi atau transmisi SMTP gagal.
func (c *Client) SendEmail(ctx context.Context, msg EmailMessage) error {
	if msg.To == "" {
		return fmt.Errorf("recipient email address cannot be empty")
	}
	if msg.Subject == "" {
		msg.Subject = "Caelus Cloud Notification"
	}

	// Jika server SMTP tidak dikonfigurasi (misal di lokal/dev), log simulasi pengiriman
	if c.cfg.Host == "" {
		return nil
	}

	contentType := "text/plain; charset=UTF-8"
	if msg.IsHTML {
		contentType = "text/html; charset=UTF-8"
	}

	header := make(map[string]string)
	header["From"] = c.cfg.From
	header["To"] = msg.To
	header["Subject"] = msg.Subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = contentType
	header["Date"] = time.Now().Format(time.RFC1123Z)

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + msg.Body

	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)
	var auth smtp.Auth
	if c.cfg.Username != "" && c.cfg.Password != "" {
		auth = smtp.PlainAuth("", c.cfg.Username, c.cfg.Password, c.cfg.Host)
	}

	toAddrs := strings.Split(msg.To, ",")
	for i, addr := range toAddrs {
		toAddrs[i] = strings.TrimSpace(addr)
	}

	err := smtp.SendMail(addr, auth, c.cfg.From, toAddrs, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send SMTP email: %w", err)
	}

	return nil
}

// BuildAlertHTMLTemplate menghasilkan template HTML email yang bersih untuk notifikasi insiden atau otomasi.
func BuildAlertHTMLTemplate(title, ruleName, triggerEvent, details string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #0d0d0d; color: #ededed; margin: 0; padding: 24px; }
    .container { max-width: 600px; margin: 0 auto; background-color: #171717; border: 1px solid #2e2e2e; border-radius: 12px; overflow: hidden; }
    .header { padding: 20px 24px; background-color: #1a1a1a; border-bottom: 1px solid #2e2e2e; }
    .header h2 { margin: 0; font-size: 18px; color: #10b981; }
    .content { padding: 24px; }
    .card { background-color: #141414; border: 1px solid #262626; border-radius: 8px; padding: 16px; margin: 16px 0; }
    .label { font-size: 11px; text-transform: uppercase; color: #707070; margin-bottom: 4px; font-weight: 600; }
    .value { font-size: 14px; color: #ededed; font-weight: 500; font-family: monospace; }
    .footer { padding: 16px 24px; background-color: #141414; border-top: 1px solid #262626; font-size: 11px; color: #707070; text-align: center; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h2>Caelus Cloud Automation Alert</h2>
    </div>
    <div class="content">
      <p style="margin-top:0; font-size: 14px; color: #ededed;">%s</p>
      <div class="card">
        <div class="label">Triggered Rule</div>
        <div class="value">%s</div>
      </div>
      <div class="card">
        <div class="label">Event Trigger</div>
        <div class="value">%s</div>
      </div>
      <div class="card">
        <div class="label">Execution Details</div>
        <div class="value" style="white-space: pre-wrap;">%s</div>
      </div>
    </div>
    <div class="footer">
      Sent automatically by Caelus Cloud Automation Engine &bull; %s
    </div>
  </div>
</body>
</html>`, title, ruleName, triggerEvent, details, time.Now().UTC().Format(time.RFC1123))
}
