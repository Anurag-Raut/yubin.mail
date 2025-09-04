package command

import (
	"github.com/Yubin-email/internal/smtp/auth"
	"github.com/Yubin-email/internal/smtp/reply"
	"github.com/Yubin-email/internal/state"
)

type CommandContext struct {
	mailState       *state.MailState
	EventChan       chan CommandEvent
	isTLS           bool
	isAuthenticated bool
}

type CommandEvent struct {
	Name string
	Data any
}

func NewCommandContext(mailState *state.MailState, isTLS bool) *CommandContext {
	return &CommandContext{
		mailState: mailState,
		isTLS:     isTLS,
		EventChan: make(chan CommandEvent),
	}
}

// EHLO
type EHLO_CMD struct {
	Command
	domain string
}

func (cmd *EHLO_CMD) ParseCommand() error {
	domain, err := cmd.parser.ParseEHLO()
	cmd.domain = domain
	return err
}

func (cmd *EHLO_CMD) ProcessCommand(ctx *CommandContext, replyChannel chan reply.ReplyInterface) {
	defer close(replyChannel)

	err := ctx.mailState.SetMailStep(state.EHLO)
	if err != nil {
		replyChannel <- reply.NewReply(503, err.Error())
		return
	}
	ctx.mailState.ClearAll()
	if !ctx.isTLS {
		replyChannel <- reply.NewEhloReply(250, false)
	} else {
		replyChannel <- reply.NewEhloReply(250, true)
	}
}

// MAIL
type MAIL_CMD struct {
	Command
	reversePath string
}

func (cmd *MAIL_CMD) ParseCommand() error {
	reversepath, err := cmd.parser.ParseMail()
	if err != nil {
		return err
	}
	cmd.reversePath = reversepath
	return nil
}
func (cmd *MAIL_CMD) ProcessCommand(ctx *CommandContext, replyChannel chan reply.ReplyInterface) {
	defer close(replyChannel)
	if !ctx.isAuthenticated {
		replyChannel <- reply.NewReply(530, "Authentication Required")
		return
	}
	err := ctx.mailState.SetMailStep(state.MAIL)
	if err != nil {
		replyChannel <- reply.NewReply(503, err.Error())
		return
	}
	ctx.mailState.ClearAll()
	ctx.mailState.AppendReversePatahBuffer([]byte(cmd.reversePath))
	replyChannel <- reply.NewReply(250)
}

// RCPT
type RCPT_CMD struct {
	Command
	forwardPath string
}

func (cmd *RCPT_CMD) ParseCommand() error {
	forwardPath, err := cmd.parser.ParseRCPT()
	if err != nil {
		return err
	}
	cmd.forwardPath = forwardPath
	return nil
}

func (cmd *RCPT_CMD) ProcessCommand(ctx *CommandContext, replyChannel chan reply.ReplyInterface) {
	defer close(replyChannel)
	err := ctx.mailState.SetMailStep(state.RCPT)
	if err != nil {
		replyChannel <- reply.NewReply(503, err.Error())
		return
	}
	ctx.mailState.AppendForwardPathBuffer([]byte(cmd.forwardPath))
	replyChannel <- reply.NewReply(250)
}

// DATA
type DATA_CMD struct {
	Command
	data string
}

func (cmd *DATA_CMD) ParseCommand() error {
	return cmd.parser.ParseData()
}

func (cmd *DATA_CMD) ProcessCommand(ctx *CommandContext, replyChannel chan reply.ReplyInterface) {
	defer close(replyChannel)
	err := ctx.mailState.SetMailStep(state.DATA)
	if err != nil {
		replyChannel <- reply.NewReply(503, err.Error())
		return
	}
	replyChannel <- reply.NewReply(354, "Start mail input; end with <CRLF>.<CRLF>")

	for {
		line, err := cmd.parser.ParseDataLine()
		if err != nil {
			replyChannel <- reply.NewReply(502, err.Error())
			return
		}
		if line == "." {
			break
		}
		ctx.mailState.AppendMailDataBuffer([]byte(line))
	}
	if err := ctx.mailState.StoreBuffer(); err != nil {
		replyChannel <- reply.NewReply(503, err.Error())
		return
	}
	ctx.mailState.ClearAll()
	replyChannel <- reply.NewReply(250)
}

// RESET
type RESET_CMD struct{ Command }

func (cmd *RESET_CMD) ParseCommand() error {
	return cmd.parser.ParseReset()
}
func (cmd *RESET_CMD) ProcessCommand(ctx *CommandContext, replyChannel chan reply.ReplyInterface) {
	defer close(replyChannel)
	ctx.mailState.ClearAll()
	ctx.mailState.SetMailStep(state.EHLO)
	replyChannel <- reply.NewReply(250, "OK")
}

// NOOP
type NOOP_CMD struct{ Command }

func (cmd *NOOP_CMD) ParseCommand() error {
	return cmd.parser.ParseNoop()
}
func (cmd *NOOP_CMD) ProcessCommand(ctx *CommandContext, replyChannel chan reply.ReplyInterface) {
	defer close(replyChannel)
	replyChannel <- reply.NewReply(250, "OK")
}

// QUIT
type QUIT_CMD struct{ Command }

func (cmd *QUIT_CMD) ParseCommand() error {
	return cmd.parser.ParseQuit()
}
func (cmd *QUIT_CMD) ProcessCommand(ctx *CommandContext, replyChannel chan reply.ReplyInterface) {
	defer close(replyChannel)
	ctx.mailState.SetMailStep(state.IDLE)
	ctx.mailState.ClearAll()
	replyChannel <- reply.NewReply(221, "Bye")
}

// AUTH
type AUTH_CMD struct {
	Command
	mechanism       string
	initialResponse *string
}

func (cmd *AUTH_CMD) ParseCommand() (err error) {
	cmd.mechanism, cmd.initialResponse, err = cmd.parser.ParseAuth()
	return err
}

func (cmd *AUTH_CMD) ProcessCommand(ctx *CommandContext, replyChannel chan reply.ReplyInterface) {
	defer close(replyChannel)
	mechObj, err := auth.HandleMechanism(cmd.mechanism, cmd.initialResponse)
	if err != nil {
		replyChannel <- reply.NewReply(503)
		return
	}
	if err := mechObj.ProcessCommand(cmd.parser, replyChannel); err != nil {
		ctx.mailState.ClearAll()
		return
	}
	ctx.isAuthenticated = true
}

// STARTTLS
type STARTTLS_CMD struct{ Command }

func (cmd *STARTTLS_CMD) ParseCommand() error {
	return cmd.parser.ParseStartTLS()
}
func (cmd *STARTTLS_CMD) ProcessCommand(ctx *CommandContext, replyChannel chan reply.ReplyInterface) {
	defer close(replyChannel)
	replyChannel <- reply.NewReply(220, "Ready to start TLS")
	ctx.EventChan <- CommandEvent{Name: "TLS_UPGRADE"}
}
