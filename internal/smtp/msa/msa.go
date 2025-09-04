package msa

import (
	"net"

	"github.com/Yubin-email/internal/session"
)

func HandleConn(conn net.Conn) {

	session := session.NewSession(conn, "msa")
	session.Begin(false)

}
