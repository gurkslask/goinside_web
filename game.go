package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

//Game in a hub
type Game struct {
	answer             chan Message
	vote               chan Message
	answerVoteActive   bool
	impostorVoteActive bool
	hub                *Hub
	gameSecret         string
	impostor           *Client
	master             *Client
	secret             chan string
	result             map[string]int
}

//Newgame creates a newgame in hub
func Newgame(h *Hub) *Game {
	return &Game{
		answer:             make(chan Message),
		vote:               make(chan Message),
		answerVoteActive:   false,
		impostorVoteActive: false,
		hub:                h,
		secret:             make(chan string),
		result:             make(map[string]int),
	}
}

//Answervote runs at the end of a game
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
				fmt.Printf("%v: %v", key, value)
			}
			g.result = result
			return
		}
	}
}

func (g *Game) rungame(h *Hub) {
	rand.Seed(time.Now().Unix())
	var sliceClients []*Client
	for key := range h.clients {
		sliceClients = append(sliceClients, key)
	}
	h.game.master = sliceClients[rand.Intn(len(sliceClients))]
	h.Whisper(h.game.master, "You are the master, please choose a word to be secret word, you got 10 seconds, prefix with /secret")
	select {
	case h.game.gameSecret = <-h.game.secret:
		h.Whisper(h.game.master, fmt.Sprintf("You chose %v", h.game.gameSecret))
	case <-time.After(10 * time.Second):
		h.Announce("The master took too long :(")
		return
	}
	sliceClients = deletefromslice(sliceClients, h.game.master)
	h.game.impostor = sliceClients[rand.Intn(len(sliceClients))]
	h.Whisper(h.game.impostor, fmt.Sprintf("You are impostor, the secret word is %v", h.game.gameSecret))
	h.Announce("The game begins")
	<-time.After(10 * time.Second)
	h.game.answerVoteActive = true
	h.Announce("Game is over please vote with /answer")
	h.game.Answervote()
	//var resultmessage []string
	for key, value := range h.game.result {
		h.Announce(fmt.Sprintf("%v: %v \n", key, value))
	}
}
