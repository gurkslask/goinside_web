// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send    chan Message
	receive chan []byte
	id      int
	name    string
	db      dbClient
	// room     int
}

func newClient(hub *Hub, conn *websocket.Conn, name string) *Client {
	c := Client{
		hub:     hub,
		conn:    conn,
		send:    make(chan Message, 256),
		receive: make(chan []byte, 256),
		name:    name,
	}
	return &c
}

func newTestClient() *Client {
	c := Client{
		hub:     nil,
		conn:    nil,
		send:    make(chan Message, 256),
		receive: make(chan []byte, 256),
	}
	return &c
}

//Message that holds the client that sent it
type Message struct {
	client  *Client
	message []byte
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, text, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		// message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		var message Message
		message.client = c
		message.message = bytes.TrimSpace(bytes.Replace(text, newline, space, -1))
		c.hub.broadcast <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.receive:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.receive)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// Get username
	session, err := store.Get(r, "user")
	u := newUser()
	u.id = session.Values["uid"].(int)
	u.dbReadName(db)

	client := newClient(hub, conn, u.name)
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

//Changename of the player
func (c *Client) Changename(name string) {
	c.name = name
	c.hub.Whisper(c, "You have changed your name to: "+name)
}

func (c *Client) dbAdd(db *sql.DB) error {
	statement, err := db.Prepare("insert into dbClient(name) values(?)")
	if err != nil {
		return err
	}
	defer statement.Close()
	_, err = statement.Exec(c.name)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) dbDelete(db *sql.DB) error {
	stmt, err := db.Prepare("delete from dbClient where id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(c.id)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) dbUpdateName(db *sql.DB, newName string) error {
	stmt, err := db.Prepare("update dbClient set name=? where id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	fmt.Println(newName, c.id)
	_, err = stmt.Exec(newName, c.id)
	if err != nil {
		return err
	}
	return nil
}
func (c *Client) dbUpdatePassword(db *sql.DB, newPassword []byte) error {
	stmt, err := db.Prepare("update dbClient set password=? where id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	hashed, err := bcrypt.GenerateFromPassword(newPassword, 0)
	fmt.Println(string(newPassword), hashed)
	_, err = stmt.Exec(hashed, c.id)
	if err != nil {
		return err
	}
	return nil
}
func (c *Client) dbComparePassword(db *sql.DB, Password []byte) (bool, error) {
	var pwHash string
	stmt, err := db.Prepare("select password from dbClient where id = ?")
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	row := stmt.QueryRow(c.id)
	if err != nil {
		return false, err
	}
	err = row.Scan(&pwHash)
	if err != nil {
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(pwHash), Password)
	if err != nil {
		return false, err
	}
	return true, nil
}
func (c *Client) dbRead(db *sql.DB) error {
	var id int
	var name string
	stmt, err := db.Prepare("select id, name from dbClient where name = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	rows, err := stmt.Query(c.name)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&id, &name)
		if err != nil {
			return err
		}
	}
	c.id = id
	c.name = name
	return nil
}
func (c *Client) dbReadAll(db *sql.DB) error {
	var id int
	var name string
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("select id, name from dbClient")
	if err != nil {
		return err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&id, &name)
		if err != nil {
			return err
		}
		fmt.Printf("%v: %v \n", id, name)
	}
	return nil
}
