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
 - 多种输出规则
 - io.Writer 接口
 - io.ReaderFrom 接口
 - 友好输出格式易于分析
 - Loggers, 设计思路来自 https://github.com/uniqush/log.

Import
======

	import "github.com/achun/tom-toml"

Usage
=====

```go
package main

import (
	"bytes"
	"fmt"
	"github.com/typepress/log"
)

func main(){
	w := bytes.NewBuffer([]byte{})
	m := log.Multi(log.New(w, // io.Writer
		"prefix", // prefix string
		0,        // flag, 0 equal log.LZero
	))

	m.Add(log.New(w,
		"",
		log.LstdFlags|log.Lshortfile,
		log.Lmicroseconds,
	))

	m.Info("info")
	m.Notify("Notify")

	fmt.Println(w.String())
}
```

output:

	prefix [I] "info"
	[I] 2014-02-17 23:26:05.780139 <hello.go:22> "info"
	prefix [N] "Notify"
	[N] 2014-02-17 23:26:05.783140 <hello.go:23> "Notify"

License
=======
BSD-2-Clause
