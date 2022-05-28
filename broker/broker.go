package broker

import "ws/msg"

type Broker interface {
	Pub(msg *msg.PushMsg) error
	Sub() chan []byte
	Close() error
}

type Producer interface {
	Pub(msg *msg.PushMsg) error
	Close() error
}

type Consumer interface {
	Sub() chan []byte
	Close() error
}
