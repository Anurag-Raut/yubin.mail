package session

import (
	"io"
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

	err := command.SendEHLO(s.writer)
	if err != nil {
		logger.Println("Error sending EHLO command:", err)
		return
	}
	logger.Println("Sent EHLO command")
	s.readAvailableRaw()
	_, err = reply.GetReply(parser.Ehlo, p)
	if err != nil {
		logger.Println("Error receiving EHLO reply:", err)
		return
	}
	logger.Println("Received EHLO reply")

	err = command.SendMail(s.writer, from)
	if err != nil {
		logger.Println("Error sending MAIL FROM command:", err)
		return
	}
	logger.Println("Sent MAIL FROM command")

	_, err = reply.GetReply(parser.ReplyLine, p)
	if err != nil {
		logger.Println("Error receiving MAIL FROM reply:", err)
		return
	}
	logger.Println("Received MAIL FROM reply")

	err = command.SendRcpt(s.writer, to[0])
	if err != nil {
		logger.Println("Error sending RCPT TO command for ", to[0], ":", err)
		return
	}
	logger.Println("Sent RCPT TO command for ", to[0])

	_, err = reply.GetReply(parser.ReplyLine, p)
	if err != nil {
		logger.Println("Error receiving RCPT TO reply:", err)
		return
	}
	logger.Println("Received RCPT TO reply")

	err = command.SendBody(s.writer, p, "anurag@gmail.com")
	if err != nil {
		logger.Println("Error sending message body:", err)
		return
	}
	logger.Println("Sent message body")

	_, err = reply.GetReply(parser.ReplyLine, p)
	if err != nil {
		logger.Println("Error receiving final response after message body:", err)
		return
	}
	logger.Println("Received final response after message body")

	err = command.SendQuit(s.writer)
	if err != nil {
		logger.Println("Error sending QUIT command:", err)
		return
	}
	logger.Println("Sent QUIT command")

	_, err = reply.GetReply(parser.ReplyLine, p)
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
