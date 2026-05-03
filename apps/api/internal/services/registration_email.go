package services

import (
	"bytes"
	"fmt"
	"html"
	"mime"
	"net/mail"
	"net/smtp"
	"strings"

	"test-iq-ku/apps/api/internal/models"
)

const (
	lineCommunityURL     = "https://line.me/R/ti/g/smvw5u7zP6"
	telegramCommunityURL = "https://t.me/TryOutKNMPdanKDMP"
	whatsappCommunityURL = "https://chat.whatsapp.com/DuXmlJrighGJ5GZ3vLLd19?mode=gi_t"
)

func (s *AuthService) sendRegistrationCommunityEmail(user models.User) error {
	if strings.TrimSpace(s.cfg.SMTPHost) == "" ||
		strings.TrimSpace(s.cfg.SMTPUsername) == "" ||
		strings.TrimSpace(s.cfg.SMTPAppPassword) == "" {
		return nil
	}

	fromEmail := strings.TrimSpace(s.cfg.EmailFrom)
	if fromEmail == "" {
		fromEmail = strings.TrimSpace(s.cfg.SMTPUsername)
	}

	from := mail.Address{
		Name:    strings.TrimSpace(s.cfg.EmailFromName),
		Address: fromEmail,
	}
	to := mail.Address{
		Name:    strings.TrimSpace(user.Name),
		Address: strings.TrimSpace(user.Email),
	}

	message := buildRegistrationCommunityEmail(from, to)
	addr := fmt.Sprintf("%s:%d", strings.TrimSpace(s.cfg.SMTPHost), s.cfg.SMTPPort)
	auth := smtp.PlainAuth("", strings.TrimSpace(s.cfg.SMTPUsername), s.cfg.SMTPAppPassword, strings.TrimSpace(s.cfg.SMTPHost))

	return smtp.SendMail(addr, auth, from.Address, []string{to.Address}, message)
}

func buildRegistrationCommunityEmail(from mail.Address, to mail.Address) []byte {
	subject := "Selamat bergabung di Try Out KNMP dan KDMP"
	recipientName := strings.TrimSpace(to.Name)
	if recipientName == "" {
		recipientName = "Peserta"
	}

	plainBody := fmt.Sprintf(`Halo %s,

Terima kasih sudah mendaftar di Try Out KNMP dan KDMP.

Silakan bergabung ke grup komunitas melalui salah satu link berikut:

Line: %s
Telegram: %s
WA: %s

Salam,
Try Out KNMP dan KDMP
`, recipientName, lineCommunityURL, telegramCommunityURL, whatsappCommunityURL)

	htmlBody := fmt.Sprintf(`<!doctype html>
<html>
  <body style="font-family: Arial, sans-serif; color: #0f172a; line-height: 1.6;">
    <p>Halo %s,</p>
    <p>Terima kasih sudah mendaftar di <strong>Try Out KNMP dan KDMP</strong>.</p>
    <p>Silakan bergabung ke grup komunitas melalui salah satu link berikut:</p>
    <ul>
      <li><strong>Line:</strong> <a href="%s">%s</a></li>
      <li><strong>Telegram:</strong> <a href="%s">%s</a></li>
      <li><strong>WA:</strong> <a href="%s">%s</a></li>
    </ul>
    <p>Salam,<br>Try Out KNMP dan KDMP</p>
  </body>
</html>
`, html.EscapeString(recipientName), lineCommunityURL, lineCommunityURL, telegramCommunityURL, telegramCommunityURL, whatsappCommunityURL, whatsappCommunityURL)

	var message bytes.Buffer
	boundary := "registration-community-links"
	message.WriteString("From: " + from.String() + "\r\n")
	message.WriteString("To: " + to.String() + "\r\n")
	message.WriteString("Subject: " + mime.QEncoding.Encode("utf-8", subject) + "\r\n")
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: " + fmt.Sprintf("multipart/alternative; boundary=%q", boundary) + "\r\n")
	message.WriteString("\r\n")
	message.WriteString("--" + boundary + "\r\n")
	message.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	message.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	message.WriteString(plainBody + "\r\n")
	message.WriteString("--" + boundary + "\r\n")
	message.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	message.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	message.WriteString(htmlBody + "\r\n")
	message.WriteString("--" + boundary + "--\r\n")

	return message.Bytes()
}
