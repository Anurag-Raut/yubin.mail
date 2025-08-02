package parser

import (
	"fmt"
	"strconv"
)

type AuthReply struct {
	code               int
	message            string
	enhancedStatusCode *string
}

func (p *ReplyParser) ParseCRAMReply() (string, error) {
	code, err := p.expect(CODE)
	if err != nil {
		return "", err
	}
	if code != "334" {
		return "", fmt.Errorf("Expected code as 334 , got %s instead ", code)
	}
	_, err = p.expect(SPACE)
	if err != nil {
		return "", err
	}
	challenge, err := p.expect(TEXT_STRING)
	if err != nil {
		return "", err
	}
	_, err = p.expect(CRLF)
	if err != nil {
		return "", err
	}
	return challenge, nil
}

func (p *ReplyParser) ParseAuthReply(enhancedStatusCode bool) (*AuthReply, error) {
	authReplyObj := &AuthReply{}
	code, err := p.expect(CODE)

	if err != nil {
		return authReplyObj, err
	}
	authReplyObj.code, err = strconv.Atoi(code)
	if err != nil {
		return authReplyObj, err
	}

	enhancedStatusCodeString := ""
	// Loop to parse enhanced status code: e.g., 2.7.0
	if enhancedStatusCode {

		for i := range [3]int{} {
			d, err := p.expect(DIGIT)
			if err != nil {
				return authReplyObj, err
			}
			enhancedStatusCodeString += d
			if i < 2 {
				dot, err := p.expect(DOT)
				if err != nil {
					return authReplyObj, err
				}
				enhancedStatusCodeString += dot
			}
		}
	}
	authReplyObj.enhancedStatusCode = &enhancedStatusCodeString
	text, err := p.expect(TEXT_STRING)
	authReplyObj.message = text
	return authReplyObj, nil

}
