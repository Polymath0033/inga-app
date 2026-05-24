package queue

import (
	"errors"
)

var (
	ErrQueueClosed = errors.New("email queue is closed")
	ErrQueueFull   = errors.New("email queue is full, please retry later")
)
