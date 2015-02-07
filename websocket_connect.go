package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	//maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WSConn struct {
	send   chan *Packet
	wsConn *websocket.Conn
}

func NewWSConn(w http.ResponseWriter, r *http.Request) (*WSConn, error) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
	}
	wc := make(chan *Packet)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading ws connection.", err)
	}
	return &WSConn{
		wsConn: conn,
		send:   wc,
	}, err
}

func (ws *WSConn) write(mt int, payload []byte) error {
	ws.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
	return ws.wsConn.WriteMessage(mt, payload)
}

func (ws *WSConn) writeJson(payload *Packet) error {
	ws.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
	return ws.wsConn.WriteJSON(payload)
}

func (ws *WSConn) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		ws.wsConn.Close()
	}()
	for {
		select {
		case message := <-ws.send:
			if err := ws.writeJson(message); err != nil {
				return
			}
		case <-ticker.C:
			if err := ws.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
