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
	type bleh struct {
		Hubs     map[string]string
		Username string
	}
	var m bleh
	m.Username = ""
	m.Hubs = make(map[string]string)
	// Get names of active Hubs
	for key, value := range hh.hubs {
		m.Hubs[strconv.Itoa(key)] = value.name
	}
	if r.Method == "POST" {
		r.ParseForm()
		http.SetCookie(w, &http.Cookie{
			Name:  "Username",
			Value: r.Form["name"][0],
			Path:  "/",
		})
	}
	t, err := template.ParseFiles("templates/home.html")
	if err != nil {
		fmt.Println(err)
	}
	if userName, err := r.Cookie("Username"); err == nil {
		m.Username = userName.Value
	}
	fmt.Println(m.Username)
	t.Execute(w, m)
}
func serveCreateHub(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		hh.NewHub(r.Form["name"][0])
	}
	http.Redirect(w, r, "/", http.StatusPermanentRedirect)
}

func serveWShandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Path[len("/ws/"):])
	serveWs(hh.hubs[id], w, r)
}

func main() {
	flag.Parse()
	hh.NewHub("test1")
	hh.NewHub("test2")
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
