package parser

import (
	"strings"
)

func (p *Parser) ParseAuth() (mechanism string, initialResponse *string, err error) {
	_, err = p.Expect(SPACE)
	if err != nil {
		return mechanism, initialResponse, err
	}

	mechanism, err = p.Expect(ATEXT)
	switch strings.ToUpper(mechanism) {
	case "PLAIN":
		{
			_, err := p.Expect(SPACE)
			if err == nil {
				_, err := p.Expect(CRLF)
				if err != nil {
					return mechanism, initialResponse, err
				}
			}
			initialResponseString, err := p.Expect(ATEXT)
			if err != nil {
				return mechanism, initialResponse, err
			}
			initialResponse = &initialResponseString

			_, err = p.Expect(CRLF)
			if err != nil {
				return mechanism, initialResponse, err
			}

		}
	}
	return mechanism, initialResponse, nil
}

func (p *Parser) ParseAuthResponse() (response string, err error) {
	response, err = p.Expect(ATEXT)
	if err != nil {
		return "", err
	}
	_, err = p.Expect(CRLF)
	if err != nil {
		return "", err
	}
	return response, nil
}
