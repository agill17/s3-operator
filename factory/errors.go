package factory

type ErrIsNotBucketObject struct {
	Message string
}

func (e ErrIsNotBucketObject) Error() string { return e.Message }
