package main

import (
	"database/sql"
	//"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

func main() {
	os.Remove("./data.db")

	db, err := sql.Open("sqlite3", "data.db")
	checkerr(err)
	defer db.Close()

	sqlStmt := `
	create table users  (id integer not null primary key, name text);
	delete from users;
	`
	_, err = db.Exec(sqlStmt)
	checkerr(err)

	tx, err := db.Begin()
	checkerr(err)
	stmt, err := tx.Prepare("insert into users(id, name) values(?, ?)")
	checkerr(err)
	defer stmt.Close()
	_, err = stmt.Exec(1, "Alex")
	tx.Commit()

}

func checkerr(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
