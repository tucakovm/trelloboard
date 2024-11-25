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
	msg := []byte(fmt.Sprintf("Subject: Verification Code\n\nYour code is: %s", code))
	err := smtp.SendMail(fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort), auth, cfg.SMTPUser, []string{email}, msg)

	log.Println(cfg.SMTPPassword)
	if err != nil {
		log.Println("Failed to send email:", err)
		log.Println("user email:", []string{email})
		log.Println()
		return err
	}
	return nil
}
func SendMagicLinkEmail(email, magicLink string) error {
	cfg, _ := config.LoadConfig()
	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	msg := []byte(fmt.Sprintf("Subject: Your Magic Login Link\n\nClick the link below to log in:\n\n%s", magicLink))
	err := smtp.SendMail(fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort), auth, cfg.SMTPUser, []string{email}, msg)

	if err != nil {
		log.Println("Failed to send magic link email:", err)
		return err
	}
	return nil
}
