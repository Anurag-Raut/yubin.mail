package parser

import "fmt"

type TokenNotFound struct {
	token TokenType
}

func (t TokenNotFound) Error() string {
	return fmt.Sprintf("Token Not found:  %d ", t.token)
}
