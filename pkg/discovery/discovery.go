package discovery

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/yaien/p2p/pkg/network"
)

// Discovery is a p2p discovery helper for send and receive messages from multiple nodes
type Discovery interface {
	Start()
	Errors() chan error
	Messages() chan *Message
	Ping()
	Request(data interface{}) error
	Respond(data interface{}, address string) error
}

type node struct {
	ID       string
	Address  string
	LastSeen time.Time
}

type Message struct {
	Type    string
	Headers *headers
	Data    interface{}
}

type headers struct {
	ID      string
	Address string
}

type discovery struct {
	nw       network.Network
	nodes    map[string]*node
	errors   chan error
	messages chan *Message
	headers  *headers
}

func (d *discovery) add(p *headers) {
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
		for _, n := range d.nodes {
			if n.LastSeen.Before(time.Now().Add(-5 * time.Second)) {
				log.Println("Deleted", n.ID)
				delete(d.nodes, n.ID)
			}
		}
		time.Sleep(3 * time.Second)
	}
}

func (d *discovery) handle(msg *Message) {
	if msg.Headers.ID == d.headers.ID {
		return
	}
	switch msg.Type {
	case "ping":
		d.add(msg.Headers)
	case "message":
		d.messages <- msg

	}
}

func (d *discovery) master() (*node, error) {
	length := len(d.nodes)
	if length == 0 {
		return nil, errors.New("No peers connected")
	}
	index := rand.Intn(length)
	i := 0
	for _, n := range d.nodes {
		if i == index {
			return n, nil
		}
		i++
	}
	return nil, errors.New("No peers connected")
}

func (d *discovery) Start() {
	go d.nw.Start()
	go d.check()
	for {
		select {
		case message := <-d.nw.Messages():
			var r Message
			if err := json.Unmarshal(message, &r); err != nil {
				d.errors <- err
				continue
			}
			d.handle(&r)
		case err := <-d.nw.Errors():
			d.errors <- err
		}
	}
}

func (d *discovery) Request(data interface{}) error {
	res := Message{Type: "message", Headers: d.headers, Data: data}
	buff, err := json.Marshal(res)
	if err != nil {
		return err
	}
	//n, err := d.master()
	if err != nil {
		return err
	}

	return d.nw.Broadcast(buff)
}

func (d *discovery) Respond(data interface{}, address string) error {
	r := &Message{Type: "message", Headers: d.headers, Data: data}
	buff, err := json.Marshal(&r)
	if err != nil {
		return err
	}
	return d.nw.Broadcast(buff)
}

func (d *discovery) Ping() {
	msg := &Message{Type: "ping", Headers: d.headers}
	data, _ := json.Marshal(msg)
	for {
		err := d.nw.Broadcast(data)
		if err != nil {
			d.errors <- err
		}
		time.Sleep(1 * time.Second)
	}
}

func (d *discovery) Errors() chan error {
	return d.errors
}

func (d *discovery) Messages() chan *Message {
	return d.messages
}

// New returns a new discovery
func New(nw network.Network) Discovery {
	return &discovery{
		nw:       nw,
		errors:   make(chan error),
		messages: make(chan *Message),
		nodes:    make(map[string]*node),
		headers: &headers{
			ID:      uuid.New().String(),
			Address: nw.Address(),
		},
	}
}
