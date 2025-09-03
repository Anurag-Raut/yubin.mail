package main

import (
	"net"
	"os"

	"github.com/Yubin-email/internal/logger"
	"github.com/Yubin-email/smtp-server/store"
	"github.com/joho/godotenv"
)

type Server struct {
	port     string
	adddress string
	listner  net.Listener
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
	port := "587"
	if c.address != nil {
		addr = *c.address
	}
	if c.port != nil {
		port = *c.port
	}
	server := Server{
		port:     port,
		adddress: addr,
		done:     make(chan struct{}),
	}
	err := store.InitStore()
	if err != nil {
		//TODO: improve this , you can create store object and handle the operation on that
		panic(err)
	}
	return &server
}

func (s *Server) Listen() {
	logger.Println("Listening on port", s.port)
	newListner, err := net.Listen("tcp", s.adddress+":"+s.port)
	s.listner = newListner
	if err != nil {
		logger.Println("Error: ", err.Error())
	}

	for {
		c, err := s.listner.Accept()
		if err != nil {
			select {
			case <-s.done:
				logger.Println("server is shutting down")
				return
			default:
				logger.Println("Error", err.Error())
			}
		}
		go handleConn(c)
	}
}

func (s *Server) Close() {
	close(s.done)
	s.listner.Close()
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
	} else {
		envFile = "dev.env"
	}
	godotenv.Load(envFile)
	cfg := NewConfig()

	cfg.SetAddr(os.Getenv("ADDRESS"))
	cfg.SetPort(os.Getenv("PORT"))
	clientServer := NewServer(cfg)
	clientServer.Listen()
}
