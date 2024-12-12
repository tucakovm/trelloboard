package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateCode() string {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(1000000) // Generates a 6-digit code
	return fmt.Sprintf("%06d", code)
}
