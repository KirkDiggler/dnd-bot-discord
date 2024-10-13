package internal

type BaseError string

func (e BaseError) Error() string {
	return string(e)
}

const (
	ErrMissingParam BaseError = "missing parameter"
	ErrInvalidParam BaseError = "invalid parameter"
	ErrNotFound     BaseError = "not found"
)

type MissingParamError struct {
	param string
}

func (e *MissingParamError) Error() string {
	return string(ErrMissingParam) + ": " + e.param
}

func (e *MissingParamError) Unwrap() error {
	return ErrMissingParam
}

func NewMissingParamError(param string) error {
	return &MissingParamError{
		param: param,
	}
}

type InvalidParamError struct {
	msg string
}

func (e *InvalidParamError) Error() string {
	return string(ErrInvalidParam) + ": " + e.msg
}

func (e *InvalidParamError) Unwrap() error {
	return ErrInvalidParam
}

func NewInvalidParamError(msg string) error {
	return &InvalidParamError{
		msg: msg,
	}
}