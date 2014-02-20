package file

import (
	"github.com/achun/testing-want"
	"os"
	"testing"
	"time"
)

func TestFile(t *testing.T) {
	Testfile(t, `_test/`)
	Testfile(t, `_test`)
	Testfile(t, `./_test`)
	Testfile(t, `./_test/`)

	Testfile(t, `_test/log`)
	Testfile(t, `_test/log`)
	Testfile(t, `./_test/log`)
	Testfile(t, `./_test/log`)
	Testfile(t, `_test/access-log`)
}
func Testfile(t *testing.T, path string) {
	f, err := New(path)
	want.Nil(t, err)
	defer func() {
		want.Nil(t, f.Close())
		println(f.Name)
		want.Nil(t, os.Remove(f.Name))
	}()

	want.Nil(t, want.Err(
		f.Write([]byte("string line\n")),
	))
	f.fd.Sync()
	now := time.Now()
	println(f.Name)
	name := f.Name
	f.Rotate(now, now)
	want.Nil(t, os.Remove(name))
}
