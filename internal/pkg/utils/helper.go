package utils

import "regexp"

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