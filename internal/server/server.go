package server

import (
	"fmt"
)

type Server struct {
	opts ServerOptions
}

type ServerOptions struct {
	Port int `default:"8080" required:"true"`
}

func Run(opts ServerOptions) error {
	return fmt.Errorf("implement me")
}
