package udp

import (
	"context"
	"fmt"
	"net"

	"github.com/yaien/p2p/pkg/network"
)

// Config is the configuration of the UDP broadcast based network
type Config struct {
	Port      uint
	Broadcast string
	Address   string
}

type udp struct {
	config   *Config
	running  bool
	errors   chan error
	messages chan []byte
}

func (u *udp) address() string {
	return fmt.Sprintf("%s:%d", u.config.Address, u.config.Port)
}

func (u *udp) broadcast() string {
	return fmt.Sprintf("%s:%d", u.config.Broadcast, u.config.Port)
}

func (u *udp) connection() (net.PacketConn, error) {
	config := net.ListenConfig{Control: control}
	listener, err := config.ListenPacket(context.Background(), "udp", u.address())
	if err != nil {
		return nil, err
	}

	return listener, nil
}

func (u *udp) Address() string {
	return u.address()
}

func (u *udp) Start() {
	conn, err := u.connection()
	if err != nil {
		u.errors <- err
		return
	}
	defer conn.Close()
	u.running = true
	for {
		if !u.running {
			return
		}
		message := make([]byte, 1024)
		n, _, err := conn.ReadFrom(message)

		if err != nil {
			u.errors <- err
			continue
		}

		u.messages <- message[:n]
	}
}

func (u *udp) Broadcast(data []byte) error {
	return u.Send(data, u.broadcast())
}

func (u *udp) Send(data []byte, address string) error {
	conn, err := u.connection()
	if err != nil {
		return err
	}
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}
	_, err = conn.WriteTo(data, addr)
	return err
}

func (u *udp) Messages() chan []byte {
	return u.messages
}

func (u *udp) Errors() chan error {
	return u.errors
}

// NewNetwork returns a UPD network
func NewNetwork(config *Config) network.Network {
	return &udp{
		config:   config,
		errors:   make(chan error),
		messages: make(chan []byte),
	}
}

// Network returns a new udp network with the default configuration
func Network() network.Network {
	return NewNetwork(&Config{
		Port:      uint(1024),
		Address:   "0.0.0.0",
		Broadcast: "255.255.255.255",
	})
}
