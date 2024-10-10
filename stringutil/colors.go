package stringutil

import "regexp"

// IsValidHexCode Returns if a string is a valid hex code
func IsValidHexCode(code string) bool {
	re := regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
	return re.MatchString(code)
}
