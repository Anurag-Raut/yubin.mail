package session

import (
	"github.com/Yubin-email/smtp-server/dto/command"
	"github.com/Yubin-email/smtp-server/dto/reply"
	"github.com/Yubin-email/smtp-server/io/reader"
	"github.com/Yubin-email/smtp-server/io/writer"
	"github.com/Yubin-email/smtp-server/parser"
	"github.com/Yubin-email/smtp-server/state"
)

type Session struct {
	mailState *state.MailState
	writer    *writer.Writer
	reader    *reader.Reader
}

func NewSession(r *reader.Reader, w *writer.Writer) *Session {

	return &Session{
		mailState: &state.MailState{},
		reader:    r,
		writer:    w,
	}
}
func (s *Session) Begin() {
	reply.Greet(s.writer)
	p := parser.NewParser(s.reader)
	for {
		cmd, err := command.GetCommand(p)
		if err != nil {
			return
			// reply.HandleParseError(writer, cmd.GetCommandType(), err)
		}

		replyChannel := make(chan reply.ReplyInterface)
		go cmd.ProcessCommand(s.mailState, replyChannel)
		for {
			var responseReply reply.ReplyInterface
			responseReply, ok := <-replyChannel
			if !ok {
				break
			}

			err := responseReply.HandleSmtpReply(s.writer)
			if err != nil {
				break
			}
		}
	}

}
