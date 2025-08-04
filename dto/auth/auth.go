package auth

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/Yubin-email/smtp-client/io/writer"
	"github.com/Yubin-email/smtp-client/parser"
)

const email = ""
const username = "your_username"
const password = "your_password"
const accessToken = ""

func HandleAuth(mechanismsString string, enhancedStatusCode bool, w *writer.Writer, p *parser.ReplyParser) error {
	mechanisms := strings.Split(mechanismsString, " ")
	for _, mechanism := range mechanisms {
		mechanism = strings.ToUpper(mechanism)
		switch mechanism {
		case "PLAIN":
			{
				authStr := "\x00" + username + "\x00" + password
				fmt.Fprintf(w, "AUTH PLAIN %s\r\n", base64.StdEncoding.EncodeToString([]byte(authStr)))
				_, err := p.ParseAuthReply(enhancedStatusCode)
				if err != nil {
					return err
				}
				return nil
			}
		case "CRAM-MD5":
			{
				fmt.Fprintf(w, "AUTH CRAM-MD5")
				challengeB64, err := p.ParseCRAMReply()
				if err != nil {
					return err
				}
				challenge, err := base64.StdEncoding.DecodeString(challengeB64)
				if err != nil {
					return err
				}

				//computing mac
				mac := hmac.New(md5.New, []byte("password"))
				mac.Write(challenge)
				digest := mac.Sum(nil)
				response := fmt.Sprintf("%s %x", username, digest)
				responseB64 := base64.StdEncoding.EncodeToString([]byte(response))

				fmt.Fprintf(w, "%s\r\n", responseB64)
				_, err = p.ParseAuthReply(enhancedStatusCode)
				if err != nil {
					return err
				}
				return nil
			}
		case "XOAUTH2":
			{
				xoauthStr := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", username, accessToken)
				xoauthEncoded := base64.StdEncoding.EncodeToString([]byte(xoauthStr))

				fmt.Fprintf(w, "AUTH XOAUTH2 %s\r\n", xoauthEncoded)
				_, err := p.ParseAuthReply(enhancedStatusCode)
				if err != nil {
					return err
				}
				return nil
			}
		default:
			{
				continue
			}
		}
	}

	return errors.New("Auth Mechanism handler Not found")
}

func HandleTLS(w *writer.Writer, p *parser.ReplyParser) error {
	fmt.Fprint(w, "STARTTLS")
	_, err := p.ParseStartTLSReply()
	// tlsReply := &tlsReplyObj.(parser.StartTlsReply)

	return err
}
