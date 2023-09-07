package jwt

import "strings"

// StripBearerPrefixFromToken remove the bearer prefix in token string.
func StripBearerPrefixFromToken(token string) string {
	if strings.ToUpper(token[0:7]) == "BEARER " {
		return token[7:]
	}

	return token
}
