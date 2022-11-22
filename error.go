package niuhe

import "fmt"

type ICommError interface {
	error
	GetCode() int
	GetMessage() string
}

type CommError struct {
	Code    int
	Message string
}

func NewCommError(code int, message string) *CommError {
	return &CommError{
		Code:    code,
		Message: message,
	}
}

func (err *CommError) GetCode() int {
	if nil == err {
		return 0
	}
	return err.Code
}

func (err *CommError) GetMessage() string {
	if nil == err {
		return ""
	}
	return err.Message
}

func (err *CommError) Error() string {
	if nil == err {
		return ""
	}
	return fmt.Sprintf("%d:%s", err.Code, err.Message)
}

func NewNotice(message string) *CommError {
	return NewCommError(0, message)
}
