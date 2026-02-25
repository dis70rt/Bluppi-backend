package party

import "time"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func CaptureServerReceiveUs() int64 {
	return time.Now().UnixMicro()
}

func CaptureServerSendUs() int64 {
	return time.Now().UnixMicro()
}

