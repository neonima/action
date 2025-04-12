package action

import "errors"

var (
	ErrAlreadyStarted = errors.New("runner already started")
	ErrNilContext     = errors.New("context is nil")
)
