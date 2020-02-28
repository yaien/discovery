package main

import (
	"github.com/yaien/p2p"
	"github.com/yaien/p2p/pkg/discovery"
	"github.com/yaien/p2p/pkg/network/udp"
)

func main() {
	nw := udp.Network()
	d := discovery.New(nw)
	responder := p2p.NewResponder(d)
	responder.On("greet", func(data interface{}) (interface{}, error) {
		res := map[string]string{"greet": "hello world"}
		return res, nil
	})
	responder.Start()
}
