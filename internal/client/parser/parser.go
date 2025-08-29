package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/Yubin-email/smtp-client/io/reader"
	"github.com/Yubin-email/smtp-client/logger"
)

func GetDomainFromEmail(email string) (domain string, err error) {
	splits := strings.Split(email, "@")
	if len(splits) != 2 {
		return "", errors.New("Invalid Email , domain couldn't be parsed")
	}
	return splits[1], nil
}

type ReplyParser struct {
	reader *reader.Reader
}

func NewReplyParser(r *reader.Reader) *ReplyParser {
	return &ReplyParser{reader: r}
}

func (p *ReplyParser) ParseGreeting() (identifier string, textStrings []string, err error) {
	statusCodeString, err := p.expect(CODE)
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

func (p *ReplyParser) ParseEhloResponse() (replyCode int, domain string, ehlo_lines []string, err error) {
	replyCodeString, err := p.expect(CODE)
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
			_, greetErr := p.expect(EHLO_GREET) //TODO: ignoring greeting for now, later do something with it or return it in function
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
				_, greetErr := p.expect(EHLO_GREET)
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

func (p *ReplyParser) parseEhloMultiline(replyCode int) (ehlo_lines []string, err error) {
	logger.Println("Starting parseEhloMultiline with replyCode:", replyCode)
	for {
		logger.Println("Expecting CODE...")
		code, err := p.expect(CODE)
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

func (p *ReplyParser) parseEhloLine() (ehlo_line string, err error) {
	ehlo_line, err = p.expect(EHLO_KEYWORD)
	if err != nil {
		return ehlo_line, err
	}
	for {
		_, err := p.expect(SPACE)
		if err == nil {
			ehlo_param, err := p.expect(EHLO_PARAM)
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
func (p *ReplyParser) ParseReplyLine() (replyCode int, textStrings []string, err error) {

	for {
		codeString, err := p.expect(CODE)
		if err != nil {
			return replyCode, textStrings, err
		}

		_, err = strconv.Atoi(codeString)
		if err != nil {
			return replyCode, textStrings, err
		}

		_, err = p.expect(HYPHEN)
		if err == nil {
			txtString, err := p.parserTextString()
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
				txtString, err := p.parserTextString()
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

func (p *ReplyParser) parseSingleLine() (identifier string, textStrings []string, err error) {
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
		textString, err = p.parserTextString()
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

func (p *ReplyParser) parseMultiLineTextString() (identifier string, textStrings []string, err error) {

	identifier, err = p.parseDomain()

	if err != nil {
		identifier, err = p.parseAddressLiteral()
		if err != nil {
			return identifier, textStrings, err
		}

	}

	_, err = p.expect(SPACE)
	if err == nil {
		textString, err := p.parserTextString()
		if err != nil {
			return identifier, textStrings, err
		}
		textStrings = append(textStrings, textString)
	}
	for {
		codeString, err := p.expect(CODE)
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
			textStr, err := p.parserTextString()
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
		textString, err := p.parserTextString()
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
func (p *ReplyParser) parserTextString() (string, error) {

	textString, err := p.expect(TEXT_STRING)

	return textString, err
}

func (p *ReplyParser) parseDomain() (string, error) {
	subDomain, err := p.parseSubDomain()
	if err != nil {
		return "", err
	}
	for {
		_, err := p.expect(DOT)
		if err != nil {
			if (errors.Is(err, TokenNotFound{token: DOT})) {
				break
			} else {
				return "", err
			}

		}
		subDomain += "."
		newSubDomain, err := p.parseSubDomain()
		if err != nil {
			return "", err
		}
		subDomain += newSubDomain
	}
	return subDomain, nil
}

func (p *ReplyParser) parseSubDomain() (string, error) {
	firstVal, err := p.expectMultiple(ALPHA, DIGIT)
	if err != nil {
		return "", err
	}
	middleVal := ""
	for {
		ch, err := p.expectMultiple(ALPHA, DIGIT, HYPHEN)
		if err != nil {
			if (errors.Is(err, TokenNotFound{})) {
				break
			} else {
				return "", err
			}
		}
		middleVal += ch
	}
	// byh, _ := p.reader.Peek(1)
	// logger.ClientLogger.Println("PRITING BYTE AFTER in subdomain, ", string(byh))

	// if len(middleVal) > 0 {
	//
	// 	_, err = p.expectMultiple(ALPHA, DIGIT)
	//
	// 	if err != nil {
	// 		return firstVal + middleVal, err
	// 	}
	//
	// }
	return firstVal + middleVal, nil

}

func (p *ReplyParser) parseAddressLiteral() (string, error) {
	_, err := p.expect(LEFT_SQUARE_BRAC)
	if err != nil {
		return "", err
	}

	ip4addres, err := p.parseIPV4_AddressLiteral()
	if err != nil {
		return "", err
	}
	_, err = p.expect(RIGHT_SQUARE_BRAC)
	if err != nil {
		return "", err
	}
	return ip4addres, nil
}

func (p *ReplyParser) parseIPV4_AddressLiteral() (string, error) {
	ipv4_address := ""
	for i := 0; i < 3; i++ {
		val, err := p.expect(DIGIT)
		if err != nil {
			return "", err
		}
		ipv4_address += val
	}

	for j := 0; j < 3; j++ {
		dotString, err := p.expect(DOT)
		if err != nil {
			return "", err
		}
		ipv4_address += dotString
		for i := 0; i < 3; i++ {
			val, err := p.expect(DIGIT)
			if err != nil {
				return "", err
			}
			ipv4_address += val
		}
	}

	return ipv4_address, nil

}

func (p *ReplyParser) expectMultiple(tokens ...TokenType) (string, error) {
	for _, token := range tokens {
		value, err := p.expect(token)
		if err == nil {
			return value, nil
		}
	}
	return "", TokenNotFound{}

}

func (p *ReplyParser) expect(token TokenType) (string, error) {
	switch token {

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
			bytes, err := p.reader.Peek(1)
			if err != nil {
				return "", err
			}

			if !unicode.IsDigit(rune(bytes[0])) {
				return "", TokenNotFound{token: token}
			}
			_, err = p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			return string(bytes), nil

		}

	case TEXT_STRING:
		{
			var result []byte
			for {
				bytes, err := p.reader.Peek(1)
				if err != nil {
					return "", err
				}

				// logger.Println("Peeking byte:", bytes[0])

				if len(bytes) == 0 || !(bytes[0] == '\t' || (bytes[0] >= 32 && bytes[0] <= 126)) {
					logger.Println("Encountered invalid byte or end of input. Breaking the loop.", bytes)
					break
				}

				// logger.Println("Reading byte:", bytes[0])

				_, err = p.reader.ReadByte()
				if err != nil {
					return "", err
				}

				result = append(result, bytes[0])

				// logger.Println("Current result:", string(result))
			}

			logger.Println("Final textstring:", string(result))
			return string(result), nil
		}
	case EHLO_GREET:
		{
			var result []byte
			for {
				bytes, err := p.reader.Peek(1)
				if err != nil {
					return "", err
				}

				if len(bytes) == 0 || !((bytes[0] >= 0 && bytes[0] <= 9) || (bytes[0] >= 11 && bytes[0] <= 12) || (bytes[0] >= 14 && bytes[0] <= 127)) {
					logger.Println("Encountered invalid byte or end of input. Breaking the loop.")
					break
				}

				logger.Println("Reading byte:", bytes[0])

				_, err = p.reader.ReadByte()
				if err != nil {
					return "", err
				}

				result = append(result, bytes[0])

			}

			return string(result), nil
		}

	case EHLO_KEYWORD:
		{

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

	case EHLO_PARAM:
		{
			var ehlo_param string
			for {
				bytes, err := p.reader.Peek(1)
				if err != nil {
					return "", err
				}
				if len(bytes) == 0 || !(bytes[0] >= 33 && bytes[0] <= 126) {
					logger.Println("Encountered invalid byte or end of input. Breaking the loop.")
					break
				}
				ehlo_param += string(bytes)
				_, err = p.reader.ReadByte()
				if err != nil {
					logger.Println("Error reading byte:", err)
					return "", err
				}
			}
			return ehlo_param, nil
		}

	case CODE:
		{

			bytes, err := p.reader.Peek(3)
			logger.Println("CODE:", string(bytes))

			if err != nil {
				return "", nil
			}
			code := string(bytes)
			first := code[0]
			second := code[1]
			third := code[2]

			if first < 0x32 || first > 0x35 {
				return "", errors.New("Expected first value of code to be between 0x32 and 0x35")
			}

			if second < 0x30 || second > 0x35 {
				return "", errors.New("Expected second value of code to be between 0x30 and 0x35")
			}

			if third < 0x30 || third > 0x39 {
				return "", errors.New("Expected third value of code to be between 0x30 and 0x39")
			}
			byteArray := make([]byte, 3)
			_, err = p.reader.Read(byteArray)
			if err != nil {
				return "", err
			}
			return code, nil
		}
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

	case HYPHEN:
		{
			bytes, err := p.reader.Peek(1)
			if err != nil {
				return "", err
			}
			if string(bytes) != "-" {
				return "", TokenNotFound{token: token}
			}
			_, err = p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			return string(bytes), nil
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
	case LEFT_SQUARE_BRAC:
		{
			bytes, err := p.reader.Peek(1)
			if err != nil {
				return "", err
			}

			if string(bytes) != "[" {
				return "", TokenNotFound{token: token}
			}
			_, err = p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			return string(bytes), nil

		}
	case RIGHT_SQUARE_BRAC:
		{
			bytes, err := p.reader.Peek(1)
			if err != nil {
				return "", err
			}
			if string(bytes) != "]" {
				return "", TokenNotFound{token: token}
			}
			_, err = p.reader.ReadByte()
			if err != nil {
				return "", err
			}
			return string(bytes), nil

		}

	default:
		{
			return "", errors.New("No Matching token found " + strconv.Itoa(int(token)))
		}
	}

}
