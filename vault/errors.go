package vault

// A generic error to indicate reconcile a requeue could potentially fix this error next time
type ErrRequeueNeeded struct {
	Message string
}

func (e ErrRequeueNeeded) Error() string { return e.Message }
