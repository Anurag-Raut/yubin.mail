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
	_, err := w.WriteString(fmt.Sprintf("EHLO %s\r\n", config.ClientDomain))
	return err
}

func SendMail(w *writer.Writer, reversePath string) error {
	_, err := w.WriteString(fmt.Sprintf("MAIL FROM:<%s>\r\n", reversePath))
	return err
}

func SendRcpt(w *writer.Writer, forwardPath string) error {
	_, err := w.WriteString(fmt.Sprintf("RCPT TO:<%s>\r\n", forwardPath))
	return err
}

func SendBody(w *writer.Writer, p *parser.ReplyParser, body string) error {
	// Send DATA command to the server
	_, err := w.WriteString(fmt.Sprintf("DATA\r\n"))
	if err != nil {
		return err
	}

	// Await server reply after sending DATA command
	_, err = reply.GetReply(parser.ReplyLine, p)
	if err != nil {
		return err
	}

	// Generate a unique Message-ID
	msgID := fmt.Sprintf("<%s@%s>", uuid.New().String(), config.ClientDomain)
	logger.Println("Message-ID:", msgID)

	// Send headers: From, To, Subject, and Message-ID
	_, err = w.WriteString(fmt.Sprintf("From: sender@example.com\r\n"))
	if err != nil {
		return err
	}

	_, err = w.WriteString(fmt.Sprintf("To: recipient@example.com\r\n"))
	if err != nil {
		return err
	}

	_, err = w.WriteString(fmt.Sprintf("Subject: Test Subject\r\n"))
	if err != nil {
		return err
	}

	_, err = w.WriteString(fmt.Sprintf("Message-ID: %s\r\n", msgID))
	if err != nil {
		return err
	}

	// Send the body of the email
	logger.Println("BODY", body)
	_, err = w.WriteString(fmt.Sprintf("%s\r\n", body))
	if err != nil {
		return err
	}

	// End the message with a single dot on a line by itself
	_, err = w.WriteString(fmt.Sprintf(".\r\n"))
	if err != nil {
		return err
	}

	return nil
}

func SendReset(w *writer.Writer) error {
	_, err := w.WriteString(fmt.Sprintf("RSET\r\n"))

	return err
}

func SendVerify(w *writer.Writer, identifier string) error {
	_, err := w.WriteString(fmt.Sprintf("VRFY %s\r\f", identifier))
	return err
}

func SendExpand(w *writer.Writer, mailingList string) error {
	_, err := w.WriteString(fmt.Sprintf("EXPN %s\r\n", mailingList))
	return err
}
func SendHelp(w *writer.Writer, argument *string) error {
	if argument != nil {
		_, err := w.WriteString(fmt.Sprintf("HELP %s\r\n", *argument))
		return err
	} else {
		_, err := w.WriteString(fmt.Sprintf("HELP\r\n"))
		return err
	}

}

func SendNoop(w *writer.Writer) error {

	_, err := w.WriteString(fmt.Sprintf("NOOP\r\n"))
	return err
}

func SendQuit(w *writer.Writer) error {
	_, err := w.WriteString(fmt.Sprintf("QUIT\r\f"))
	return err
}
