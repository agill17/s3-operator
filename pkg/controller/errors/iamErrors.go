package customErrors

type ErrorIAMK8SSecretNeedsUpdate struct {
	Message string
}

type ErrorIAMInlinePolicyNeedsUpdate struct {
	Message string
}

func (e ErrorIAMK8SSecretNeedsUpdate) Error() string {
	return e.Message
}


func (e ErrorIAMInlinePolicyNeedsUpdate) Error() string {
	return e.Message
}
