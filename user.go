package main

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// User is used for remembering the user and id
type User struct {
	id   int
	name string
	// room     int
}

func newUser() *User {
	u := User{
		id:   0,
		name: "",
	}
	return &u
}

func (u *User) dbAdd(db *sql.DB) error {
	statement, err := db.Prepare("insert into dbClient(name) values(?)")
	if err != nil {
		return err
	}
	defer statement.Close()
	_, err = statement.Exec(u.name)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) dbDelete(db *sql.DB) error {
	stmt, err := db.Prepare("delete from dbClient where id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(u.id)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) dbUpdateName(db *sql.DB, newName string) error {
	stmt, err := db.Prepare("update dbClient set name=? where id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	fmt.Println(newName, u.id)
	_, err = stmt.Exec(newName, u.id)
	if err != nil {
		return err
	}
	return nil
}
func (u *User) dbUpdatePassword(db *sql.DB, newPassword []byte) error {
	stmt, err := db.Prepare("update dbClient set password=? where id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	hashed, err := bcrypt.GenerateFromPassword(newPassword, 0)
	_, err = stmt.Exec(hashed, u.id)
	if err != nil {
		return err
	}
	return nil
}
func (u *User) dbComparePassword(db *sql.DB, Password []byte) (bool, error) {
	var pwHash string
	stmt, err := db.Prepare("select password from dbClient where id = ?")
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	row := stmt.QueryRow(u.id)
	if err != nil {
		return false, err
	}
	err = row.Scan(&pwHash)
	if err != nil {
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(pwHash), Password)
	if err != nil {
		return false, err
	}
	return true, nil
}
func (u *User) dbRead(db *sql.DB) error {
	var id int
	var name string
	stmt, err := db.Prepare("select id, name from dbClient where name = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	rows, err := stmt.Query(u.name)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&id, &name)
		if err != nil {
			return err
		}
	}
	u.id = id
	u.name = name
	return nil
}
func (u *User) dbReadAll(db *sql.DB) error {
	var id int
	var name string
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("select id, name from dbClient")
	if err != nil {
		return err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&id, &name)
		if err != nil {
			return err
		}
		fmt.Printf("%v: %v \n", id, name)
	}
	return nil
}
