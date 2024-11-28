package utils

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"
)

func GenerateCode() string {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(1000000) // Generates a 6-digit code
	return fmt.Sprintf("%06d", code)
}

func IsValidEmail(email string) bool {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}
