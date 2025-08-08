package resp

type Response interface {
	GetStatus() int
	GetMessage() string
	GetTimestamp() int64
}
