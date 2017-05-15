package main

import "github.com/gorilla/websocket"

type Client struct {
	Username string
	WS       *websocket.Conn
}

// ClientQ is a FIFO queue
type ClientQ struct {
	clients []*Client
	size    int
	head    int
	tail    int
	Count   int
}

// NewQueue returns a new queue with the given initial size.
func NewQueue(size int) *ClientQ {
	return &ClientQ{
		clients: make([]*Client, size),
		size:    size,
	}
}

// Push adds a node to the queue.
func (q *ClientQ) Add(client *Client) {
	if q.head == q.tail && q.Count > 0 {
		nodes := make([]*Client, len(q.clients)+q.size)
		copy(nodes, q.clients[q.head:])
		copy(nodes[len(q.clients)-q.head:], q.clients[:q.head])
		q.head = 0
		q.tail = len(q.clients)
		q.clients = nodes
	}
	q.clients[q.tail] = client
	q.tail = (q.tail + 1) % len(q.clients)
	q.Count++
}

// Pop removes and returns a node from the queue in first to last order.
func (q *ClientQ) Next() *Client {
	if q.Count == 0 {
		return nil
	}
	node := q.clients[q.head]
	q.head = (q.head + 1) % len(q.clients)
	q.Count--
	return node
}
