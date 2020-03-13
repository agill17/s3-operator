package customErrors

type ErrorIAMK8SSecretNeedsUpdate struct {
	Message string
}

func (e ErrorIAMK8SSecretNeedsUpdate) Error() string {
	return e.Message
}
