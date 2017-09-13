// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

var addr = flag.String("addr", ":8080", "http service address")
var hh = newHubHandler()

func serveHub(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	id, _ := strconv.Atoi(r.URL.Path[len("/joinhub/"):])
	fmt.Println(id)
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	t, _ := template.ParseFiles("templates/hub.html")
	t.Execute(w, id)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/home.html")
}

func serveLogin(w http.ResponseWriter, r *http.Request) {
	m := make(map[int]string)
	for key, value := range hh.hubs {
		m[key] = value.name
	}
	if r.Method == "POST" {
		r.ParseForm()
	}
	t, _ := template.ParseFiles("templates/home.html")
	t.Execute(w, m)
}
func serveCreateHub(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		hh.NewHub(r.Form["name"][0])
	}
}

func serveWShandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Path[len("/ws/"):])
	fmt.Println("ws", id)
	serveWs(hh.hubs[id], w, r)
	//serveWs(hub, w, r)
}

func main() {
	flag.Parse()
	//hub := newHub("name")
	//go hub.run()
	hh.NewHub("test1")
	hh.NewHub("test2")
	for key, value := range hh.hubs {
		fmt.Println(key, *value, "\n")
	}
	go hh.hubs[1].run()
	go hh.hubs[0].run()
	http.HandleFunc("/", serveLogin)
	http.HandleFunc("/login", serveLogin)
	http.HandleFunc("/createhub/", serveCreateHub)
	http.HandleFunc("/joinhub/", serveHub)
	http.HandleFunc("/ws/", serveWShandler)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
