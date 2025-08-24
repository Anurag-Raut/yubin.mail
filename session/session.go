package session

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/Yubin-email/smtp-server/dto/command"
	"github.com/Yubin-email/smtp-server/dto/reply"
	"github.com/Yubin-email/smtp-server/io/reader"
	"github.com/Yubin-email/smtp-server/io/writer"
	"github.com/Yubin-email/smtp-server/logger"
	"github.com/Yubin-email/smtp-server/parser"
	"github.com/Yubin-email/smtp-server/state"
)

type Session struct {
	mailState *state.MailState
	writer    *writer.Writer
	reader    *reader.Reader
	conn      net.Conn
}

func NewSession(conn net.Conn) *Session {
	r := reader.NewReader(conn)
	w := writer.NewWriter(conn)
	return &Session{
		mailState: &state.MailState{},
		reader:    r,
		writer:    w,
		conn:      conn,
	}
}

func (s *Session) upgradeToTLS() {
	cer, _ := tls.LoadX509KeyPair("server.crt", "server.key")
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	tlsConn := tls.Server(s.conn, config)
	s.conn = tlsConn
	err := tlsConn.Handshake()
	if err != nil {
		fmt.Println("TLS handshake failed:", err)
		return
	}
	s.mailState.ClearAll()
	s.Begin(true)
}

func (s *Session) Begin(isTls bool) {
	reply.Greet(s.writer)
	p := parser.NewParser(s.reader)
	cmdCtx := command.NewCommandContext(s.mailState, s.conn, isTls)

	for {
		logger.Println("Exptectin new command")
		cmd, err := command.GetCommand(p)
		logger.Println(cmd, "COMMANDcommand")
		if err != nil {
			return
			// reply.HandleParseError(writer, cmd.GetCommandType(), err)
		}

		replyChannel := make(chan reply.ReplyInterface)
		go cmd.ProcessCommand(cmdCtx, replyChannel)
		select {
		case responseReply := <-replyChannel:
			_ = responseReply.HandleSmtpReply(s.writer)
		case event := <-cmdCtx.EventChan:
			if event.Name == "TLS_UPGRADE" && !isTls {
				s.upgradeToTLS()
			}
		}
	}

}
