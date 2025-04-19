package main

func main() {
	cServer := NewClientServer("127.0.0.1", "8000")
	cServer.Listen()
}
