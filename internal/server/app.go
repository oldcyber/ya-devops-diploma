package server

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
	Cfg    *Config
	Queue  chan int
}

const (
	// compressLevel     = 5
	readHeaderTimeout = 3 * time.Second
)

func (a *App) Initialize(conn string) {
	var err error
	conn += "&connect_timeout=10"
	a.DB, err = sql.Open("postgres", conn)
	if err != nil {
		log.Fatal(err)
	}
	if err := a.DB.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Println("Postgres connected!")

	// Create Tables in DB
	log.Info("Create tables in DB")
	a.InitDB()
	a.Router = mux.NewRouter()
	a.appRouter()
}

func (a *App) Run(addr string) {
	c := cors.AllowAll()
	handler := c.Handler(a.Router)

	myServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	log.Info("Server started on: ", addr)
	err := myServer.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
