package terraform

import "errors"

// Errors that might be returned from procesing the HCL
var (
	ErrNoQueries       = errors.New("no terraform entities found, looks empty")
	ErrNoKnownProvider = errors.New("terraform providers are not yet supported")
	ErrNoProviders     = errors.New("no valid providers found")
)
