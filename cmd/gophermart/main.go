package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/oldcyber/ya-devops-diploma/internal/server"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := server.NewConfig()
	if err := cfg.InitFromEnv(); err != nil {
		log.Error(err)
		return
	}
	if err := cfg.InitFromServerFlags(); err != nil {
		log.Error(err)
		return
	}

	a := server.App{}
	a.Cfg = cfg
	a.Queue = make(chan int, 20)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		log.Info("Starting Get Order")
		a.GetOrders(cfg.AccrualSystemAddress)
		wg.Done()
	}()
	go func() {
		a.Initialize(cfg.DatabaseDSN)
		a.Run(cfg.Address)
		wg.Done()
	}()
	go func() {
		<-c
		log.Info("Shutdown server")
		os.Exit(1)
	}()
	wg.Wait()
}
