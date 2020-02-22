package discovery

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/yaien/discovery/pkg/network"
)

// Discovery is a p2p discovery helper for send and receive messages from multiple nodes
type Discovery interface {
	Start()
	Errors() chan error
	Messages() chan interface{}
	Ping()
}

type node struct {
	ID       string
	Address  string
	LastSeen time.Time
}

type response struct {
	Type string
	Data interface{}
}

type ping struct {
	ID      string
	Address string
}

type discovery struct {
	nw       network.Network
	nodes    map[string]*node
	errors   chan error
	messages chan interface{}
}

func (d *discovery) add(p *ping) {
	n, ok := d.nodes[p.ID]
	if ok {
		n.LastSeen = time.Now()
	} else {
		n = &node{
			Address:  p.Address,
			ID:       p.ID,
			LastSeen: time.Now(),
		}
		log.Println("Added", n.ID)
	}
	d.nodes[p.ID] = n
}

func (d *discovery) check() {
	for {
		time.Sleep(3 * time.Second)
		for _, n := range d.nodes {
			if n.LastSeen.Before(time.Now().Add(-5 * time.Second)) {
				log.Println("Deleted", n.ID)
				delete(d.nodes, n.ID)
			}
		}
	}
}

func (d *discovery) handle(r *response) {
	switch r.Type {
	case "ping":
		{
			var p ping

			if err := mapstructure.Decode(r.Data, &p); err != nil {
				d.errors <- err
				return
			}
			d.add(&p)
		}
	case "message":
		d.messages <- r.Data
	}
}

func (d *discovery) Start() {
	go d.nw.Start()
	go d.check()
	for {
		select {
		case message := <-d.nw.Messages():
			var r response
			if err := json.Unmarshal(message, &r); err != nil {
				d.errors <- err
				break
			}
			d.handle(&r)
		case err := <-d.nw.Errors():
			d.errors <- err
		}
	}
}

func (d *discovery) Ping() {
	me := &ping{
		ID:      uuid.New().String(),
		Address: d.nw.Address(),
	}
	msg := &response{Type: "ping", Data: me}
	data, _ := json.Marshal(msg)
	for {
		time.Sleep(3 * time.Second)
		err := d.nw.Broadcast(data)
		if err != nil {
			d.errors <- err
		}
	}
}

func (d *discovery) Errors() chan error {
	return d.errors
}

func (d *discovery) Messages() chan interface{} {
	return d.messages
}

// New returns a new discovery
func New(nw network.Network) Discovery {
	return &discovery{
		nw:       nw,
		errors:   make(chan error),
		messages: make(chan interface{}),
		nodes:    make(map[string]*node),
	}
}
