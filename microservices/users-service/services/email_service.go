package services

import (
	"fmt"
	"log"
	"net/smtp"
	"users_module/config"
)

func SendVerificationEmail(email, code string) error {
	cfg, _ := config.LoadConfig()
	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)

	htmlMessage := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6;">
			<h2 style="color: #333;">Verification Code</h2>
			<p>Dear user,</p>
			<p>Your verification code is:</p>
			<h3 style="color: #0056b3;">%s</h3>
			<p>Please use this code to verify your account.</p>
			<br>
			<p style="font-size: 0.9em; color: #777;">If you did not request this, you can safely ignore this email.</p>
		</body>
		</html>`, code)

	plainTextMessage := fmt.Sprintf("Verification Code\n\nYour code is: %s", code)

	msg := constructEmail(email, "Your Verification Code", plainTextMessage, htmlMessage)

	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort),
		auth,
		cfg.SMTPUser,
		[]string{email},
		[]byte(msg),
	)

	if err != nil {
		log.Println("Failed to send email:", err)
		return err
	}

	return nil
}

func SendMagicLinkEmail(email, magicLink string) error {
	cfg, _ := config.LoadConfig()
	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)

	htmlMessage := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6;">
			<h2 style="color: #333;">Magic Login Link</h2>
			<p>Dear user,</p>
			<p>Click the link below to log in:</p>
			<p><a href="%s" style="color: #0056b3; text-decoration: none;">Login to Your Account</a></p>
			<br>
			<p style="font-size: 0.9em; color: #777;">If you did not request this, you can safely ignore this email.</p>
		</body>
		</html>`, magicLink)

	plainTextMessage := fmt.Sprintf("Magic Login Link\n\nClick the link below to log in:\n\n%s", magicLink)

	msg := constructEmail(email, "Your Magic Login Link", plainTextMessage, htmlMessage)

	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort),
		auth,
		cfg.SMTPUser,
		[]string{email},
		[]byte(msg),
	)

	if err != nil {
		log.Println("Failed to send magic link email:", err)
		return err
	}

	return nil
}

func SendEmail(to, subject, body string) error {
	cfg, _ := config.LoadConfig()
	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)

	// HTML message with inline styling
	htmlMessage := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6;">
			<h2 style="color: #333;">%s</h2>
			<p>%s</p>
		</body>
		</html>`, subject, body)

	plainTextMessage := body

	msg := constructEmail(to, subject, plainTextMessage, htmlMessage)

	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort),
		auth,
		cfg.SMTPUser,
		[]string{to},
		[]byte(msg),
	)

	if err != nil {
		log.Println("Failed to send email:", err)
		return err
	}

	return nil
}

func constructEmail(to, subject, plainTextMessage, htmlMessage string) string {
	return fmt.Sprintf(`To: %s
Subject: %s
MIME-Version: 1.0
Content-Type: multipart/alternative; boundary="boundary42"

--boundary42
Content-Type: text/plain; charset="utf-8"

%s

--boundary42
Content-Type: text/html; charset="utf-8"

%s

--boundary42--`, to, subject, plainTextMessage, htmlMessage)
}
