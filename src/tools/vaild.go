package tools

import (
	"regexp"
	"strings"
)

// ValidURL 验证url
func ValidURL(v string) bool {
	v = strings.ToLower(v)
	matcher, _ := regexp.Compile(`^(http|https)://[a-zA-Z0-9-_]+\..+`)
	return matcher.MatchString(v)
}
