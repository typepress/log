package smtp

import (
	"testing"
)

func TestSmtp(t *testing.T) {
	arg := Sets{
		Identity: ``,
		Username: ``,
		Password: ``,
		Host:     `smtp.gmail.com:587`,
		Subject:  `TestSmtp`,
		To:       []string{""},
	}
	if arg.Password == "" {
		println("please configer Username Password ...")
		return
	}
	_, err := New(arg).Write(
		[]byte(`is me
中文可以么
<b>html 加粗么</b>
`))
	if err != nil {
		t.Fatal(err)
	}
}
