package main

import (
	"log"

	"github.com/yaien/discovery/pkg/discovery"
	"github.com/yaien/discovery/pkg/network/udp"
)

func main() {
	nw := udp.Network()
	d := discovery.New(nw)
	go d.Start()
	for {
		select {
		case message := <-d.Messages():
			log.Println("Message:", message)
		case err := <-d.Errors():
			log.Println("Error: ", err)
		}
	}
}
