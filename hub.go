package main

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// Registered connections.
	connections map[*WSConn]bool
	// Inbound messages from the connections.
	broadcast chan *PacketBatcher
	// Register requests from the connections.
	register chan *WSConn
	// Unregister requests from connections.
	unregister chan *WSConn
	// Close the goroutine
	quit chan bool
}

var h *hub = NewHub()

func NewHub() *hub {
	return &hub{
		broadcast:   make(chan *PacketBatcher),
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
