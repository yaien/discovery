package p2p

import (
	"fmt"
	"log"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/yaien/p2p/pkg/discovery"
)

// Requester is a helper for request a source to the p2p network
type Requester interface {
	Request(path string, data interface{}) (interface{}, error)
}

type requester struct {
	timeout time.Duration
	d       discovery.Discovery
}

type message struct {
	Path string
	Data interface{}
}

func (r *requester) start() {
	go r.d.Ping()
	r.d.Start()
}

// Request an external resource from another peer connected to the network
func (r *requester) Request(path string, data interface{}) (interface{}, error) {
	err := r.d.Request(&message{Path: path, Data: data})
	if err != nil {
		return nil, err
	}

	var response message

	msg := <-r.d.Messages()
	log.Println(msg)
	err = mapstructure.Decode(msg.Data, &response)
	if err != nil {
		return nil, err
	}
	if response.Path != path {
		return nil, fmt.Errorf("Invalid response path: '%s'", path)
	}
	return response.Data, nil

}

// NewRequester returns a new requester
func NewRequester(d discovery.Discovery) Requester {
	req := &requester{
		d:       d,
		timeout: 30 * time.Second,
	}
	go req.start()
	return req
}
