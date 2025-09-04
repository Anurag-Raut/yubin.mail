package parser

import (
	"errors"
	"strings"
)

func (p *Parser) ParseTextString(token TokenType) (s string, err error) {
	for {
		val, e := p.expect(token)
		if e != nil {
			var notFound TokenNotFound
			if errors.As(e, &notFound) {
				break
			}
			return s, e
		}
		s += val
	}
	return s, nil
}

func (p *Parser) ParseEHLO() (string, error) {
	_, err := p.expect(SPACE)
	if err != nil {
		return "", err
	}
	domain, err := p.parseDomain()

	if err == nil {
		_, err = p.expect(CRLF)
		if err != nil {
			return "", err
		}

		return domain, nil
	}
	addressLiteral, err := p.parseAddressLiteral()
	if err != nil {
		return "", err
	}
	_, err = p.expect(CRLF)
	if err != nil {
		return "", err
	}
	return addressLiteral, err
}

func (p *Parser) ParseMail() (string, error) {
	_, err := p.expect(SPACE)
	if err != nil {
		return "", err
	}
	fromString, err := p.reader.ReadStringOfLen(4)
	if err != nil {
		return "", err
	}
	if strings.ToLower(fromString) != "from" {
		return "", TokenNotFound{token: KEYWORD}
	}
	_, err = p.expect(COLON)
	if err != nil {
		return "", err
	}
	reversePath, err := p.parseReversePath()
	return reversePath, err
}

func (p *Parser) parseReversePath() (string, error) {
	path, err := p.parsePath()
	if err == nil {
		_, err = p.expect(CRLF)
		if err != nil {
			return "", err
		}

		return path, err
	}

	_, err = p.expect(LEFT_ANGLE_BRAC)
	if err != nil {
		return "", err
	}
	_, err = p.expect(LEFT_ANGLE_BRAC)
	if err != nil {
		return "", err
	}

	_, err = p.expect(CRLF)
	if err != nil {
		return "", err
	}
	return "", nil
}

func (p *Parser) parsePath() (string, error) {
	_, err := p.expect(LEFT_ANGLE_BRAC)
	if err != nil {
		return "", err
	}

	p.parseAD1() //ignore source routes

	mailbox, err := p.parseMailBox()
	if err != nil {
		return "", err
	}
	_, err = p.expect(RIGHT_ANGLE_BRAC)
	if err != nil {
		return "", err
	}

	return mailbox, nil

}

func (p *Parser) parseAD1() error {

	_, err := p.expect(AT)
	if err != nil {
		return nil
	}

	_, err = p.parseDomain()
	if err != nil {
		return err
	}
	for {
		_, err := p.expect(AT)
		if err != nil {
			break
		}

		_, err = p.parseDomain()
		if err != nil {
			return err
		}
	}

	_, err = p.expect(COLON)
	if err != nil {
		return err
	}

	return nil

}

func (p *Parser) parseMailBox() (string, error) {
	localPart, err := p.parseLocalPart()
	if err != nil {
		return "", err
	}
	_, err = p.expect(AT)
	if err != nil {
		return "", err
	}
	domain, err := p.parseDomain()
	if err == nil {

		return localPart + "@" + domain, nil
	}
	addressLiteral, err := p.parseAddressLiteral()
	if err != nil {
		return "", err
	}

	return localPart + "@" + addressLiteral, nil

}

func (p *Parser) parseLocalPart() (string, error) {
	dotString, err := p.parseDotString()
	if err == nil {
		return dotString, nil
	}
	quotedString, err := p.parseQuotedString()
	if err != nil {
		return "", err
	}

	return quotedString, nil
}

func (p *Parser) parseDotString() (string, error) {
	atom, err := p.parseAtom()
	if err != nil {
		return "", err
	}
	value := atom
	for {
		_, err := p.expect(DOT)
		if err != nil {
			break
		}

		atom, err := p.parseAtom()
		if err != nil {
			return "", err
		}
		value += atom + "." + atom
	}

	return value, nil
}

func (p *Parser) parseAtom() (string, error) {

	atom := ""
	ch, err := p.expect(ATEXT)
	if err != nil {
		return "", err
	}
	atom += ch
	for {
		ch, err := p.expectMultiple(ATEXT)
		if err != nil {
			break
		}
		atom += ch
	}
	return atom, nil
}
func (p *Parser) parseQuotedString() (string, error) {
	_, err := p.expect(DQUOTE)
	if err != nil {
		return "", err
	}
	value := ""

	for {
		ch, err := p.expect(QTEXTSMTP)
		if err != nil {
			value = ""
			break
		}
		value += ch
	}

	if value == "" {
		for {
			ch, err := p.expect(QPAIRSMTP)
			if err != nil {
				value = ""
				break
			}
			value += ch
		}

	}

	_, err = p.expect(DQUOTE)
	if err != nil {
		return "", err
	}

	return string('"') + value + string('"'), nil
}

func (p *Parser) ParseRCPT() (string, error) {
	_, err := p.expect(SPACE)
	if err != nil {
		return "", err
	}
	toString, err := p.reader.ReadStringOfLen(2)
	if err != nil {
		return "", err
	}
	if strings.ToLower(toString) != "to" {
		return "", errors.New("ERROR: Expected TO")
	}
	_, err = p.expect(COLON)
	if err != nil {
		return "", err
	}
	path, err := p.parsePath()
	if err != nil {
		return "", err
	}
	_, err = p.expect(CRLF)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (p *Parser) ParseData() error {
	_, err := p.expect(CRLF)
	return err
}

func (p *Parser) ParseDataLine() (line string, err error) {
	line, err = p.reader.GetLine("\r\n")
	return line, err
}

func (p *Parser) ParseQuit() (err error) {
	_, err = p.expect(CRLF)
	return err
}

func (p *Parser) ParseReset() error {
	_, err := p.expect(CRLF)
	return err
}

func (p *Parser) ParseNoop() error {
	_, err := p.expect(SPACE)
	if err == nil {
		_, err := p.expect(TEXT) //ignore the string paramater
		if err != nil {
			return err
		}

	} else {
		if (errors.As(err, &TokenNotFound{})) {
			_, err := p.expect(CRLF)
			return err

		} else {
			return err
		}
	}
	return nil
}

func (p *Parser) ParseStartTLS() error {
	_, err := p.expect(CRLF)
	return err
}

func (p *Parser) ParseCommandToken() (string, error) {
	return p.ParseTextString(ALPHA)
}
