package aws

type ErrAWSFailedToExtractCredentialsFromK8sSecret struct {
	Message string
}

func (e *ErrAWSFailedToExtractCredentialsFromK8sSecret) Error() string {
	return e.Message
}
