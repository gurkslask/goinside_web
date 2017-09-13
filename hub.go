// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strconv"
	"strings"
)

// hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Name of the hub
	name string
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	game *Game
}

func newHub(name string) *Hub {
	return &Hub{
		name:       name,
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case data := <-h.broadcast:
			msg := string(data.message)
			if msg[0] == '/' {
				command := strings.Split(msg, " ")
				switch command[0][1:] {
				case "name":
					if len(command) != 2 {
						break
					}
					data.client.Changename(command[1], h)
				case "answer":
					// If more or less than two words, skip it
					if len(command) != 2 {
						break
					}
					if h.game.answervote_active {
						var temp Message
						temp.client = data.client
						temp.message = []byte(command[1])
						h.game.answer <- temp
					}
				case "start":
					h.game = Newgame(h)
					go h.game.rungame(h)

				case "secret":
					if data.client == h.game.master {
						h.game.secret <- command[1]
					}
				}
				break

			}
			message := fmt.Sprintf("%v: %v", data.client.name, string(data.message))
			for client := range h.clients {
				select {
				case client.receive <- []byte(message):
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) Whisper(c *Client, msg string) {
	c.receive <- []byte(fmt.Sprintf("Announcer: %v", msg))
}

func (h *Hub) Announce(msg string) {

	msg = fmt.Sprintf("Announcer: %v \n", msg)
	for client := range h.clients {
		select {
		case client.receive <- []byte(msg):
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}
func Deletefromslice(slice []*Client, target *Client) []*Client {
	var index int
	for enum, value := range slice {
		if value == target {
			index = enum
		}
	}
	return append(slice[:index], slice[index+1:]...)
}

type hubHandler struct {
	hubs      map[int]*Hub
	roomcount int
}

func newHubHandler() *hubHandler {
	return &hubHandler{
		hubs:      make(map[int]*Hub),
		roomcount: 0,
	}
}
func (hh *hubHandler) NewHub(name string) {
	hh.hubs[hh.roomcount] = newHub(name)
	hh.roomcount++
}

func (hh hubHandler) String() string {
	s := ""
	for key, value := range hh.hubs {
		s = fmt.Sprintf("%v\n%v: %v \n", s, strconv.Itoa(key), value.name)
	}
	return s
}
