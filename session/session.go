package session

import (
	"io"
	"net"
	"net/http"

	"github.com/Yubin-email/smtp-client/dto/auth"
	"github.com/Yubin-email/smtp-client/dto/command"
	"github.com/Yubin-email/smtp-client/dto/reply"
	"github.com/Yubin-email/smtp-client/io/reader"
	"github.com/Yubin-email/smtp-client/io/writer"
	"github.com/Yubin-email/smtp-client/logger"
	"github.com/Yubin-email/smtp-client/parser"
)

type Session struct {
	smtpConn   net.Conn
	reader     *reader.Reader
	writer     *writer.Writer
	httpWriter http.ResponseWriter
}

func NewSession(conn net.Conn, w http.ResponseWriter) *Session {
	return &Session{
		reader:     reader.NewReader(conn),
		writer:     writer.NewWriter(conn),
		httpWriter: w,
		smtpConn:   conn,
	}
}

func (s *Session) SendEmail(from string, to []string, body *string) {
	logger.Println("Starting SendEmail")
	logger.Println("From: ", from)
	logger.Println("To:", to)

	p := parser.NewReplyParser(s.reader)
	logger.Println("Initialized reply parser")

	command.SendEHLO(s.writer)
	logger.Println("Sent EHLO command")

	ehloReplyInterface, err := reply.GetReply(parser.Ehlo, p)
	logger.Println("Received EHLO reply")
	if err != nil {
		panic(err)
	}
	ehloReply, ok := ehloReplyInterface.(*reply.EhloReply)
	if !ok {
		panic("cannot convert to ehlo from ehlo reply")
	}
	present, val := ehloReply.GetKey("AUTH")
	if present {
		auth.HandleAuth(val, s.writer, p)
	}
	command.SendMail(s.writer, from)
	logger.Println("Sent MAIL FROM command")

	reply.GetReply(parser.ReplyLine, p)
	logger.Println("Received MAIL FROM reply")

	command.SendRcpt(s.writer, to[0])
	logger.Println("Sent RCPT TO command for ", to[0])

	reply.GetReply(parser.ReplyLine, p)
	logger.Println("Received RCPT TO reply")

	command.SendBody(s.writer, p, *body, from, to[0])
	logger.Println("Sent message body")

	reply.GetReply(parser.ReplyLine, p)
	logger.Println("Received final response after message body")
	command.SendQuit(s.writer)

	logger.Println("SendEmail finished")
}

func (s *Session) Begin() error {
	p := parser.NewReplyParser(s.reader)

	_, err := reply.GetReply(parser.Greeting, p)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) readAvailableRaw() {

	buf := make([]byte, 4096)
	n, err := s.reader.Read(buf)
	if err != nil && err != io.EOF {
		logger.Println("Error rading: ", err)
		return
	}
	if n > 0 {
		logger.Println("Raw read data: ", string(buf[:n]))
	} else {
		logger.Println("No data available to read.")
	}
}
