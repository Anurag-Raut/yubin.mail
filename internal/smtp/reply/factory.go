package reply

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/Yubin-email/internal/io/writer"
	"github.com/Yubin-email/internal/parser"
	"gopkg.in/yaml.v3"
)

type ehloConfig struct {
	Starttls              bool     `yaml:"starttls"`
	AllowedAuthMechanisms []string `yaml:"allowedAuthMechanisms"`
	EnhancedStatusCode    bool     `yaml:"enhancedStatusCode"`
	ServerName            string   `yaml:"serverName"`
}

func getEhloConfig() ehloConfig {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	cfg := ehloConfig{}
	data, err := os.ReadFile(filepath.Join(dir, "ehlo_config.yaml"))
	if err != nil {
		panic("Ehlo config file not found")
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		panic("Error while unmarshalling ehlo yaml")
	}
	return cfg
}

var ehloCfg = getEhloConfig()

type ReplyInterface interface {
	ParseReply() error
	Execute() error
	GetReplyCode() string
	format() string
	HandleSmtpReply(w *writer.Writer) error
}

type Reply struct {
	code        uint16
	textStrings []string
	parser      *parser.Parser
}

type GreetingReply struct {
	serverIdentifier string
	Reply
}

type EhloReply struct {
	domain string
	Reply
}
