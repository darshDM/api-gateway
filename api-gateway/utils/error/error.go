package error

import "fmt"

type GatewayError struct {
	Service string
	Message string
	Code    int
}

func (e *GatewayError) Error() string {
	return fmt.Sprintf("Backend Service '%s' failed: %s (HTTP %d)", e.Service, e.Message, e.Code)
}
