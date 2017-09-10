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

//hh := newHubHandler()

func serveHub(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	http.ServeFile(w, r, "templates/hub.html")
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/home.html")
}

func serveLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method)
	if r.Method == "POST" {
		r.ParseForm()
	}
	t, _ := template.ParseFiles("templates/home.html")
	t.Execute(w, hh)
}
func serveCreateHub(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		hh.NewHub(hh, r.Form["name"][0])
		//Hub := newHub(r.Form["name"][0])
		//hubs = append(hubs, Hub)
	}
}
func serveJoinHub(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Path[len("/joinhub/"):])
	serveWs(hh.hubs[id], w, r)
}

/*func serveWShandler(w http.ResponseWriter, r *http.Request) {
	serveWs(hub, w, r)
}
*/

func main() {
	flag.Parse()
	// hub := newHub()
	// go hub.run()
	http.HandleFunc("/", serveLogin)
	http.HandleFunc("/login", serveLogin)
	http.HandleFunc("/createhub", serveCreateHub)
	http.HandleFunc("/joinhub", serveJoinHub)
	//http.HandleFunc("/ws", serveWShandler)
	//http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
	//serveWs(hub, w, r)
	//})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
