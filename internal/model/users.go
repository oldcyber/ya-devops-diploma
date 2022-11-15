package model

import (
	"database/sql"

	"github.com/oldcyber/ya-devops-diploma/internal/auth"
	log "github.com/sirupsen/logrus"
	// "github.com/oldcyber/ya-devops-diploma/internal/auth"
)

type User struct {
	UserID   int    `json:"user_id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

// AddUser adds new user to the database (POST /api/user/register)
func (u *User) AddUser(db *sql.DB, login, pas string) int {
	fRes := findUser(db, u.Login)
	switch {
	case fRes == Ok:
		_, err := db.Exec(`INSERT INTO users (login, password) VALUES ($1, crypt($2, gen_salt('bf')))
ON CONFLICT (login) DO UPDATE SET (login, password) = ($1, crypt($2, gen_salt('bf')))`, login, pas)
		if err != nil {
			log.Error(err)
			return InternalError
		}
	case fRes == AlienOrderNum:
		return AlienOrderNum
	}
	return Ok
}

// LoginUser checks user login and password (POST /api/user/login)
func (u *User) LoginUser(db *sql.DB, login, passwd string) (token string, status int) {
	var (
		userID int
		err    error
	)
	res := db.QueryRow("SELECT user_id FROM users WHERE login = $1 AND password = crypt($2, password)", login, passwd)
	err = res.Scan(&userID)
	if err != nil {
		log.Println("got Scan-error: " + err.Error())
		return "", InternalError
	}
	token, err = auth.CreateToken(userID)
	if err != nil {
		log.Println("got CreateToken=error: " + err.Error())
		return "", InternalError
	}
	return token, Ok
}
