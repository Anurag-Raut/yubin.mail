package parser

import (
	"strings"

	"github.com/Yubin-email/internal/logger"
)

func (p *Parser) ParseAuth() (mechanism string, initialResponse *string, err error) {
	_, err = p.expect(SPACE)
	if err != nil {
		return mechanism, initialResponse, err
	}

	mechanism, err = p.ParseTextString(ATEXT)
	if err != nil {
		return mechanism, initialResponse, err
	}
	switch strings.ToUpper(mechanism) {
	case "PLAIN":
		{
			_, err := p.expect(SPACE)
			if err != nil {
				_, err := p.expect(CRLF)
				if err != nil {
					return mechanism, initialResponse, err
				}
			}

			logger.Println("IN PLAIN")
			initialResponseString, err := p.ParseTextString(ATEXT)
			if err != nil {
				return mechanism, initialResponse, err
			}
			initialResponse = &initialResponseString
			_, err = p.expect(CRLF)
			if err != nil {
				return mechanism, initialResponse, err
			}
		}
	}
	return mechanism, initialResponse, nil
}

func (p *Parser) ParseAuthResponse() (response string, err error) {
	response, err = p.expect(ATEXT)
	if err != nil {
		return "", err
	}
	_, err = p.expect(CRLF)
	if err != nil {
		return "", err
	}
	return response, nil
}
