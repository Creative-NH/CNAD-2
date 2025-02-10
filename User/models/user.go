package models

import (
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidLogin = errors.New("invalid login")
)

func RegisterUser(username, password string) error {

	cost := bcrypt.DefaultCost

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)

	if err != nil {
		return err
	}

	insertQuery := "INSERT INTO users (name, username, password, is_active) VALUES (?, ?, ?,?)"
	_, err = Db.Exec(insertQuery, "user", username, hash, 1)

	if err != nil {
		fmt.Printf("Error 501: %v", err)
		return err
	}

	return nil

}

func AuthenticateUser(username, password string) error {

	var hashPassword string
	query := "SELECT password FROM users WHERE username = ?"
	err := Db.QueryRow(query, username).Scan(&hashPassword)

	if err == sql.ErrNoRows {
		return ErrUserNotFound
	} else if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))

	if err != nil {
		return ErrInvalidLogin
	}

	return nil
}
