package reply

// import (
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"runtime"
// 	"strconv"

// 	"github.com/Yubin-email/internal/io/writer"
// 	"github.com/Yubin-email/internal/logger"
// 	"gopkg.in/yaml.v3"
// )

// type ehlo_config struct {
// 	Starttls              bool     `yaml:"starttls"`
// 	AllowedAuthMechanisms []string `yaml:"allowedAuthMechanisms"`
// 	EnhancedStatusCode    bool     `yaml:"enhancedStatusCode"`
// 	ServerName            string   `yaml:"serverName"`
// }

// func getEhloConfig() ehlo_config {
// 	_, filename, _, _ := runtime.Caller(0)
// 	dir := filepath.Dir(filename)
// 	cfg := ehlo_config{}
// 	data, err := os.ReadFile(filepath.Join(dir, "ehlo_config.yaml"))
// 	if err != nil {
// 		panic("Ehlo config file not found")
// 	}
// 	err = yaml.Unmarshal(data, &cfg)
// 	if err != nil {
// 		panic("Errror while Unmarshalling ehlo yaml")
// 	}

// 	return cfg
// }

// var ehloCfg = getEhloConfig()

// type EhloReply struct {
// 	Reply
// 	domain string
// }

// func (r *EhloReply) format() string {
// 	replyString := strconv.Itoa(int(r.code))
// 	if len(r.textStrings) > 0 {
// 		replyString += "-"
// 	} else {
// 		replyString += " "
// 	}
// 	replyString += r.domain

// 	replyString += CRLF
// 	if len(r.textStrings) > 0 {

// 		for i, textline := range r.textStrings {
// 			if i == len(r.textStrings)-1 {

// 				replyString += fmt.Sprintf("%d %s", r.code, textline)
// 			} else {
// 				replyString += fmt.Sprintf("%d-%s", r.code, textline)
// 			}
// 			replyString += CRLF
// 		}
// 	}

// 	fmt.Println("REPL STRING", replyString)
// 	return (replyString)
// }
// func NewEhloReply(code uint16, isTlHandshakeAlreadyDone bool) ReplyInterface {
// 	var textlines []string
// 	if len(ehloCfg.AllowedAuthMechanisms) > 0 {
// 		authLine := "AUTH"

// 		for _, mech := range ehloCfg.AllowedAuthMechanisms {
// 			authLine += fmt.Sprintf(" %s", mech)
// 		}
// 		textlines = append(textlines, authLine)
// 	}
// 	if ehloCfg.EnhancedStatusCode {
// 		textlines = append(textlines, ("ENHANCEDSTATUSCODES"))
// 	}
// 	logger.Println("start tls", ehloCfg.Starttls, "ISTLSHANDSHAKE", isTlHandshakeAlreadyDone)
// 	if ehloCfg.Starttls && !isTlHandshakeAlreadyDone {
// 		textlines = append(textlines, "STARTTLS")
// 	}
// 	return &EhloReply{
// 		Reply: Reply{

// 			code:        code,
// 			textStrings: textlines,
// 		},
// 		domain: "gmail.com",
// 	}
// }
// func (r *EhloReply) HandleSmtpReply(w *writer.Writer) error {
// 	_, err := w.WriteString(r.format())
// 	return err
// }
