// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages to the
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

	// Stop hub
	stop chan bool
}

func newHub(name string) *Hub {
	return &Hub{
		name:       name,
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		stop:       make(chan bool),
	}
}

func (h *Hub) run() {
	stopbool := false
	for !stopbool {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case <-h.stop:
			stopbool = true
		case data := <-h.broadcast:
			msg := string(data.message)
			if msg[0] == '/' {
				command := strings.Split(msg, " ")
				switch command[0][1:] {
				case "name":
					if len(command) != 2 {
						break
					}
					data.client.Changename(command[1])
				case "answer":
					// If more or less than two words, skip it
					if len(command) != 2 {
						break
					}
					if h.game.answerVoteActive {
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

// Whisper makes a message only for the selected client
func (h *Hub) Whisper(c *Client, msg string) {
	c.receive <- []byte(fmt.Sprintf("Announcer: %v", msg))
}

// Announce makes a message from a neutral announcer
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

// deletefromslice deletes target from slice and returns the modified slice
func deletefromslice(slice []*Client, target *Client) []*Client {
	var index int
	for enum, value := range slice {
		if value == target {
			index = enum
		}
	}
	return append(slice[:index], slice[index+1:]...)
}

// deletefromslice deletes target int from slice and returns the modified slice
func deleteFromintSlice(slice []int, target int) []int {
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
	i := &hubHandler{
		hubs:      make(map[int]*Hub),
		roomcount: 0,
	}
	go i.checkInactivity()
	return i
}
func (hh *hubHandler) NewHub(name string) {
	hh.hubs[hh.roomcount] = newHub(name)
	go hh.hubs[hh.roomcount].run()
	hh.roomcount++
}
func (hh *hubHandler) RemoveHub(id int) {
	hh.hubs[id].stop <- true
	delete(hh.hubs, id)
}

func (hh hubHandler) String() string {
	s := ""
	for key, value := range hh.hubs {
		s = fmt.Sprintf("%v\n%v: %v \n", s, strconv.Itoa(key), value.name)
	}
	return s
}

// Runs as a goroutine in hubhandler and closes hubs that arent used
func (hh *hubHandler) checkInactivity() {
	var idsToRemove []int
	for {
		for id := range idsToRemove {
			hh.RemoveHub(id)
			idsToRemove = deleteFromintSlice(idsToRemove, id)
		}
		for id, hub := range hh.hubs {
			if len(hub.clients) == 0 {
				idsToRemove = append(idsToRemove, id)
			}
		}
		<-time.After(time.Minute * 5)
	}
}
