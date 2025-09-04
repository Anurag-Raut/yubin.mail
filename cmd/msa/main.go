package main

import (
	"net"
	"os"

	"github.com/Yubin-email/internal/logger"
	"github.com/Yubin-email/internal/smtp/msa"
	"github.com/Yubin-email/internal/store"
	"github.com/joho/godotenv"
)

type Server struct {
	port     string
	address  string
	listener net.Listener
	done     chan struct{}
}

type Config struct {
	port    *string
	address *string
}

func NewConfig() Config {
	return Config{}
}

func (c *Config) SetPort(port string) {
	c.port = &port
}

func (c *Config) SetAddr(address string) {
	c.address = &address
}

func NewServer(c Config) *Server {
	addr := "127.0.0.1"
	port := "587" // standard port for MSA
	if c.address != nil {
		addr = *c.address
	}
	if c.port != nil {
		port = *c.port
	}

	server := Server{
		port:    port,
		address: addr,
		done:    make(chan struct{}),
	}

	err := store.InitStore()
	if err != nil {
		panic(err)
	}
	return &server
}

func (s *Server) Listen() {
	logger.Println("MSA server listening on", s.address+":"+s.port)
	newListener, err := net.Listen("tcp", s.address+":"+s.port)
	s.listener = newListener
	if err != nil {
		logger.Println("Error: ", err.Error())
		return
	}

	for {
		c, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				logger.Println("MSA server shutting down")
				return
			default:
				logger.Println("Error", err.Error())
			}
		}
		go msa.HandleConn(c)
	}
}

func (s *Server) Close() {
	close(s.done)
	s.listener.Close()
	err := store.CloseStore()
	if err != nil {
		panic(err)
	}
}

func main() {
	env := os.Getenv("APP_ENV")
	envFile := "dev.env"
	if env == "prod" {
		envFile = ".env"
	}
	_ = godotenv.Load(envFile)

	cfg := NewConfig()
	cfg.SetAddr(os.Getenv("MSA_ADDRESS"))
	cfg.SetPort(os.Getenv("MSA_PORT"))

	msaServer := NewServer(cfg)
	msaServer.Listen()
}
