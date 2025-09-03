package parser

import "errors"

func (p *Parser) parseDomain() (string, error) {
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

func (p *Parser) parseSubDomain() (string, error) {
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

func (p *Parser) parseAddressLiteral() (string, error) {
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

func (p *Parser) parseIPV4_AddressLiteral() (string, error) {
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
