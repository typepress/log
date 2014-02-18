package log

import (
	"bytes"
	"testing"
)

func TestLog(t *testing.T) {
	w := bytes.NewBuffer(nil)
	l := New(w, "prefix", 0)
	l.Info("info")
	check(t, w, `prefix [I] "info"`)

	l.Alert("alert")
	check(t, w, `prefix [A] "alert"`)

	l = New(w, "", 0, MODE_NONE_NAME)
	l.Info("info")
	check(t, w, `"info"`)

	l.Alert("alert")
	check(t, w, `"alert"`)

	l = New(w, "", 0, MODE_NONE_NAME, MODE_EQUAL, LAlert)
	l.Info("info")
	check(t, w, "")

	l.Alert("alert")
	check(t, w, `"alert"`)

	l = New(w, "", 0, MODE_NONE_NAME, MODE_DONT_EXIT)
	l.Fatal("fatal")
	check(t, w, `"fatal"`)

	l.Notifyf("%#v %v %v", "notifyf", 1, true)
	check(t, w, `"\"notifyf\" 1 true"`)

	l = New(w, "", MODE_NONE_NAME, Lshortfile)
	l.Info("report")
	check(t, w, `<log_test.go:39> "report"`)
	l.Print("Print")
	check(t, w, `<log_test.go:41> "Print"`)

	l.Write([]byte("Write\n"))
	check(t, w, `Write`)

	m := Multi(l)
	l = New(w, "", 0, MODE_NONE_NAME)
	m.Join(l)
	m.Info("Multi Info")
	check(t, w, `<log_test.go:50> "Multi Info"`+"\n"+`"Multi Info"`)
}

func check(t *testing.T, w *bytes.Buffer, want string) {
	defer w.Reset()
	got := w.String()
	if len(want) == 0 && len(got) != 0 || len(want) != 0 && got != want+"\n" {
		t.Errorf("want: %#v, but got: %#v", want, got)
	}
}
