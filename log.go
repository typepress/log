/*
  Copyright (c) 2014, YU HengChun
  All rights reserved.

  Redistribution and use in source and binary forms, with or without modification,
  are permitted provided that the following conditions are met:

  * Redistributions of source code must retain the above copyright notice, this
    list of conditions and the following disclaimer.

  * Redistributions in binary form must reproduce the above copyright notice, this
    list of conditions and the following disclaimer in the documentation and/or
    other materials provided with the distribution.

  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
  ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
  WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
  DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
  ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
  (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
  LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
  ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
  (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
  SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

// few code fork go log.
package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const (
	Ldate = 1 << iota
	Ltime
	Lmicroseconds
	Llongfile
	Lshortfile
	LstdFlags = Ldate | Ltime
)

// dont change order
const (
	LZero = -iota
	LFatal
	LPanic // recover panic
	LAlert
	LError
	LReport
	LNotify
	LInfo
	LDebug
	nr_levels
)

var levelsName [-nr_levels - 1]string

func init() {
	levelsName[-LFatal-1] = "[F]"
	levelsName[-LPanic-1] = "[P]"
	levelsName[-LAlert-1] = "[A]"
	levelsName[-LError-1] = "[E]"
	levelsName[-LReport-1] = "[R]"
	levelsName[-LNotify-1] = "[N]"
	levelsName[-LInfo-1] = "[I]"
	levelsName[-LDebug-1] = "[D]"
}

// dont change order
const (
	MODE_EQUAL     = -iota - 100 // equal level mode
	MODE_NONE_NAME               // dont write default level name
	MODE_DONT_EXIT               // dont exec os.Exit when Fatal
	MODE_DONT_PANIC
	MODE_RECOVER
)

type nullWriter struct{}

func (f *nullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

type BaseLogger interface {
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
	Alert(v ...interface{})
	Alertf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Report(v ...interface{})
	Reportf(format string, v ...interface{})
	Notify(v ...interface{})
	Notifyf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
}

type Logger interface {
	BaseLogger
	io.ReaderFrom
	Write([]byte) (int, error)
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Output(calldepth int, s string, optionLevel ...int) error
}

var _ Logger = &logger{}

type logger struct {
	mu     sync.Mutex // ensures atomic writes; protects the following fields
	prefix string     // prefix to write at beginning of each line
	flag   int        // properties
	out    io.Writer  // destination for output
	buf    []byte     // for accumulating text to write

	level     int
	equal     bool
	noneName  bool
	dontExit  bool
	dontPanic bool
	recover   bool
}

// Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
// Knows the buffer has capacity.
func itoa(buf *[]byte, i int, wid int) {
	var u uint = uint(i)
	if u == 0 && wid <= 1 {
		*buf = append(*buf, '0')
		return
	}

	// Assemble decimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; u > 0 || wid > 0; u /= 10 {
		bp--
		wid--
		b[bp] = byte(u%10) + '0'
	}
	*buf = append(*buf, b[bp:]...)
}

func (l *logger) formatHeader(buf *[]byte, t time.Time, file string, line, level int) {
	if len(l.prefix) != 0 {
		*buf = append(*buf, l.prefix...)
		if len(l.prefix) != 0 {
			*buf = append(*buf, ' ')
		}
	}

	if !l.noneName && level > nr_levels && level <= 0 {
		*buf = append(*buf, levelsName[-level]...)
		*buf = append(*buf, ' ')
	}

	if l.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if l.flag&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '-')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '-')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	if l.flag&(Lshortfile|Llongfile) != 0 {
		if l.flag&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, '<')
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, `> `...)
	}
}

func printf(format string, v []interface{}) string {
	if len(format) == 0 {
		return fmt.Sprint(v...)
	} else {
		return fmt.Sprintf(format, v...)
	}
}

func (l *logger) Output(calldepth int, s string, optionLevel ...int) (err error) {
	level := LZero
	if len(optionLevel) != 0 {
		level = optionLevel[0]
	}
	if level == LZero || l.equal && level == l.level || !l.equal && level > l.level {
		level++
	} else {
		return
	}
	now := time.Now() // get this early.
	var file string
	var line int
	if len(s) != 0 {
		s = strconv.Quote(s)
	}
	if l.flag&(Lshortfile|Llongfile) != 0 {
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
	}
	l.mu.Lock()
	defer func() {
		l.mu.Unlock()
		if l.recover {
			_ = recover() // ignore panic
		}
	}()

	l.buf = l.buf[:0]

	l.formatHeader(&l.buf, now, file, line, level)

	l.buf = append(l.buf, s...)
	if len(s) > 0 && s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	n := 0
	count := 0
	for len(l.buf) != 0 && (err == nil || err == io.ErrShortWrite) && count < 10 {
		n, err = l.out.Write(l.buf)
		l.buf = l.buf[n:]
		if err == io.ErrShortWrite {
			count++
		}
	}
	return
}

func (l *logger) ReadFrom(src io.Reader) (n int64, err error) {
	l.mu.Lock()
	defer func() {
		l.mu.Unlock()
		if l.recover {
			_ = recover() // ignore panic
		}
	}()
	return io.Copy(l.out, src)
}

func (l *logger) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer func() {
		l.mu.Unlock()
		if l.recover {
			_ = recover() // ignore panic
		}
	}()
	wBytes := 0
	count := 0
	for len(p) != 0 && (err == nil || err == io.ErrShortWrite) && count < 10 {
		wBytes, err = l.out.Write(p)
		n += wBytes
		p = p[n:]
		if err == io.ErrShortWrite {
			count++
		}
	}
	return
}

func (l *logger) Print(v ...interface{}) {
	l.Output(2, printf("", v), LZero)
}

func (l *logger) Printf(format string, v ...interface{}) {
	l.Output(2, printf(format, v), LZero)
}

func (l *logger) Debug(v ...interface{}) {
	l.Output(2, printf("", v), LDebug)
}

func (l *logger) Debugf(format string, v ...interface{}) {
	l.Output(2, printf(format, v), LDebug)
}

func (l *logger) Info(v ...interface{}) {
	l.Output(2, printf("", v), LInfo)
}

func (l *logger) Infof(format string, v ...interface{}) {
	l.Output(2, printf(format, v), LInfo)
}

func (l *logger) Notify(v ...interface{}) {
	l.Output(2, printf("", v), LNotify)
}

func (l *logger) Notifyf(format string, v ...interface{}) {
	l.Output(2, printf(format, v), LNotify)
}

func (l *logger) Report(v ...interface{}) {
	l.Output(2, printf("", v), LReport)
}

func (l *logger) Reportf(format string, v ...interface{}) {
	l.Output(2, printf(format, v), LReport)
}

func (l *logger) Error(v ...interface{}) {
	l.Output(2, printf("", v), LError)
}

func (l *logger) Errorf(format string, v ...interface{}) {
	l.Output(2, printf(format, v), LError)
}

func (l *logger) Alert(v ...interface{}) {
	l.Output(2, printf("", v), LAlert)
}

func (l *logger) Alertf(format string, v ...interface{}) {
	l.Output(2, printf(format, v), LAlert)
}

func (l *logger) Panic(v ...interface{}) {
	l.Output(2, printf("", v), LPanic)
	if !l.dontPanic {
		panic(v)
	}
}

func (l *logger) Panicf(format string, v ...interface{}) {
	l.Output(2, printf(format, v), LPanic)
	if !l.dontPanic {
		panic(v)
	}
}

func (l *logger) Fatal(v ...interface{}) {
	l.Output(2, printf("", v), LFatal)
	if !l.dontExit {
		os.Exit(1)
	}
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	l.Output(2, printf(format, v), LFatal)
	if !l.dontExit {
		os.Exit(1)
	}
}

func New(writer io.Writer, prefix string, flags ...int) Logger {
	ret := new(logger)
	ret.level = nr_levels
	hasflags := false
	for _, flag := range flags {
		switch flag {
		case MODE_EQUAL:
			ret.equal = true
		case MODE_NONE_NAME:
			ret.noneName = true
		case MODE_DONT_EXIT:
			ret.dontExit = true
		case MODE_DONT_PANIC:
			ret.dontPanic = true
		case MODE_RECOVER:
			ret.recover = true
		default:

			if flag >= 0 {
				hasflags = true
				ret.flag = ret.flag | flag
			} else {
				ret.level = flag
			}
		}
	}
	// defaults to LstdFlags.
	if !hasflags {
		ret.flag = LstdFlags
	}

	if ret.level <= nr_levels {
		ret.level = nr_levels + 1
	}

	ret.prefix = prefix
	ret.out = writer
	return ret
}
