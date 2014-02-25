log
===

[![wercker status](https://app.wercker.com/status/cf8711e6d779a53c6a7921716db35df5/s/ "wercker status")](https://app.wercker.com/project/bykey/cf8711e6d779a53c6a7921716db35df5)
提供日志输出支持.

The documentation is available at
[gowalker.org](http://gowalker.org/github.com/typepress/log).

Support
=======

 - 并发安全
 - 日志级别
 - 设定输出级别
 - 多种输出规则
 - io.WriteCloser 接口
 - 支持日志分割 RotateWriter 接口
 - 友好输出格式易于分析
 - Loggers, Multi-Logger 设计思路来自 https://github.com/uniqush/log.
 - 内建 File, Smtp 实现

Import
======

	import "github.com/typepress/log"

Usage
=====

```go
package main

import (
	"bytes"
	"fmt"
	"github.com/typepress/log"
)

func main() {
	w := bytes.NewBuffer(nil)
	m := log.Multi(
		log.New(
			w,        // io.Writer
			"prefix", // prefix string
			0,        // flag, 0 means none auto prefix.
		))

	l := log.New(w, "",
		log.LstdFlags|log.Lshortfile,
		log.Lmicroseconds,
	)

	m.Join(l)

	m.Info("info")
	m.Notify("Notify")

	l.Output(1, "Output")

	l = log.New(w, "ModeEqualReport",
		0,
		log.MODE_EQUAL, // equal mode
		log.LReport,    // set level
	)

	m.Join(l)

	m.Error("Error")
	m.Report("Report")

	fmt.Println(w.String())
}
```

output:

	prefix [I] "info"
	[I] 2014-02-18 17:30:27.154305 <hello.go:25> "info"
	prefix [N] "Notify"
	[N] 2014-02-18 17:30:27.156305 <hello.go:26> "Notify"
	[Z] 2014-02-18 17:30:27.156305 <hello.go:28> "Output"
	[A] 2014-02-18 17:30:27.156305 <hello.go:31> "SetPrintLevel: LAlert"
	prefix [E] "Error"
	[E] 2014-02-18 17:30:27.156305 <hello.go:38> "Error"
	prefix [R] "Report"
	[R] 2014-02-18 17:30:27.156305 <hello.go:39> "Report"
	ModeEqualReport [R] "Report"

License
=======
BSD-2-Clause
