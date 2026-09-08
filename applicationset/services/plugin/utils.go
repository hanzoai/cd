package plugin

import (
	"strings"

	"github.com/hanzoai/cd/common"
)

// ParseSecretKey retrieves secret appSetName if different from common SecretName.
func ParseSecretKey(key string) (secretName string, tokenKey string) {
	if strings.Contains(key, ":") {
		parts := strings.Split(key, ":")
		secretName = parts[0][1:]
		tokenKey = "$" + parts[1]
	} else {
		secretName = common.SecretName
		tokenKey = key
	}
	return secretName, tokenKey
}
