package vault

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cast"
)

// readSecretPath
func (v *vaultClient) ReadSecretPath(path, key string) (string, error) {
	secret, err := v.vaultLogicalClient.ReadWithData(path, map[string][]string{"version": {"-1"}})
	if err != nil {
		return "", err
	}
	secretData := secret.Data["data"]
	spew.Dump(secretData)
	stringMapSecretData := cast.ToStringMap(secretData)
	spew.Dump(stringMapSecretData)
	secretVal, found := stringMapSecretData[key]
	if !found {
		return "", errors.New("ErrVaultKeyNotFound")
	}

	return fmt.Sprint(secretVal), nil
}
