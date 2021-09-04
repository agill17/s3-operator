package gcp

type ErrGCPFailedToExtractCredentialsFromK8sSecret struct {
	Message string
}

func (e *ErrGCPFailedToExtractCredentialsFromK8sSecret) Error() string {
	return e.Message
}
