package parser

import (
	"unicode"

	"github.com/Yubin-email/internal/io/reader"
)

type Parser struct {
	reader *reader.Reader
}

func NewParser(reader *reader.Reader) *Parser {
	return &Parser{
		reader: reader,
	}
}

func (p *Parser) expectMultiple(tokens ...TokenType) (string, error) {
	for _, token := range tokens {
		value, err := p.expect(token)
		if err == nil {
			return value, nil
		}
	}
	return "", TokenNotFound{}
}

func (p *Parser) expect(token TokenType) (string, error) {
	switch token {
	case SPACE:
		{
			bytes, err := p.reader.Peek(1)
			if err != nil {
				return "", err
			}
			if string(bytes) != " " {
				return "", TokenNotFound{token: token}
			}
			_, err = p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			return string(bytes), nil

		}
	case LEFT_ANGLE_BRAC:
		{
			bytes, err := p.reader.Peek(1)
			if err != nil {
				return "", err
			}
			if string(bytes) != "<" {
				return "", TokenNotFound{token: token}
			}
			_, err = p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			return string(bytes), nil

		}
	case RIGHT_ANGLE_BRAC:
		{
			bytes, err := p.reader.Peek(1)
			if err != nil {
				return "", err
			}
			if string(bytes) != ">" {
				return "", TokenNotFound{token: token}
			}
			_, err = p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			return string(bytes), nil

		}
	case ALPHA:
		{
			bytes, err := p.reader.Peek(1)
			if err != nil {
				return "", err
			}

			if !unicode.IsLetter(rune(bytes[0])) {
				return "", TokenNotFound{token: token}
			}
			_, err = p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			return string(bytes), nil

		}

	case DIGIT:
		{
			b, err := p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			if b < '0' || b > '9' {
				return "", TokenNotFound{token: token}
			}
			return string(b), nil
		}

	case AT:
		{
			bytes, err := p.reader.Peek(1)
			if err != nil {
				return "", err
			}
			if string(bytes) != "@" {
				return "", TokenNotFound{token: token}
			}
			_, err = p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			return string(bytes), nil

		}
	case DOT:
		{
			bytes, err := p.reader.Peek(1)
			if err != nil {
				return "", err
			}
			if string(bytes) != "." {
				return "", TokenNotFound{token: token}
			}
			_, err = p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			return string(bytes), nil

		}

	case COLON:
		{
			bytes, err := p.reader.Peek(1)
			if err != nil {
				return "", err
			}
			if string(bytes) != ":" {
				return "", TokenNotFound{token: token}
			}
			_, err = p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			return string(bytes), nil

		}

	case ATEXT:
		{
			bytes, err := p.reader.Peek(1)
			if err != nil {
				return "", err
			}
			ch := bytes[0]
			if !((ch >= 'A' && ch <= 'Z') ||
				(ch >= 'a' && ch <= 'z') ||
				(ch >= '0' && ch <= '9') ||
				ch == '!' || ch == '#' || ch == '$' || ch == '%' ||
				ch == '&' || ch == '\'' || ch == '*' || ch == '+' ||
				ch == '-' || ch == '/' || ch == '=' || ch == '?' ||
				ch == '^' || ch == '_' || ch == '`' || ch == '{' ||
				ch == '|' || ch == '}' || ch == '~') {
				return "", TokenNotFound{token: token}
			}
			_, err = p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			return string(ch), nil
		}

	case EHLO_PARAM_CHAR:
		{
			b, err := p.reader.Peek(1)
			if err != nil {
				return "", err
			}
			ch := b[0]
			if !(ch >= 33 && ch <= 126) {
				return "", TokenNotFound{token: token}
			}
			_, err = p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			return string(ch), nil
		}
	case EHLO_GREET_CHAR:
		{
			b, err := p.reader.Peek(1)
			if err != nil {
				return "", err
			}
			ch := b[0]
			// valid: everything except CR (13) and LF (10)
			if ch == 10 || ch == 13 {
				return "", TokenNotFound{token: token}
			}
			_, err = p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			return string(ch), nil
		}

	case TEXTSTRING_CHAR:
		{
			b, err := p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			if !(b == '\t' || (b >= 32 && b <= 126)) {
				return "", TokenNotFound{token: token}
			}
			return string(b), nil
		}

	case CRLF:
		{
			bytes, err := p.reader.Peek(2)
			if err != nil {
				return "", err
			}
			if string(bytes) != "\r\n" {
				return "", TokenNotFound{token: token}
			}
			crlfBytes := make([]byte, 2)
			_, err = p.reader.Read(crlfBytes)
			if err != nil {
				return "", err
			}
			return string(crlfBytes), nil
		}

	}

	return "", TokenNotFound{}
}
