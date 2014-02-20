package smtp

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

const (
	subjectPhrase = "Diagnostic message from server"
)

type Sets struct {
	Identity string
	Username string
	Password string
	Host     string
	Subject  string
	To       []string
}

type Smtp struct {
	arg  Sets
	auth smtp.Auth
}

func New(arg Sets) *Smtp {
	arg.To = append([]string{}, arg.To...)
	names := strings.Split(arg.Username, `<`)
	if len(names) == 2 {
		names[0] = strings.TrimSuffix(names[1], ">")
	}
	auth := smtp.PlainAuth(arg.Identity, names[0], arg.Password, strings.Split(arg.Host, ":")[0])
	return &Smtp{arg, auth}
}

func (s *Smtp) Rotate(begin, now time.Time) {
}

func (s *Smtp) Write(b []byte) (n int, err error) {
	header := fmt.Sprintf(
		"Subject: %s\r\nFrom: %s\r\nTo: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n",
		s.arg.Subject, s.arg.Username, strings.Join(s.arg.To, ";"))

	err = smtp.SendMail(
		s.arg.Host,
		s.auth,
		s.arg.Username,
		s.arg.To,
		append([]byte(header), b...),
	)

	n = len(b)
	return
}

func (s *Smtp) Close() (err error) {
	return
}
