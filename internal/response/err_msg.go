package resp

import (
	"time"
)

type ErrorMsg struct {
	Status    int    `json:"status"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`

	ErrorCodeEnum int `json:"error_code"`
}

func (e ErrorMsg) GetStatus() int {
	return e.Status
}

func (e ErrorMsg) GetMessage() string {
	return e.Message
}

func (e ErrorMsg) GetTimestamp() int64 {
	return e.Timestamp
}

type ErrorMsgBuilder struct {
	error ErrorMsg
}

func NewErrorMsgBuilder() *ErrorMsgBuilder {
	return &ErrorMsgBuilder{}
}

func (b *ErrorMsgBuilder) WithStatus(status int) *ErrorMsgBuilder {
	b.error.Status = status
	return b
}

func (b *ErrorMsgBuilder) WithCode(code int) *ErrorMsgBuilder {
	b.error.ErrorCodeEnum = code
	return b
}

func (b *ErrorMsgBuilder) WithMessage(e string) *ErrorMsgBuilder {
	b.error.Message = e
	return b
}

func (b *ErrorMsgBuilder) Build() ErrorMsg {
	b.error.Timestamp = time.Now().Unix()
	return b.error
}
