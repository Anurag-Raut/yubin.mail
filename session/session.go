package session

import (
	"net"
	"net/http"

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

	reply.GetReply(parser.Ehlo, p)
	logger.Println("Received EHLO reply")

	command.SendMail(s.writer, from)
	logger.Println("Sent MAIL FROM command")

	reply.GetReply(parser.ReplyLine, p)
	logger.Println("Received MAIL FROM reply")

	command.SendRcpt(s.writer, to[0])
	logger.Println("Sent RCPT TO command for", to[0])

	reply.GetReply(parser.ReplyLine, p)
	logger.Println("Received RCPT TO reply")

	command.SendBody(s.writer, p, "anurag@gmail.com")
	logger.Println("Sent message body")

	reply.GetReply(parser.ReplyLine, p)
	logger.Println("Received final response after message body")

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
