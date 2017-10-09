package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

func dbInitf(db *sql.DB) {
	log.Print("dbinit active")
	err := initDatabase(db)
	if err != nil {
		log.Fatalf("Could not create database %v", err)
	}
}

func dbTestf(db *sql.DB) {
	log.Print("dbtest active")
	c := newTestClient()
	c.name = "Alex"
	c2 := newTestClient()
	c2.name = "Alex2"
	fmt.Println("1")
	err := c.dbAdd(db)
	if err != nil {
		log.Fatalf("Could not create user row: %v", err)
	}
	fmt.Println("2")
	err = c2.dbAdd(db)
	if err != nil {
		log.Fatalf("Could not create user row: %v", err)
	}
	fmt.Println("3")
	err = c.dbRead(db)
	if err != nil {
		log.Fatalf("Could not read user row: %v", err)
	}
	err = c2.dbRead(db)
	if err != nil {
		log.Fatalf("Could not read user row: %v", err)
	}
	fmt.Println("4")
	err = c.dbDelete(db)
	if err != nil {
		log.Fatalf("Could not delete user row: %v", err)
	}
	err = c2.dbUpdateName(db, "Kalle")
	if err != nil {
		log.Fatalf("Could not update: %v", err)
	}
	err = c2.dbUpdatePassword(db, []byte("Mycket hemligt"))
	if err != nil {
		log.Fatalf("Could not update password: %v", err)
	}
	ok, err := c2.dbComparePassword(db, []byte("Mycket hemligt"))
	if err != nil {
		log.Fatalf("Could not Compare password: %v", err)
	}
	if ok {
		fmt.Println("Password ok!")
	}
	ok, err = c2.dbComparePassword(db, []byte("Inte så hemligt"))
	if err != nil {
		log.Fatalf("Could not Compare password: %v", err)
	}
	if ok {
		fmt.Println("Password ok!")
	}
	err = c.dbReadAll(db)
	if err != nil {
		log.Fatalf("Coudl not read all: %v", err)
	}
	os.Exit(2)
}
