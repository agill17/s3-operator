package vault

import (
	"errors"
	"fmt"
	vaultApi "github.com/hashicorp/vault/api"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	saJwtFile          = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	K8sAuthBackendPath = "VAULT_K8S_AUTH_BACKEND_PATH"
	K8sAuthBackendRole = "VAULT_K8S_AUTH_BACKEND_ROLE"
)

var vaultClientCache sync.Map

type vaultLogicalClientInterface interface {
	ReadWithData(path string, data map[string][]string) (*vaultApi.Secret, error)
}

type vaultClient struct {
	vaultClient          *vaultApi.Client
	vaultLogicalClient   vaultLogicalClientInterface
	createdLeaseAt       time.Time
	leaseDuration        int
	expectedLeaseToEndAt time.Time
	// cache key name is set so when a re-auth is needed, some other func can delete the cached client from the cache
	cacheKeyName string
}

// returns a vaultClient -- assumes k8s auth backend for authN
// providerCreds will be passed from Provider.Spec.Credentials
// which will have all vault keys needed to authenticate against k8s backend
// providerCreds must have the following keys for successful authentication
// 	VAULT_ADDR: <OPTIONAL> -- you may set this in provider-cr or can be specified as a global env var in operator pod
// 	VAULT_SKIP_VERIFY: <OPTIONAL> -- you may set this in provider-cr or can be specified as a global env var in operator pod
// 	VAULT_K8S_AUTH_BACKEND_PATH: <REQUIRED> -- path + authBackend name of kubernetes. For ex: auth/kubernetes
// 	VAULT_K8S_AUTH_BACKEND_ROLE: <REQUIRED> -- role to authenticate against that exists in that k8s auth backend
func NewVaultClient(providerCreds map[string][]byte) (*vaultClient, error) {
	authBackendPath, authBackendRole, err := getVaultK8sBackendInfoFromProvider(providerCreds)
	if err != nil {
		return nil, err
	}
	cacheKeyName := fmt.Sprintf("%s-%s", strings.ReplaceAll(authBackendPath, "/", "-"), authBackendRole)
	cachedClient, found := vaultClientCache.Load(cacheKeyName)
	if !found {
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
		c := &vaultClient{
			vaultClient:          vClient,
			vaultLogicalClient:   vClient.Logical(),
			createdLeaseAt:       created,
			leaseDuration:        loginSecret.Auth.LeaseDuration,
			expectedLeaseToEndAt: created.Add(time.Second * time.Duration(loginSecret.Auth.LeaseDuration)),
			cacheKeyName:         cacheKeyName,
		}
		vaultClientCache.Store(cacheKeyName, c)
		return c, nil
	}
	cachedVaultClient, _ := cachedClient.(*vaultClient)

	// in case the lease is expired, delete that key from cache and force re-queue
	if time.Now().After(cachedVaultClient.expectedLeaseToEndAt) {
		vaultClientCache.Delete(cacheKeyName)
		return nil, ErrRequeueNeeded{}
	}
	return cachedClient.(*vaultClient), nil
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
