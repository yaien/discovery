package main

import (
	"github.com/yaien/discovery/pkg/discovery"
	"github.com/yaien/discovery/pkg/network/udp"
)

func main() {
	nw := udp.Network()
	d := discovery.New(nw)
	d.Ping()
}
