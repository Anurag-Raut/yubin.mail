package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"fmt"

	"github.com/Yubin-email/smtp-server/dto/reply"
	"github.com/Yubin-email/smtp-server/parser"
)

type MECHANISM_COMMAND interface {
	ProcessCommand(*parser.Parser, chan reply.ReplyInterface) error
}

type PLAIN_AUTH_MECH struct {
	initialResponse *string
}

func (m *PLAIN_AUTH_MECH) ProcessCommand(p *parser.Parser, replyChannel chan reply.ReplyInterface) error {
	if m.initialResponse == nil || *m.initialResponse == "" {
		replyChannel <- reply.NewReply(334, "")
		res, err := p.ParseAuthResponse()
		if err != nil {
			replyChannel <- reply.NewReply(501, "5.5.2 Invalid auth response")
			return nil
		}
		m.initialResponse = &res
	}

	initialResBytesDecoded, err := base64.StdEncoding.DecodeString(*m.initialResponse)
	if err != nil {
		replyChannel <- reply.NewReply(501, "5.5.2 Cannot decode Base64 string")
		return nil
	}

	parts := bytes.Split(initialResBytesDecoded, []byte{0})
	if len(parts) != 3 {
		replyChannel <- reply.NewReply(501, "5.5.2 Invalid PLAIN auth format")
		return nil
	}

	authzID := string(parts[0])
	username := string(parts[1])
	password := string(parts[2])

	if username != "user@example.com" || password != "secret" {
		replyChannel <- reply.NewReply(535, "5.7.8 Authentication credentials invalid")
		return nil
	}

	replyChannel <- reply.NewReply(235, "2.7.0 Authentication successful")
	fmt.Println("Authenticated user:", username, "AuthZID:", authzID)

	return nil
}

type CRAM_MD5_AUTH_MECH struct {
	challenge string
}

func (m *CRAM_MD5_AUTH_MECH) ProcessCommand(p *parser.Parser, replyChannel chan reply.ReplyInterface) error {
	if m.challenge == "" {
		m.challenge = "<12345.67890@example.com>"
		replyChannel <- reply.NewReply(334, base64.StdEncoding.EncodeToString([]byte(m.challenge)))
	}

	resp, err := p.ParseAuthResponse()
	if err != nil {
		replyChannel <- reply.NewReply(501, "5.5.2 Invalid CRAM-MD5 response")
		return nil
	}

	decoded, err := base64.StdEncoding.DecodeString(resp)
	if err != nil {
		replyChannel <- reply.NewReply(501, "5.5.2 Cannot decode Base64 string")
		return nil
	}

	parts := bytes.SplitN(decoded, []byte(" "), 2)
	if len(parts) != 2 {
		replyChannel <- reply.NewReply(501, "5.5.2 Invalid CRAM-MD5 format")
		return nil
	}

	username := string(parts[0])
	digest := parts[1]

	expected := hmac.New(md5.New, []byte("secret"))
	expected.Write([]byte(m.challenge))
	expectedHex := fmt.Sprintf("%x", expected.Sum(nil))

	if !hmac.Equal(digest, []byte(expectedHex)) {
		replyChannel <- reply.NewReply(535, "5.7.8 Authentication credentials invalid")
		return nil
	}

	replyChannel <- reply.NewReply(235, "2.7.0 Authentication successful")
	fmt.Println("Authenticated user:", username)

	return nil
}

func HandleMechanism(mechanism string, initialResponse *string) (MECHANISM_COMMAND, error) {
	switch mechanism {
	case "PLAIN":
		return &PLAIN_AUTH_MECH{
			initialResponse: initialResponse,
		}, nil
	case "CRAM-MD5":
		return &CRAM_MD5_AUTH_MECH{}, nil
	}
	return nil, fmt.Errorf("unsupported auth mechanism: %q", mechanism)
}
