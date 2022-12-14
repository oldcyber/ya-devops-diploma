package server

import (
	"flag"

	"github.com/caarlos0/env/v6"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Address              string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	DatabaseDSN          string `env:"DATABASE_URI"  envDefault:""`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"  envDefault:"http://localhost:9090"`
}

func (c *Config) GetAddress() string {
	return c.Address
}

func (c *Config) GetDatabaseDSN() string {
	return c.DatabaseDSN
}

func (c *Config) GetAccrualSystemAddress() string {
	return c.AccrualSystemAddress
}

func (c *Config) InitFromEnv() error {
	if err := env.Parse(c); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// func checkEnv(key string) bool {
//	_, ok := os.LookupEnv(key)
//	if !ok {
//		return false
//	} else {
//		return true
//	}
//}

func (c *Config) InitFromServerFlags() error {
	flag.StringVar(&c.Address, "a", c.Address, "address to listen on")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "database uri")
	flag.StringVar(&c.AccrualSystemAddress, "r", c.AccrualSystemAddress, "address of accrual system")
	flag.Parse()
	return nil
}

func NewConfig() *Config {
	return &Config{
		Address:              "localhost:8080",
		DatabaseDSN:          "postgresql://Sniff5090:qjn2hM6VJX8A95ZPcZ@localhost:5432/diploma?sslmode=disable",
		AccrualSystemAddress: "http://localhost:9090",
	}
}
