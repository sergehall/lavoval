package mailer

import "fmt"

type DeliveryErrorKind string

const (
	DeliveryErrorTemporary DeliveryErrorKind = "temporary"
	DeliveryErrorPermanent DeliveryErrorKind = "permanent"
)

type DeliveryError struct {
	Kind DeliveryErrorKind
	Code string
	Err  error
}

func (e *DeliveryError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return fmt.Sprintf("mail delivery %s", e.Code)
	}
	return fmt.Sprintf("mail delivery %s: %v", e.Code, e.Err)
}

func (e *DeliveryError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *DeliveryError) Retryable() bool {
	return e != nil && e.Kind == DeliveryErrorTemporary
}

func temporaryDeliveryError(code string, err error) error {
	return &DeliveryError{Kind: DeliveryErrorTemporary, Code: code, Err: err}
}

func permanentDeliveryError(code string, err error) error {
	return &DeliveryError{Kind: DeliveryErrorPermanent, Code: code, Err: err}
}
