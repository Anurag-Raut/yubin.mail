package auth

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/Yubin-email/smtp-client/io/writer"
	"github.com/Yubin-email/smtp-client/logger"
	"github.com/Yubin-email/smtp-client/parser"
)

const email = ""
const username = "user@example.com"
const password = "secret"
const accessToken = ""

func IsError(r *parser.AuthReply) error {
	if r.Code >= 400 && r.Code <= 599 {
		return errors.New(r.Message)
	}
	return nil
}

func HandleAuth(mechanismsString string, enhancedStatusCode bool, w *writer.Writer, p *parser.ReplyParser) error {
	logger.Println("in handle auth")
	mechanisms := strings.Split(mechanismsString, " ")
	for _, mechanism := range mechanisms {
		mechanism = strings.ToUpper(mechanism)
		logger.Println("mechanism", mechanism)
		switch mechanism {
		case "PLAIN":
			{
				logger.Println("In plain")
				authStr := "\x00" + username + "\x00" + password
				w.Fprintf("AUTH PLAIN %s\r\n", base64.StdEncoding.EncodeToString([]byte(authStr)))
				authReply, err := p.ParseAuthReply(enhancedStatusCode)
				if err != nil {
					return err
				}
				err = IsError(authReply)
				if err != nil {
					continue
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
				authReply, err := p.ParseAuthReply(enhancedStatusCode)
				if err != nil {
					return err
				}
				err = IsError(authReply)
				if err != nil {
					continue
				}
				return nil

			}
		case "XOAUTH2":
			{
				xoauthStr := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", username, accessToken)
				xoauthEncoded := base64.StdEncoding.EncodeToString([]byte(xoauthStr))

				fmt.Fprintf(w, "AUTH XOAUTH2 %s\r\n", xoauthEncoded)
				authReply, err := p.ParseAuthReply(enhancedStatusCode)
				if err != nil {
					return err
				}
				err = IsError(authReply)
				if err != nil {
					continue
				}
				return nil
			}
		default:
			{
				continue
			}
		}
	}

	return errors.New("No auth mechanism couldn't handle it")
}

func HandleTLS(w *writer.Writer, p *parser.ReplyParser) error {
	w.WriteString("STARTTLS\r\n")
	_, err := p.ParseStartTLSReply()
	// tlsReply := &tlsReplyObj.(parser.StartTlsReply)

	return err
}
