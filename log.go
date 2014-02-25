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

// +dl zh-cn
/*
  为 golang 提供日志输出. 部分代码 fork 自 go log.
*/
// +dl

// logger for golang.
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

// +dl zh-cn
// 自动前缀 flags 常量, 生成相应的前缀. 缺省值 LstdFlags.
// +dl

// auto flags constants for prefix.
// see also http://godoc.org/log#pkg-constants
const (
	Ldate = 1 << iota
	Ltime
	Lmicroseconds
	Llongfile
	Lshortfile
	LstdFlags = Ldate | Ltime // default
)

// +dl zh-cn
// 级别 flags 常量. 每一个级别都对应一个级别缩写前缀.
// +dl

// flags constants for logger level.
const (
	LZero   = -iota // [Z], use only for Logger.Output.
	LFatal          // [F]
	LPanic          // [P]
	LAlert          // [A]
	LError          // [E]
	LReport         // [R]
	LNotify         // [N]
	LInfo           // [I], default
	LDebug          // [D]
	nr_levels
)

var levelsName [-nr_levels]string

func init() {
	levelsName[-LZero] = "[Z]"
	levelsName[-LFatal] = "[F]"
	levelsName[-LPanic] = "[P]"
	levelsName[-LAlert] = "[A]"
	levelsName[-LError] = "[E]"
	levelsName[-LReport] = "[R]"
	levelsName[-LNotify] = "[N]"
	levelsName[-LInfo] = "[I]"
	levelsName[-LDebug] = "[D]"
}

// +dl zh-cn
/*
  mode flags 常量.
  MODE_EQUAL 表示 API 调用级别和 Logger 级别一致才输出日志.
  MODE_NONE_NAME 不输出级别对应的缩写前缀.
  MODE_NONE_EOR  禁止写 EOR, 默认通过 Output 的输出最后都写 []byte{}, 表示 End Of Recorde.
  MODE_DONT_EXIT 调用 Fatal/Fatalf 时不执行 os.Exit(1).
  MODE_DONT_PANIC 调用 Panic/Panicf 时不抛出 panic.
  MODE_RECOVER 输出日志时使用 recover() 捕获并忽略 panic.
*/

// +dl

// flags constants for logger mode.
const (
	MODE_EQUAL      = -iota - 100 // equal level mode
	MODE_RECOVER                  // recover panic and ignore
	MODE_NONE_NAME                // dont output builtin level name
	MODE_NONE_EOR                 // send []byte{} on Output.
	MODE_DONT_EXIT                // dont exec os.Exit when Fatal
	MODE_DONT_PANIC               // dont exec panic when Panic
	nr_modes
)

const (
	_equal = 1 << iota
	_recover
	_none_name
	_none_eor
	_dont_exit
	_dont_panic
)

