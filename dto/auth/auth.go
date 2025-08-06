package auth

import (
	"fmt"

	"github.com/Yubin-email/smtp-server/dto/reply"
	"github.com/Yubin-email/smtp-server/parser"
)

type MECHANISM_COMMAND interface {
	ProcessCommand(*parser.Parser, chan reply.ReplyInterface) error
}

type PLAIN_AUTH_MECH struct {
	input *string
}

func (m *PLAIN_AUTH_MECH) ProcessCommand(p *parser.Parser, replyChannel chan reply.ReplyInterface) error {
	return nil
}

func HandleMechanism(mechanism string, initialResponse *string) (MECHANISM_COMMAND, error) {
	switch mechanism {
	case "PLAIN":
		{

			return &PLAIN_AUTH_MECH{
				input: initialResponse,
			}, nil
		}
	}

	return nil, fmt.Errorf("unsupported auth mechanism: %q", mechanism)

}
