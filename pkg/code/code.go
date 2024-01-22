package code

import (
	"fmt"
	"github.com/pkg/errors"
	"sync/atomic"
)

var (
	_messages atomic.Value
	_codes    = map[int]CodeC{}
)

// Codes is an interface for error code specification.
type Codes interface {
	Error() string
	Code() int
	Message() string
	Reason(reason error) Codes
}

type CodeC struct {
	code    int
	reason  string
	message string
}

func (e CodeC) Error() string {
	//return strconv.FormatInt(int64(e.code), 10)
	//return status.Errorf(codes.Code(e.code), "Invalid argument: ID cannot be zero").Error()
	//return status.Error(codes.Code(e.code), e.reason).Error()
	return e.reason
}

func (e CodeC) Code() int {
	return e.code
}

// Message return error message
func (e CodeC) Message() string {
	if cm, ok := _messages.Load().(map[int]CodeC); ok {
		if msg, ok := cm[e.Code()]; ok {
			return msg.message
		}
	}
	return e.Error()
}

func (e CodeC) Reason(reason error) Codes {
	if cm, ok := _messages.Load().(map[int]CodeC); ok {
		if msg, ok := cm[e.Code()]; ok {
			msg.reason = reason.Error()
			return msg
		}
	}
	return e
}

func add(e int, msg string) CodeC {
	if _, ok := _codes[e]; ok {
		panic(fmt.Sprintf("ecode: %d already exists", e))
	}
	_code := CodeC{
		code:    e,
		message: msg,
		reason:  msg,
	}
	_codes[e] = _code
	_messages.Store(_codes)
	return _code
}

func New(c int, msg string) Codes {
	if c <= 0 {
		panic("business ecode must be greater than zero")
	}
	return add(c, msg)
}

// Cause from error to code
func Cause(e error) Codes {
	if e == nil {
		return OK
	}
	ec, ok := errors.Cause(e).(Codes)
	if ok {
		return ec
	}
	return InternalServerError
}

// Code 通过错误码获取对应的 Codes 接口对象
func Code(code int) Codes {
	if cm, ok := _codes[code]; ok {
		return cm
	}

	// 如果找不到对应的错误码，返回一个默认的 InternalServerError
	return InternalServerError.Reason(fmt.Errorf("unknown error code: %d", code))
}
