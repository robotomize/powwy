package setup

import (
	"time"

	"github.com/robotomize/powwy/internal/quotes"
)

type ServerConfig struct {
	Addr                      string        `env:"ADDR,default=localhost:3333"`
	Network                   string        `env:"NETWORK,default=tcp"`
	GracefulConnCloseDeadline time.Duration `env:"GRACEFUL_CONN_CLOSE_DEADLINE,default=5s"`
}

type Config struct {
	Server ServerConfig
	Quotes quotes.Config
}
