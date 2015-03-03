/*  OpenBCI golang server allows users to control, visualize and store data
    collected from the OpenBCI microcontroller.
    Copyright (C) 2015  Kevin Schiesser

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as
    published by the Free Software Foundation, either version 3 of the
    License, or (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 1 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Max message size allowed to be written to server over websocket.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096 * 16,
}

type WSConn struct {
	send   chan *Message
	wsConn *websocket.Conn
}

func NewWSConn(w http.ResponseWriter, r *http.Request) (*WSConn, error) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading ws connection.", err)
	}
	return &WSConn{
		wsConn: conn,
		send:   make(chan *Message, 32),
	}, err
}

func (ws *WSConn) write(mt int, payload []byte) error {
	ws.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
	return ws.wsConn.WriteMessage(mt, payload)
}

func (ws *WSConn) writeJson(payload *Message) error {
	ws.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
	return ws.wsConn.WriteJSON(payload)
}

func (ws *WSConn) WritePump(h *hub) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		h.unregister <- ws
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

//The brower is responsible for closing the websocket connection.
//To do so it will write a close conn message picked up by ReadPump.
func (ws *WSConn) ReadPump(h *hub) {
	defer func() {
		h.unregister <- ws
		ws.wsConn.Close()
	}()
	ws.wsConn.SetReadLimit(maxMessageSize)
	ws.wsConn.SetReadDeadline(time.Now().Add(pongWait))
	ws.wsConn.SetPongHandler(func(string) error { ws.wsConn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.wsConn.ReadMessage()
		if err != nil {
			break
		}
	}
}
