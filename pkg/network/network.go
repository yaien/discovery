package network

// Network an interface to start the listening discovery server process
type Network interface {
	Start()
	Address() string
	Send(data []byte, address string) error
	Broadcast(data []byte) error
	Messages() chan []byte
	Errors() chan error
}
