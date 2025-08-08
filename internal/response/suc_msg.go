package resp

import (
	"time"
)

type SuccessMsg struct {
	Status    int    `json:"status"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`

	Data any `json:"data"`
}

func (s SuccessMsg) GetStatus() int {
	return s.Status
}

func (s SuccessMsg) GetMessage() string {
	return s.Message
}

func (s SuccessMsg) GetTimestamp() int64 {
	return s.Timestamp
}

type SuccessMsgBuilder struct {
	success SuccessMsg
}

func NewSuccessMsgBuilder() *SuccessMsgBuilder {
	return &SuccessMsgBuilder{}
}

func (b *SuccessMsgBuilder) WithStatus(status int) *SuccessMsgBuilder {
	b.success.Status = status
	return b
}

func (b *SuccessMsgBuilder) WithMessage(message string) *SuccessMsgBuilder {
	b.success.Message = message
	return b
}

func (b *SuccessMsgBuilder) WithData(data any) *SuccessMsgBuilder {
	b.success.Data = data
	return b
}

func (b *SuccessMsgBuilder) Build() SuccessMsg {
	b.success.Timestamp = time.Now().Unix()
	return b.success
}
