package reply

import (
	"strconv"

	"github.com/Yubin-email/smtp-server/io/writer"
	"github.com/Yubin-email/smtp-server/logger"
)

var CLRF = "\r\n"

type ReplyInterface interface {
	format() string
	HandleSmtpReply(w *writer.Writer) error
}

type Reply struct {
	code uint16
	text []string
}

type GreetingReply struct {
	Reply
	domain string
}

func (r *Reply) format() string {
	replyString := strconv.Itoa(int(r.code))
	if r.text != nil && len(r.text) > 0 {

		replyString += " "
		//BUG: check this out later
		replyString += (r.text[0])
	}

	replyString += CLRF
	logger.Println("SENDING", replyString)
	return replyString

}

func (r *GreetingReply) format() string {

	replyString := strconv.Itoa(int(r.code))
	replyString += " "
	replyString += r.domain
	if r.text != nil {

		replyString += " "
		//BUG: check this out later
		replyString += (r.text[0])
	}

	replyString += CLRF
	return (replyString)
}

func Greet(w *writer.Writer) error {
	text := []string{"Anurag Server"}
	rp := GreetingReply{
		Reply: Reply{
			code: 220,
			text: text,
		},
		domain: "gmail.com",
	}
	_, err := w.WriteString(rp.format())
	if err != nil {
		return err
	}
	return w.Flush()
}

func NewReply(code uint16, textlines ...string) ReplyInterface {
	return &Reply{
		code: code,
		text: textlines,
	}
}

func (r *Reply) HandleSmtpReply(w *writer.Writer) error {
	_, err := w.WriteString(r.format())
	return err
}

func (r *GreetingReply) HandleSmtpReply(w *writer.Writer) error {
	_, err := w.WriteString(r.format())
	return err
}
