package server

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func (a *App) InitDB() {
	query, err := os.ReadFile("./sql/create_struct.sql")
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	if _, err := a.DB.Exec(string(query)); err != nil {
		log.Fatal(err)
		panic(err)
	}
}
