package command

type CommandToken string

const (
	EHLO     CommandToken = "EHLO"
	HELO     CommandToken = "HELO"
	MAIL     CommandToken = "MAIL"
	RCPT     CommandToken = "RCPT"
	DATA     CommandToken = "DATA"
	RSET     CommandToken = "RSET"
	VRFY     CommandToken = "VRFY"
	EXPN     CommandToken = "EXPN"
	HELP     CommandToken = "HELP"
	NOOP     CommandToken = "NOOP"
	QUIT     CommandToken = "QUIT"
	AUTH     CommandToken = "AUTH"
	STARTTLS CommandToken = "STARTTLS"
)

var AllCommands = []CommandToken{
	EHLO,
	HELO,
	MAIL,
	RCPT,
	DATA,
	RSET,
	VRFY,
	EXPN,
	HELP,
	NOOP,
	QUIT,
	AUTH,
	STARTTLS,
}

func IsValidCommand(cmd string) bool {
	for _, c := range AllCommands {
		if string(c) == cmd {
			return true
		}
	}
	return false
}
