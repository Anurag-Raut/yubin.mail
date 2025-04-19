package main

func main() {
	cfg := NewConfig()
	cfg.SetAddr("127.0.0.1")
	cfg.SetPort("8000")
	clientServer := NewServer(cfg)
	clientServer.Listen()
}
