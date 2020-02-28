package main

import (
	"log"
	"time"

	"github.com/yaien/p2p"
	"github.com/yaien/p2p/pkg/discovery"
	"github.com/yaien/p2p/pkg/network/udp"
)

func main() {
	nw := udp.Network()
	d := discovery.New(nw)
	requester := p2p.NewRequester(d)
	time.Sleep(3 * time.Second)
	res, err := requester.Request("greet", nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Response", res)
}
