package p2p

import (
	"fmt"
	"log"

	"github.com/mitchellh/mapstructure"
	"github.com/yaien/p2p/pkg/discovery"
)

type response struct {
	Path   string
	Status string
	Error  string
	Data   interface{}
}

// Handler receives a request data and returns a response data
type Handler = func(data interface{}) (interface{}, error)

// Responder is a router for handle incoming requests from external nodes
type Responder interface {
	On(path string, handler Handler)
	Start()
}

type responder struct {
	d        discovery.Discovery
	handlers map[string]Handler
}

func (r *responder) handle(request *message) *response {
	handler, ok := r.handlers[request.Path]
	if !ok {
		return &response{Path: request.Path, Status: "NOT_FOUND"}
	}
	data, err := handler(request.Data)
	if err != nil {
		return &response{Path: request.Path, Status: "ERROR", Error: err.Error()}
	}
	return &response{Path: request.Path, Status: "OK", Data: data}
}

func (r *responder) Start() {
	go r.d.Ping()
	go r.d.Start()
	for {
		select {
		case msg := <-r.d.Messages():
			var req message
			mapstructure.Decode(msg.Data, &req)
			res := r.handle(&req)
			fmt.Println(msg)
			err := r.d.Respond(res, msg.Headers.Address)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (r *responder) On(path string, handler Handler) {
	r.handlers[path] = handler
}

// NewResponder return a new responder
func NewResponder(d discovery.Discovery) Responder {
	res := &responder{
		d:        d,
		handlers: make(map[string]Handler),
	}
	return res
}
