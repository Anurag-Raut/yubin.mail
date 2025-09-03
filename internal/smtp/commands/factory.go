package command

import (
	"github.com/Yubin-email/smtp-server/logger"
	"github.com/Yubin-email/smtp-server/parser"
)

type CommandInterface interface {
	GetCommandType() CommandToken
	ParseCommand() error
	ProcessCommand(ctx *CommandContext, replyChannel chan ReplyInterface)
}

type Command struct {
	commandToken CommandToken
	parser       *parser.Parser
}

func (cmd *Command) GetCommandType() CommandToken {
	return cmd.commandToken
}

func NewCommand(commandString string, p *parser.Parser) CommandInterface {
	logger.Println("command string", commandString)
	switch commandString {
	case "EHLO":
		return &EHLO_CMD{Command{commandToken: EHLO, parser: p}}
	case "MAIL":
		return &MAIL_CMD{Command{commandToken: MAIL, parser: p}}
	case "RCPT":
		return &RCPT_CMD{Command{commandToken: RCPT, parser: p}}
	case "DATA":
		return &DATA_CMD{Command{commandToken: DATA, parser: p}}
	case "NOOP":
		return &NOOP_CMD{Command{commandToken: NOOP, parser: p}}
	case "RSET":
		return &RESET_CMD{Command{commandToken: RSET, parser: p}}
	case "QUIT":
		return &QUIT_CMD{Command{commandToken: QUIT, parser: p}}
	case "AUTH":
		return &AUTH_CMD{Command{commandToken: AUTH, parser: p}}
	case "STARTTLS":
		return &STARTTLS_CMD{Command{commandToken: STARTTLS, parser: p}}
	default:
		return nil
	}
}

func GetCommand(p *parser.Parser) (CommandInterface, error) {
	cmdToken, err := p.ParseCommandToken()
	if err != nil {
		return nil, err
	}
	cmdObj := NewCommand(cmdToken, p)
	err = cmdObj.ParseCommand()
	if err != nil {
		return nil, err
	}
	return cmdObj, nil
}
