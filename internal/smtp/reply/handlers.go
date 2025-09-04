// handler.go
package reply

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Yubin-email/internal/io/writer"
	"github.com/Yubin-email/internal/logger"
	"github.com/Yubin-email/internal/parser"
)

const CRLF = "\r\n"

func (r *Reply) format() string {
	replyString := strconv.Itoa(int(r.code))
	if len(r.textStrings) > 0 {
		replyString += " " + r.textStrings[0]
	}
	replyString += CRLF
	logger.Println("SENDING", replyString)
	return replyString
}

func (r *GreetingReply) format() string {
	replyString := strconv.Itoa(int(r.code)) + " " + r.serverIdentifier
	if len(r.textStrings) > 0 {
		replyString += " " + r.textStrings[0]
	}
	replyString += CRLF
	return replyString
}

func (r *EhloReply) format() string {
	replyString := strconv.Itoa(int(r.code))
	if len(r.textStrings) > 0 {
		replyString += "-"
	} else {
		replyString += " "
	}
	replyString += r.domain + CRLF

	for i, textline := range r.textStrings {
		if i == len(r.textStrings)-1 {
			replyString += strconv.Itoa(int(r.code)) + " " + textline
		} else {
			replyString += strconv.Itoa(int(r.code)) + "-" + textline
		}
		replyString += CRLF
	}

	return replyString
}

func (r *Reply) HandleSmtpReply(w *writer.Writer) error {
	n, err := w.WriteString(r.format())
	logger.Println("WRITING THIS MANY BYTES", n, r.format())
	return err
}

func (r *GreetingReply) HandleSmtpReply(w *writer.Writer) error {
	_, err := w.WriteString(r.format())
	return err
}

func (r *EhloReply) HandleSmtpReply(w *writer.Writer) error {
	_, err := w.WriteString(r.format())
	return err
}

func Greet(w *writer.Writer) error {
	text := []string{"Anurag Server"}
	rp := GreetingReply{
		Reply: Reply{
			code:        220,
			textStrings: text,
		},
		serverIdentifier: "gmail.com",
	}
	_, err := w.WriteString(rp.format())
	if err != nil {
		return err
	}
	return w.Flush()
}

func (r *EhloReply) GetKey(key string) (isPresent bool, val *string) {
	for _, line := range r.textStrings {
		parts := strings.SplitN(line, " ", 2)
		keyInLine := parts[0]
		if len(parts) > 1 {
			val = &parts[1]
		}
		if keyInLine == key {
			return true, val
		}
	}
	return false, nil
}

func (r *GreetingReply) ParseReply() error {
	identifier, textStrings, err := r.parser.ParseGreeting()
	if err != nil {
		return err
	}
	r.code = 220
	r.textStrings = textStrings
	r.serverIdentifier = identifier
	return nil
}

func (r *EhloReply) ParseReply() error {
	replyCode, domain, textStrings, err := r.parser.ParseEhloResponse()
	if err != nil {
		return err
	}
	r.code = uint16(replyCode)
	r.textStrings = textStrings
	r.domain = domain
	return nil
}

func (r *Reply) ParseReply() error {
	replyCode, textStrings, err := r.parser.ParseReplyLine()
	if err != nil {
		return err
	}
	r.code = uint16(replyCode)
	r.textStrings = textStrings
	return nil
}

func (r *Reply) GetReplyCode() string {
	return strconv.Itoa(int(r.code))
}

func (r *Reply) Execute() error         { return nil }
func (r *GreetingReply) Execute() error { return nil }

func GetReply(token ReplyToken, p *parser.Parser) (ReplyInterface, error) {
	switch token {
	case REPLY_LINE:
		return &Reply{parser: p}, nil
	case GREETING:
		return &GreetingReply{Reply: Reply{parser: p}}, nil
	case EHLO:
		return &EhloReply{Reply: Reply{parser: p}}, nil
	default:
		return nil, errors.New("could not find the Reply")
	}
}

func NewEhloReply(code uint16, isTlHandshakeAlreadyDone bool) ReplyInterface {
	var textlines []string
	if len(ehloCfg.AllowedAuthMechanisms) > 0 {
		authLine := "AUTH"
		for _, mech := range ehloCfg.AllowedAuthMechanisms {
			authLine += fmt.Sprintf(" %s", mech)
		}
		textlines = append(textlines, authLine)
	}
	if ehloCfg.EnhancedStatusCode {
		textlines = append(textlines, "ENHANCEDSTATUSCODES")
	}
	logger.Println("start tls", ehloCfg.Starttls, "ISTLSHANDSHAKE", isTlHandshakeAlreadyDone)
	if ehloCfg.Starttls && !isTlHandshakeAlreadyDone {
		textlines = append(textlines, "STARTTLS")
	}
	return &EhloReply{
		Reply: Reply{
			code:        code,
			textStrings: textlines,
		},
		domain: ehloCfg.ServerName,
	}
}

func NewReply(code uint16, textlines ...string) ReplyInterface {
	return &Reply{
		code:        code,
		textStrings: textlines,
	}
}
