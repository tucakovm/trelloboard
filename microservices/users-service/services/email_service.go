package services

import (
	"fmt"
	"log"
	"net/smtp"
)

func SendVerificationEmail(email, code string) error {
	cfg := config.LoadConfig()
	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	msg := []byte(fmt.Sprintf("Subject: Verification Code\n\nYour code is: %s", code))
	err := smtp.SendMail(fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort), auth, cfg.SMTPUser, []string{email}, msg)
	if err != nil {
		log.Println("Failed to send email:", err)
		return err
	}
	return nil
}
