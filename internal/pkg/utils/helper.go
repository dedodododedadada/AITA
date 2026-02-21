package utils

import (
	"math/rand/v2"
	"regexp"
	"time"
)

 func StringPtr(s string) *string {
 	if s == "" {
 		return nil
 	}
 	return &s
}

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
func IsValidEmail(s string) bool {
	if len(s) < 3 || len(s) > 255 {
		return false
	}
	return emailRegex.MatchString(s)
}

func randomizeTD(t time.Duration) time.Duration {
	if t <= 0 {
		return 0
	}
	
	base := time.Duration(float64(t) * 0.9)
	jitterRange := time.Duration(float64(t) * 0.2)

	if jitterRange <= 0 {
		return base
	}

	return base + rand.N(jitterRange+1)
}

func GetRandomExpiration(base, jitter time.Duration) time.Duration {
	if base <= 0 {
		return randomizeTD(1 * time.Hour)
	}

	rBase := randomizeTD(base)

	if jitter <= 0 {
		return rBase
	}

	return rBase + rand.N(jitter+1)
}