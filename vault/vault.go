package vault

import (
	"errors"
	"fmt"
	"github.com/spf13/cast"
)

// readSecretPath
func (v *Client) ReadSecretPath(path, key string) (string, error) {
	secret, err := v.vaultLogicalClient.ReadWithData(path, map[string][]string{"version": {"-1"}})
	if err != nil {
		return "", err
	}
	secretData := secret.Data["data"]
	stringMapSecretData := cast.ToStringMap(secretData)
	secretVal, found := stringMapSecretData[key]
	if !found {
		return "", errors.New("ErrVaultKeyNotFound")
	}
	return fmt.Sprint(secretVal), nil
}
