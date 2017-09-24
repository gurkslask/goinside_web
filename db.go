package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func openDatabase() (*sql.DB, error) {
	//Open connection
	dbfile := "./data.db"
	os.Remove(dbfile)
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}
	return db, nil
}

func initDatabase(db *sql.DB) error {

	// Create table
	sqlStmt := `
	create table dbClientj (id integer not null primary key, name text);
	delete from dbClientj;
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}
	return nil
}

func examleInsertDatabase(db *sql.DB) error {
	// Insert data
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into dbClientj(id, name) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for i := 0; i < 100; i++ {
		_, err = stmt.Exec(i, fmt.Sprintf("dbClientj%d", i))
		if err != nil {
			log.Fatal(err)
		}
	}
	tx.Commit()
	return nil

}

func sqlAdddbClientj(db *sql.DB, dbClientj string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert into dbClientj(id, name) values(?, ?)")
	defer stmt.Close()
	_, err = stmt.Exec(sqlGetHighestID(db), dbClientj)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func sqlGetHighestID(db *sql.DB) int {
	rows, err := db.Query("select id, name from dbClient")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var tid int
	for rows.Next() {
		var id int
		tid = 0
		err = rows.Scan(&id, nil)
		fmt.Println(id)
		if err != nil {
			log.Fatal(err)
		}
		if id != tid+1 {
			break
		}
		tid = id
	}
	return tid + 1

}

type dbClient struct {
	id   int
	name string
}
