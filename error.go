package niuhe

import "fmt"

type ICommError interface {
	GetCode() int
	GetMessage() string
}

type CommError struct {
	error
	Code int
	Message string
}

func NewCommError(code int, message string) *CommError {
	return &CommError {
		Code: code,
		Message: message,
	}
}

func (err *CommError) GetCode() int {
	return err.Code
}

func (err *CommError) GetMessage() string {
	return err.Message
}

func (err *CommError) Error() string {
	return fmt.Sprintf("%d:%s", err.Code, err.Message)
}
