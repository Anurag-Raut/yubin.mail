package mta

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/Yubin-email/config"
	"github.com/Yubin-email/internal/io/reader"
	"github.com/Yubin-email/internal/io/writer"
	"github.com/Yubin-email/internal/logger"
	"github.com/Yubin-email/internal/parser"
	"github.com/Yubin-email/internal/smtp/command"
	"github.com/Yubin-email/internal/smtp/reply"
	"github.com/Yubin-email/internal/state"
)

type Session struct {
	mailState *state.MailState
	writer    *writer.Writer
	reader    *reader.Reader
	conn      net.Conn
}

func newSession(conn net.Conn) *Session {
	r := reader.NewReader(conn)
	w := writer.NewWriter(conn)
	return &Session{
		mailState: &state.MailState{},
		reader:    r,
		writer:    w,
		conn:      conn,
	}
}

func generateSelfSignedCert() (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}

	tmpl := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   config.ServerConfig.Domain,
			Organization: []string{config.ServerConfig.Organization},
		},
		DNSNames:    []string{config.ServerConfig.Domain},
		IPAddresses: []net.IP{net.ParseIP(config.ServerConfig.IP)},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	cert := tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  priv,
	}

	return cert, nil
}

var cert, _ = generateSelfSignedCert()

func (s *Session) upgradeToTLS() {
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	tlsConn := tls.Server(s.conn, config)
	s.conn = tlsConn
	s.reader = reader.NewReader(tlsConn)
	s.writer = writer.NewWriter(tlsConn)
	err := tlsConn.Handshake()
	if err != nil {
		fmt.Println("TLS handshake failed:", err)
		return
	}
	s.mailState.ClearAll()
	s.Begin(true)
}

func (s *Session) Begin(isTls bool) {
	if !isTls {
		reply.Greet(s.writer)
	}
	p := parser.NewParser(s.reader)
	cmdCtx := command.NewCommandContext(s.mailState, isTls)

	for {
		b, err := s.reader.Peek(1)
		if err != nil {
			logger.Println("Peek error:", err)
		} else {
			logger.Println("Expecting new command:", string(b))
		}
		cmd, err := command.GetCommand(p)
		logger.Println(cmd, "COMMANDcommand")
		if err != nil {
			return
			// reply.HandleParseError(writer, cmd.GetCommandType(), err)
		}

		replyChannel := make(chan reply.ReplyInterface)
		go cmd.ProcessCommand(cmdCtx, replyChannel)
	outerloop:
		for {

			select {
			case responseReply, ok := <-replyChannel:
				{
					if !ok {
						logger.Println("DONE??")
						break outerloop
					}
					_ = responseReply.HandleSmtpReply(s.writer)
				}
			case event := <-cmdCtx.EventChan:
				if event.Name == "TLS_UPGRADE" && !isTls {
					s.upgradeToTLS()
				}
			}
		}
	}

}

func HandleConn(conn net.Conn) {

	session := newSession(conn)
	session.Begin(false)

}
