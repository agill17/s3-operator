package vault

import (
	"errors"
	"fmt"
	vaultApi "github.com/hashicorp/vault/api"
	"os"
	"strconv"
	"time"
)

const (
	saJwtFile          = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	K8sAuthBackendPath = "VAULT_K8S_AUTH_BACKEND_PATH"
	K8sAuthBackendRole = "VAULT_K8S_AUTH_BACKEND_ROLE"
)

type vaultLogicalClientInterface interface {
	ReadWithData(path string, data map[string][]string) (*vaultApi.Secret, error)
}

type Client struct {
	vaultClient          *vaultApi.Client
	vaultLogicalClient   vaultLogicalClientInterface
	CreatedLeaseAt       time.Time
	LeaseDuration        int
	ExpectedLeaseToEndAt time.Time
}

// returns a vaultClient -- assumes k8s auth backend for authN
// providerCreds will be passed from Provider.Spec.Credentials
// which will have all vault keys needed to authenticate against k8s backend
// providerCreds must have the following keys for successful authentication
// 	VAULT_ADDR: <OPTIONAL> -- you may set this in provider-cr or can be specified as a global env var in operator pod
// 	VAULT_SKIP_VERIFY: <OPTIONAL> -- you may set this in provider-cr or can be specified as a global env var in operator pod
// 	VAULT_K8S_AUTH_BACKEND_PATH: <REQUIRED> -- path + authBackend name of kubernetes. For ex: auth/kubernetes
// 	VAULT_K8S_AUTH_BACKEND_ROLE: <REQUIRED> -- role to authenticate against that exists in that k8s auth backend
func NewVaultClient(providerCreds map[string][]byte) (*Client, error) {
	authBackendPath, authBackendRole, err := getVaultK8sBackendInfoFromProvider(providerCreds)
	if err != nil {
		return nil, err
	}
	vClient, errCreatingClient := vaultApi.NewClient(getVaultCfg(providerCreds))
	if errCreatingClient != nil {
		return nil, errCreatingClient
	}

	saJwt, errGettingSaJwt := getSaJWT()
	if errGettingSaJwt != nil {
		return nil, errGettingSaJwt
	}

	vaultLoginData := map[string]interface{}{}
	vaultLoginData["role"] = authBackendRole
	vaultLoginData["jwt"] = saJwt
	authBackendLoginPath := fmt.Sprintf("%s/login", authBackendPath)
	loginSecret, errLogin := vClient.Logical().Write(authBackendLoginPath, vaultLoginData)
	if errLogin != nil {
		return nil, errLogin
	}
	vClient.SetToken(loginSecret.Auth.ClientToken)
	created := time.Now()
	c := &Client{
		vaultClient:          vClient,
		vaultLogicalClient:   vClient.Logical(),
		CreatedLeaseAt:       created,
		LeaseDuration:        loginSecret.Auth.LeaseDuration,
		ExpectedLeaseToEndAt: created.Add(time.Second * time.Duration(loginSecret.Auth.LeaseDuration)),
	}
	return c, nil
}

// getVaultCfg is a helper func that returns a vault config to setup a vault client
// If the input map has VAULT_ADDR and VAULT_SKIP_VERIFY, the default config is overridden with those values
func getVaultCfg(authInfo map[string][]byte) *vaultApi.Config {
	vaultCfg := vaultApi.DefaultConfig()
	if val, ok := authInfo[vaultApi.EnvVaultAddress]; ok {
		vaultCfg.Address = string(val)
	}
	if val, ok := authInfo[vaultApi.EnvVaultInsecure]; ok {
		b, _ := strconv.ParseBool(string(val))
		vaultCfg.ConfigureTLS(&vaultApi.TLSConfig{Insecure: b})
	}
	return vaultCfg
}

// getVaultK8sBackendInfoFromProvider is a helper func that returns k8sAuthBackendPath, roleName and a error
// If the input map does not have VAULT_K8S_AUTH_BACKEND_PATH or VAULT_K8S_AUTH_BACKEND_ROLE then a error is returned
// VAULT_K8S_AUTH_BACKEND_ROLE and VAULT_K8S_AUTH_BACKEND_PATH are REQUIRED to authenticate against a k8s vault auth backend
func getVaultK8sBackendInfoFromProvider(providerCreds map[string][]byte) (string, string, error) {
	authBackendPath, found := providerCreds[K8sAuthBackendPath]
	if !found {
		return "", "", errors.New("ErrProviderCredentialsMissingVaultBackend")
	}
	authBackendRole, roleFound := providerCreds[K8sAuthBackendRole]
	if !roleFound {
		return "", "", errors.New("ErrProviderCredentialsMissingVaultRole")
	}
	return string(authBackendPath), string(authBackendRole), nil
}

// getSaJWT returns the serviceAccount JWT this pod is running with
func getSaJWT() (string, error) {
	fileContents, err := os.ReadFile(saJwtFile)
	return string(fileContents), err
}
