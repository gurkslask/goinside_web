package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type pageInfo struct {
	Hubs     map[string]string
	Username string
}

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
	log.Println("I Loogin handler")
	var m pageInfo
	var u *User
	session, err := store.Get(r, "user")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m.Hubs = make(map[string]string)
	// Get names of active Hubs
	for key, value := range hh.hubs {
		m.Hubs[strconv.Itoa(key)] = value.name
	}
	if r.Method == "POST" {
		r.ParseForm()
		formname := r.Form["name"][0]
		formpassword := r.Form["password"][0]
		fmt.Println(r.Form)

		u = newUser()
		u.name = formname
		u.dbRead(db)
		if u.id == 0 {
			// User doesnt exist
			u.name = formname
			fmt.Println("User doesnt exist")
			u.dbAdd(db)
			u.dbRead(db)
			u.dbUpdatePassword(db, []byte(formpassword))
			session.Values["authenticated"] = true
			session.Values["uid"] = u.id
		} else {
			// User exist check password
			_, err := u.dbComparePassword(db, []byte(formpassword))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			session.Values["authenticated"] = true
			session.Values["uid"] = u.id

		}
		session.Values["user"] = u.name
	}
	if r.Method == "GET" {
		log.Println(" I Get metod")
		u = newUser()
		tmpint, ok := session.Values["uid"].(int)
		if ok == false {
			log.Println("uid in session is not int")
		}
		u.id = tmpint
		log.Printf("This is u.id in GET: %v", u.id)
		if u.id > 0 {
			log.Printf("Här läser vi från databasen")
			u.dbReadName(db)
			log.Printf("Det här användarnamnet fick vi: %v", u.name)
		}
	}
	err = session.Save(r, w)
	log.Println("Förbi POST")
	if err != nil {
		log.Printf("Could not save session: %v", err)
	}
	t, err := template.ParseFiles("templates/home.html")
	if err != nil {
		fmt.Println(err)
	}
	log.Println(u)
	if u != nil {
		m.Username = u.name
		fmt.Println(m.Username)
	}
	t.Execute(w, m)
}
func serveCreateHub(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		hh.NewHub(r.Form["name"][0])
	}
	http.Redirect(w, r, "/", 301)
}

func serveWShandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Path[len("/ws/"):])
	serveWs(hh.hubs[id], w, r)
}

func serveLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		session, err := store.Get(r, "user")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session.Values["authenticated"] = false
		session.Values["uid"] = 0
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "/", 301)
}
