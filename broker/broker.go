package broker

type Broker interface {
	Pub(msg []byte) error
	Sub() chan []byte
	Close() error
}

type Producer interface {
	Pub(msg []byte) error
	Close() error
}

type Consumer interface {
	Sub() chan []byte
	Close() error
}
