package command

import (
	"github.com/Yubin-email/internal/logger"
	"github.com/Yubin-email/internal/parser"
	"github.com/Yubin-email/internal/smtp/reply"
)

type CommandInterface interface {
	GetCommandType() CommandToken
	ParseCommand() error
	ProcessCommand(ctx *CommandContext, replyChannel chan reply.ReplyInterface)
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
		return &EHLO_CMD{Command: Command{commandToken: EHLO, parser: p}}
	case "MAIL":
		return &MAIL_CMD{Command: Command{commandToken: MAIL, parser: p}}
	case "RCPT":
		return &RCPT_CMD{Command: Command{commandToken: RCPT, parser: p}}
	case "DATA":
		return &DATA_CMD{Command: Command{commandToken: DATA, parser: p}}
	case "NOOP":
		return &NOOP_CMD{Command: Command{commandToken: NOOP, parser: p}}
	case "RSET":
		return &RESET_CMD{Command: Command{commandToken: RSET, parser: p}}
	case "QUIT":
		return &QUIT_CMD{Command: Command{commandToken: QUIT, parser: p}}
	case "AUTH":
		return &AUTH_CMD{Command: Command{commandToken: AUTH, parser: p}}
	case "STARTTLS":
		return &STARTTLS_CMD{Command: Command{commandToken: STARTTLS, parser: p}}
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
