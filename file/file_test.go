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
	wt := want.T(t)
	f, err := New(path)

	wt.Nil(err)
	defer func() {
		wt.Nil(f.Close())
		wt.Nil(os.Remove(f.Name))
	}()

	wt.Nil(want.LastError(
		f.Write([]byte("string line\n")),
	))
	f.fd.Sync()
	now := time.Now()
	name := f.Name
	f.Rotate(now, now)
	wt.True(f.Name != name, "rotate failed: ", name)
	wt.Nil(os.Remove(name))
}
