// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// See https://github.com/gorilla/websocket for details.

package main

import (
	"net/http"
)

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// Registered connections.
	connections map[*WSConn]bool
	// Inbound messages from the connections.
	broadcast chan *Message
	// Register requests from the connections.
	register chan *WSConn
	// Unregister requests from connections.
	unregister chan *WSConn
	// Close the goroutine
	quit chan bool
}

func NewHub() *hub {
	return &hub{
		broadcast:   make(chan *Message, 32),
		register:    make(chan *WSConn),
		unregister:  make(chan *WSConn),
		connections: make(map[*WSConn]bool),
		quit:        make(chan bool),
	}
}

func (h *hub) Close() {
	for c, _ := range h.connections {
		c.wsConn.Close()
		h.unregister <- c
	}
	h.quit <- true
	close(h.broadcast)
	close(h.register)
}

func (h *hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
		case m := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					close(c.send)
					delete(h.connections, c)
				}
			}
		case <-h.quit:
			return
		}
	}
}

func (h *hub) wsPacketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	wsConn, err := NewWSConn(w, r)
	if err != nil {
		http.Error(w, "Method not allowed", 405)
		return
	}
	h.register <- wsConn
	go wsConn.WritePump(h)
	go wsConn.ReadPump(h)
}
