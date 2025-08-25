package command

import (
	"fmt"

	"github.com/Yubin-email/smtp-client/config"
	"github.com/Yubin-email/smtp-client/dto/reply"
	"github.com/Yubin-email/smtp-client/io/writer"
	"github.com/Yubin-email/smtp-client/logger"
	"github.com/Yubin-email/smtp-client/parser"
	"github.com/google/uuid"
)

func SendEHLO(w *writer.Writer) error {
	return w.Fprintf("EHLO %s\r\n", config.ClientDomain)
}

func SendMail(w *writer.Writer, reversePath string) error {
	return w.Fprintf("MAIL FROM:<%s>\r\n", reversePath)
}

func SendRcpt(w *writer.Writer, forwardPath string) error {
	return w.Fprintf("RCPT TO:<%s>\r\n", forwardPath)
}

func SendBody(w *writer.Writer, p *parser.ReplyParser, body string, from string, to string) error {
	err := w.Fprintf("DATA\r\n")
	if err != nil {
		return err
	}

	_, err = reply.GetReply(parser.ReplyLine, p)
	if err != nil {
		logger.Println("ERRO in send body", err)
		return err
	}

	msgID := fmt.Sprintf("<%s@%s>", uuid.New().String(), config.ClientDomain)
	logger.Println("Message-ID:", msgID)

	if err = w.Fprintf("From: %s\r\n", from); err != nil {
		return err
	}

	if err = w.Fprintf("To: %s\r\n", to); err != nil {
		return err
	}

	if err = w.Fprintf("Subject: Test Subject\r\n"); err != nil {
		return err
	}

	if err = w.Fprintf("Message-ID: %s\r\n", msgID); err != nil {
		return err
	}

	logger.Println("BODY", body)
	if err = w.Fprintf("%s\r\n", body); err != nil {
		return err
	}

	if err = w.Fprintf(".\r\n"); err != nil {
		return err
	}

	return nil
}

func SendReset(w *writer.Writer) error {
	return w.Fprintf("RSET\r\n")
}

func SendVerify(w *writer.Writer, identifier string) error {
	return w.Fprintf("VRFY %s\r\f", identifier)
}

func SendExpand(w *writer.Writer, mailingList string) error {
	return w.Fprintf("EXPN %s\r\n", mailingList)
}

func SendHelp(w *writer.Writer, argument *string) error {
	if argument != nil {
		return w.Fprintf("HELP %s\r\n", *argument)
	}
	return w.Fprintf("HELP\r\n")
}

func SendNoop(w *writer.Writer) error {
	return w.Fprintf("NOOP\r\n")
}

func SendQuit(w *writer.Writer) error {
	return w.Fprintf("QUIT\r\f")
}
