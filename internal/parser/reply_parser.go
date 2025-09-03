package parser

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Yubin-email/internal/logger"
)

func (p *Parser) parseCode() (string, error) {
	first, err := p.expect(DIGIT)
	if err != nil {
		return "", err
	}
	if first[0] < '2' || first[0] > '5' {
		return "", errors.New("invalid first digit in SMTP code")
	}

	second, err := p.expect(DIGIT)
	if err != nil {
		return "", err
	}
	if second[0] < '0' || second[0] > '5' {
		return "", errors.New("invalid second digit in SMTP code")
	}

	third, err := p.expect(DIGIT)
	if err != nil {
		return "", err
	}
	if third[0] < '0' || third[0] > '9' {
		return "", errors.New("invalid third digit in SMTP code")
	}

	return first + second + third, nil
}

func (p *Parser) ParseGreeting() (identifier string, textStrings []string, err error) {
	statusCodeString, err := p.parseCode()
	if err != nil {
		return identifier, textStrings, err
	}
	statusCode, err := strconv.Atoi(statusCodeString)
	if err != nil {
		return identifier, textStrings, errors.New("Error while converting the status code into int")
	}
	if statusCode != 220 {
		return identifier, textStrings, errors.New("Expected status Code as 220")
	}

	_, err = p.expect(SPACE)
	if err == nil {
		identifier, textString, err := p.parseSingleLine()
		return identifier, textString, err

	} else {
		_, err := p.expect(HYPHEN)
		if err != nil {
			return identifier, textStrings, err
		}
		return p.parseMultiLineTextString()

	}

}

func (p *Parser) ParseEhloResponse() (replyCode int, domain string, ehlo_lines []string, err error) {
	replyCodeString, err := p.parseCode()
	if err != nil {
		return replyCode, domain, ehlo_lines, err
	}

	replyCode, err = strconv.Atoi(replyCodeString)
	if err != nil {
		return replyCode, domain, ehlo_lines, err
	}
	_, err = p.expect(HYPHEN)
	if err == nil {
		// parse multi line reponse
		domain, err := p.parseDomain()
		if err != nil {
			return replyCode, domain, ehlo_lines, err
		}
		_, err = p.expect(SPACE)

		if err == nil {
			_, greetErr := p.parseEhloGreet() //TODO: ignoring greeting for now, later do something with it or return it in function
			if greetErr != nil {

				return replyCode, domain, ehlo_lines, err
			}
		}
		_, err = p.expect(CRLF)
		if err != nil {
			return replyCode, domain, ehlo_lines, err
		}

		ehlo_lines, err = p.parseEhloMultiline(replyCode)
		if err != nil {
			return replyCode, domain, ehlo_lines, err
		}

		return replyCode, domain, ehlo_lines, nil

	} else if (errors.As(err, &TokenNotFound{})) {

		_, err = p.expect(SPACE)

		if err == nil {
			//parse single line
			domain, err = p.parseDomain()

			if err != nil {
				return replyCode, domain, ehlo_lines, err
			}

			_, err = p.expect(SPACE)

			if err == nil {
				_, greetErr := p.parseEhloGreet()
				if greetErr != nil {

					return replyCode, domain, ehlo_lines, err
				}
			}

			_, err = p.expect(CRLF)
			return replyCode, domain, ehlo_lines, err
		} else {
			return replyCode, domain, ehlo_lines, err
		}

	} else {
		return replyCode, domain, ehlo_lines, err
	}

	return replyCode, domain, ehlo_lines, err
}

func (p *Parser) parseEhloMultiline(replyCode int) (ehlo_lines []string, err error) {
	logger.Println("Starting parseEhloMultiline with replyCode:", replyCode)
	for {
		logger.Println("Expecting CODE...")
		code, err := p.parseCode()
		if err != nil {
			logger.Println("Error while expecting CODE:", err)
			return ehlo_lines, err
		}
		logger.Println("Got CODE:", code)

		if code != strconv.Itoa(replyCode) {
			logger.Println("Unexpected CODE. Expected:", strconv.Itoa(replyCode), "Got:", code)
			return ehlo_lines, errors.New("EXPECTED REPLY CODE " + strconv.Itoa(replyCode))
		}

		logger.Println("Expecting HYPHEN...")
		_, err = p.expect(HYPHEN)
		if err == nil {
			logger.Println("HYPHEN found. Parsing EHLO line...")

			line, err := p.parseEhloLine()
			if err != nil {
				logger.Println("Error while parsing EHLO line:", err)
				return ehlo_lines, err
			}
			logger.Println("Parsed EHLO line:", line)
			ehlo_lines = append(ehlo_lines, line)

			logger.Println("Expecting CRLF after EHLO line...")
			_, err = p.expect(CRLF)
			if err != nil {
				logger.Println("Error while expecting CRLF:", err)
				return ehlo_lines, err
			}

		} else if errors.As(err, &TokenNotFound{}) {
			logger.Println("HYPHEN not found, trying to parse last EHLO line")

			_, err := p.expect(SPACE)
			if err != nil {
				logger.Println("Error while expecting SPACE:", err)
				return ehlo_lines, err
			}

			line, err := p.parseEhloLine()
			if err != nil {
				logger.Println("Error while parsing last EHLO line:", err)
				return ehlo_lines, err
			}
			logger.Println("Parsed last EHLO line:", line)
			ehlo_lines = append(ehlo_lines, line)

			logger.Println("Expecting CRLF after last EHLO line...")
			_, err = p.expect(CRLF)
			if err != nil {
				logger.Println("Error while expecting CRLF after last EHLO line:", err)
				return ehlo_lines, err
			}

			break
		} else {
			logger.Println("Unexpected error while expecting HYPHEN or SPACE:", err)
			return ehlo_lines, err
		}
	}

	logger.Println("Finished parsing EHLO lines:", ehlo_lines)
	return ehlo_lines, nil
}

