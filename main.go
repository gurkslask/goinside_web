// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

var addr = flag.String("addr", ":8080", "http service address")
var dbInit = flag.Bool("initdb", false, "Initialize database")
var dbTest = flag.Bool("testdb", false, "Test the db queries")
var hh = newHubHandler()
var store = sessions.NewCookieStore([]byte(("ThisiSVerYSecret!")))
var db *sql.DB

func main() {
	//DB connection
	var err error
	db, err = openDatabase()
	if err != nil {
		log.Printf("Failed to open database: %v", err)
	}
	defer db.Close()

	//Parse flags
	flag.Parse()
	if *dbInit {
		dbInitf(db)
	}
	if *dbTest {
		dbTestf(db)
	}

	http.HandleFunc("/", serveLogin)
	http.HandleFunc("/login", serveLogin)
	http.HandleFunc("/createhub/", serveCreateHub)
	http.HandleFunc("/joinhub/", serveHub)
	http.HandleFunc("/logout", serveLogout)
	http.HandleFunc("/ws/", serveWShandler)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