// level logger interface.
type LevelLogger interface {
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

// +dl zh-cn
// 所有 BaseLogger 方法在最后保障发送换行符 "\n".
// +dl

// All methods of BaseLogger auto send "\n".
type BaseLogger interface {
	LevelLogger
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	// SetPrintLevel to binding level for Print/Printf
	SetPrintLevel(level int)
	// +dl zh-cn
	/*
		Output 输出日志符串 s.
			参数 calldepth, optionLevel 作用于生成前缀,	其他 BaseLogger 接口方法均调用了 Output.

		calldepth 用于生成 Llongfile 或 Lshortfile 字符串.

		optionLevel 指示日志的级别, 级别和 MODE_EQUAL 共同决定是否输出日志:
			- 省略, 等同于 LZero, 总是输出日志.
			- MODE_EQUAL 下, 如果 optionLevel 等于建立 Logger 时的级别, 输出日志.
			- 非 MODE_EQUAL, 如果 optionLevel 大于等于建立 Logger 时的级别, 输出日志.
	*/
	// +dl

	/*
		Output writes the output for a logging event.
		if omit optionLevel, same to LZero, means always output.
	*/
	Output(calldepth int, s string, optionLevel ...int) error
}

type Logger interface {
	BaseLogger
	io.WriteCloser
}

var _ Logger = &logger{}

var endOfRecord []byte = []byte{}

type logger struct {
	mu     sync.Mutex // ensures atomic writes; protects the following fields
	prefix string     // prefix to write at beginning of each line
	flag   int        // properties
	out    io.Writer  // destination for output
	buf    []byte     // for accumulating text to write

	level int
	modes int

	printLevel int //
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

	if level > nr_levels && level <= 0 && 0 == _none_name&l.modes {
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
	if level >= LZero || 0 != _equal&l.modes && level == l.level || 0 == _equal&l.modes && level >= l.level {

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
		if 0 != _recover&l.modes {
			_ = recover() // ignore panic
		}
	}()

	if level > LZero {
		level = l.printLevel
	}

	l.buf = l.buf[:0]

	l.formatHeader(&l.buf, now, file, line, level)

	l.buf = append(l.buf, s...)
	if len(s) > 0 && s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err = l.out.Write(l.buf)
	if err == nil && 0 == _none_eor&l.modes {
		_, err = l.out.Write(endOfRecord)
	}
	return
}

func (l *logger) Close() error {
	l.mu.Lock()
	defer func() {
		l.mu.Unlock()
		if 0 != _recover&l.modes {
			_ = recover() // ignore panic
		}
	}()
	c, ok := l.out.(io.Closer)
	if ok {
		return c.Close()
	}
	return nil
}

func (l *logger) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer func() {
		l.mu.Unlock()
		if 0 != _recover&l.modes {
			_ = recover() // ignore panic
		}
	}()
	n, err = l.out.Write(p)
	return
}

func (l *logger) Print(v ...interface{}) {
	l.Output(2, printf("", v), 1)
}

func (l *logger) Printf(format string, v ...interface{}) {
	l.Output(2, printf(format, v), 1)
}

func (l *logger) SetPrintLevel(level int) {
	if level <= LZero && level > nr_levels {
		l.mu.Lock()
		l.printLevel = level
		l.mu.Unlock()
	}
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
	if 0 == _dont_panic&l.modes {
		panic(v)
	}
}

func (l *logger) Panicf(format string, v ...interface{}) {
	l.Output(2, printf(format, v), LPanic)
	if 0 == _dont_panic&l.modes {
		panic(v)
	}
}

func (l *logger) Fatal(v ...interface{}) {
	l.Output(2, printf("", v), LFatal)
	if 0 == _dont_exit&l.modes {
		os.Exit(1)
	}
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	l.Output(2, printf(format, v), LFatal)
	if 0 == _dont_exit&l.modes {
		os.Exit(1)
	}
}

// +dl zh-cn
/*
  New 返回 Logger 对象.
  writer 为 nil 返回 nil. 如果 writer 符合 RotateWrite 接口, 启用缺省分割日志支持.
  prefix 为自定义前缀, 自定义前缀总会被输出.
  flags 有效范围包括所有的 flags 常量, 0 特指取消自动生成的前缀.
  缺省分割日志条件: 达到任意一个 256M, 1000000 记录, 7days.
*/
// +dl

/*
  New returns Logger.
  if writer is nil, returns nil.
  flags The valid range includes all flags constants.
*/
func New(writer io.Writer, prefix string, flags ...int) Logger {
	if writer == nil {
		return nil
	}
	ret := new(logger)
	ret.level = nr_levels
	hasflags := false
	for _, flag := range flags {

		if flag > nr_modes && flag <= MODE_EQUAL {
			flag = -flag - 100
			ret.modes = ret.modes | 1<<uint(flag)
			continue
		}

		if flag >= 0 {
			hasflags = true
			ret.flag = ret.flag | flag
		} else {
			ret.level = flag
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
	rw, ok := writer.(RotateWriter)
	if ok {
		ret.out = Rotate(rw, RotateSets{1 << 28 /*256M*/, 1000000, 60 * 24 * 7 /*7days*/})
	} else {
		ret.out = writer
	}
	return ret
}