func (p *Parser) parseEhloGreet() (string, error) {
	var result []byte
	for {
		val, err := p.expect(EHLO_GREET_CHAR)
		if err != nil {
			if _, ok := err.(TokenNotFound); ok && len(result) > 0 {
				return string(result), nil
			}
			return "", err
		}
		result = append(result, val[0])
	}
}

func (p *Parser) parseEhloParam() (string, error) {
	var result []byte
	for {
		val, err := p.expect(EHLO_PARAM_CHAR)
		if err != nil {
			if _, ok := err.(TokenNotFound); ok && len(result) > 0 {
				return string(result), nil
			}
			return "", err
		}
		result = append(result, val[0])
	}
}

func (p *Parser) parseEhloKeyword() (string, error) {
	ch, err := p.expectMultiple(ALPHA, DIGIT)
	if err != nil {
		return "", err
	}
	keyword := ch

	for {
		ch, err = p.expectMultiple(ALPHA, DIGIT, HYPHEN)
		if err != nil {
			if errors.As(err, &TokenNotFound{}) {
				break
			} else {
				return keyword, err
			}
		}
		keyword += ch
	}
	return keyword, nil
}

func (p *Parser) parseEhloLine() (ehlo_line string, err error) {
	ehlo_line, err = p.parseEhloKeyword()
	if err != nil {
		return ehlo_line, err
	}
	for {
		_, err := p.expect(SPACE)
		if err == nil {
			ehlo_param, err := p.parseEhloParam()
			if err != nil {
				return ehlo_line, err
			}

			ehlo_line += fmt.Sprintf(" %s", ehlo_param)
		} else if (errors.As(err, &TokenNotFound{})) {
			break
		} else {
			return ehlo_line, err
		}
	}
	return ehlo_line, nil
}
func (p *Parser) ParseReplyLine() (replyCode int, textStrings []string, err error) {

	for {
		codeString, err := p.parseCode()
		if err != nil {
			return replyCode, textStrings, err
		}

		_, err = strconv.Atoi(codeString)
		if err != nil {
			return replyCode, textStrings, err
		}

		_, err = p.expect(HYPHEN)
		if err == nil {
			txtString, err := p.parseTextString()
			if err != nil {
				return replyCode, textStrings, nil
			}
			textStrings = append(textStrings, txtString)
			_, err = p.expect(CRLF)
			if err != nil {
				return replyCode, textStrings, nil
			}
			_, err = p.expect(CRLF)
			if err != nil {
				return replyCode, textStrings, err
			}

		} else {
			_, err := p.expect(SPACE)
			if err == nil {
				txtString, err := p.parseTextString()
				if err != nil {
					return replyCode, textStrings, err
				}
				textStrings = append(textStrings, txtString)

			}
			_, err = p.expect(CRLF)
			if err != nil {
				return replyCode, textStrings, err
			}
			return replyCode, textStrings, nil
		}

	}

}

func (p *Parser) parseSingleLine() (identifier string, textStrings []string, err error) {
	identifier, err = p.parseDomain()
	if err != nil {
		identifier, err = p.parseAddressLiteral()
		if err != nil {
			return identifier, textStrings, err
		}

	}

	_, err = p.expect(SPACE)
	textString := ""
	if err == nil {
		textString, err = p.parseTextString()
		if err != nil {
			return identifier, textStrings, err
		}
	}
	_, err = p.expect(CRLF)

	if err != nil {
		return identifier, textStrings, err
	}

	return identifier, []string{textString}, nil

}
func (p *Parser) parseTextString() (string, error) {
	var out []byte
	for {
		val, err := p.expect(TEXTSTRING_CHAR)
		if err != nil {
			if _, ok := err.(TokenNotFound); ok && len(out) > 0 {
				return string(out), nil
			}
			return "", err
		}
		out = append(out, val[0])
	}
}

func (p *Parser) parseMultiLineTextString() (identifier string, textStrings []string, err error) {

	identifier, err = p.parseDomain()

	if err != nil {
		identifier, err = p.parseAddressLiteral()
		if err != nil {
			return identifier, textStrings, err
		}

	}

	_, err = p.expect(SPACE)
	if err == nil {
		textString, err := p.parseTextString()
		if err != nil {
			return identifier, textStrings, err
		}
		textStrings = append(textStrings, textString)
	}
	for {
		codeString, err := p.parseCode()
		if err != nil {
			return identifier, textStrings, err
		}
		code, err := strconv.Atoi(codeString)
		if err != nil {
			return identifier, textStrings, err
		}
		if code != 220 {
			return identifier, textStrings, errors.New("Could not parse Code")
		}

		_, err = p.expect(HYPHEN)
		if err == nil {
			textStr, err := p.parseTextString()
			if err != nil {
				return identifier, textStrings, err
			}
			textStrings = append(textStrings, textStr)
			_, err = p.expect(CRLF)
			if err != nil {
				return identifier, textStrings, err
			}
		} else if (errors.Is(err, TokenNotFound{})) {
			break
		} else {
			return identifier, textStrings, err
		}
	}
	_, err = p.expect(SPACE)

	if err == nil {
		textString, err := p.parseTextString()
		if err != nil {
			return identifier, textStrings, err
		}
		textStrings = append(textStrings, textString)
	}
	_, err = p.expect(CRLF)
	if err != nil {
		return identifier, textStrings, err

	}

	return identifier, textStrings, nil
}
