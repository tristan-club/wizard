package error

import (
	"fmt"
	"net/http"
)

const (
	ServerError   = 500
	BusinessError = 5001
	Success       = 200
)

type Error interface {
	Error() string
	ErrorType() int
	Code() int
	Msg() string
	HttpCode() int
}

type errorImpl struct {
	errorType int    // error 类型
	code      int    // code 响应状态码
	msg       string // 显式消息，该消息一般用于反馈给用户
	error     error  // 内涵错误，相对于显式消息，将给开发人员的报错信息写入该字段
	httpCode  int    // http状态码
}

func (e *errorImpl) Error() string {
	return e.error.Error()
}

func (e *errorImpl) Code() int {
	return e.code
}

func (e *errorImpl) Msg() string {
	return e.msg
}

func (e *errorImpl) ErrorType() int {
	return e.errorType
}

func (e *errorImpl) HttpCode() int {
	return e.httpCode
}

// NewServerError 新建一个非业务逻辑报错，msg统一为服务发生错误，将实际错误写入inner字段
func NewServerError(code int, msg string, err error) Error {
	if msg == "" {
		msg = CodeToMessage(code)
	}
	if msg == "" {
		msg = err.Error()
	}
	return NewError(code, msg, err, ServerError)
}

// NewBusinessError 新建一个业务逻辑报错，将业务逻辑异常写入到msg
func NewBusinessError(code int, msg string, err error) Error {

	if code == 0 {
		code = BusinessError
	}

	if msg == "" {
		msg = CodeToMessage(code)
	}

	if err == nil {
		err = fmt.Errorf(msg)
	}

	return NewError(code, msg, err, BusinessError)
}

func NewError(code int, msg string, err error, errType int) Error {
	e := &errorImpl{
		code:      code,
		error:     err,
		msg:       msg,
		errorType: errType,
	}
	if text := http.StatusText(code); text != "" {
		e.httpCode = code
	} else {
		e.httpCode = http.StatusOK
	}

	return e
}

func CodeToMessage(code int) string {
	if res, found := CodeMessageSting[code]; found {
		return res
	}

	if text := http.StatusText(code); text != "" {
		return text
	}

	return ""
}
