package validate

import (
	"errors"
	"regexp"
)

func Token(token string) (string, error) {
	const pattern = `^[a-f0-9]{40}$`
	re := regexp.MustCompile(pattern)
	if re.MatchString(token) {
		return token, nil
	}
	return "", errors.New("invalid token format")
}
