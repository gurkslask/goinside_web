package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Game struct {
	answer              chan Message
	vote                chan Message
	answervote_active   bool
	impostorvote_active bool
	hub                 *Hub
	game_secret         string
	impostor            *Client
	master              *Client
	secret              chan string
	result              map[string]int
}

func Newgame(h *Hub) *Game {
	return &Game{
		answer:              make(chan Message),
		vote:                make(chan Message),
		answervote_active:   false,
		impostorvote_active: false,
		hub:                 h,
		secret:              make(chan string),
		result:              make(map[string]int),
	}
}

func (g *Game) Answervote() {
	result := make(map[string]int)
	channel := make(map[*Client]string)
	for {
		select {
		case data := <-g.answer:
			fmt.Printf("got %v", data.message)
			channel[data.client] = strings.ToLower(string(data.message))
		case <-time.After(10 * time.Second):
			for _, i := range channel {
				result[i]++
			}
			for key, value := range result {
				fmt.Println("%v: %v", key, value)
			}
			g.result = result
			return
		}
	}
	return
}

func (g *Game) rungame(h *Hub) {
	rand.Seed(time.Now().Unix())
	var slice_clients []*Client
	for key, _ := range h.clients {
		slice_clients = append(slice_clients, key)
	}
	h.game.master = slice_clients[rand.Intn(len(slice_clients))]
	h.Whisper(h.game.master, "You are the master, please choose a word to be secret word, you got 10 seconds, prefix with /secret")
	select {
	case h.game.game_secret = <-h.game.secret:
		h.Whisper(h.game.master, fmt.Sprintf("You chose %v", h.game.game_secret))
	case <-time.After(10 * time.Second):
		h.Announce("The master took too long :(")
		return
	}
	slice_clients = Deletefromslice(slice_clients, h.game.master)
	h.game.impostor = slice_clients[rand.Intn(len(slice_clients))]
	h.Whisper(h.game.impostor, fmt.Sprintf("You are impostor, the secret word is %v", h.game.game_secret))
	h.Announce("The game begins")
	<-time.After(10 * time.Second)
	h.game.answervote_active = true
	h.Announce("Game is over please vote with /answer")
	h.game.Answervote()
	//var resultmessage []string
	for key, value := range h.game.result {
		h.Announce(fmt.Sprintf("%v: %v \n", key, value))
	}
}
