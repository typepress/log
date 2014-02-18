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

package log

import (
	"sync"
)

// Loggers interface for set of Logger.
type Loggers interface {
	LevelLogger

	// +dl zh-cn
	// Join 增加 Logger 到 Loggers 集合.
	// +dl

	// Join logger to Loggers.
	Join(...Logger)
}

type multi struct {
	mu      sync.RWMutex
	loggers []Logger
}

var _ Loggers = &multi{}

// +dl zh-cn
/*
  Multi 把多个 Logger 合并为一个 Multi-Logger 集合 Loggers.
  调用 Loggers 的方法, Loggers 会对应遍历调用集合中的 LevelLogger 方法.
*/
// +dl

// Multi returns Loggers.
// Inspired by https://github.com/uniqush/log.
func Multi(loggers ...Logger) Loggers {
	return &multi{sync.RWMutex{}, loggers}
}

func (self *multi) Join(logger ...Logger) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.loggers = append(self.loggers, logger...)
}

func (self *multi) output(s string, level int) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	for _, l := range self.loggers {
		if l != nil {
			l.Output(3, s, level)
		}
	}
}

func (self *multi) Debug(v ...interface{}) {
	self.output(printf("", v), LDebug)
}

func (self *multi) Debugf(format string, v ...interface{}) {
	self.output(printf(format, v), LDebug)
}

func (self *multi) Info(v ...interface{}) {
	self.output(printf("", v), LInfo)
}

func (self *multi) Infof(format string, v ...interface{}) {
	self.output(printf(format, v), LInfo)
}

func (self *multi) Notify(v ...interface{}) {
	self.output(printf("", v), LNotify)
}

func (self *multi) Notifyf(format string, v ...interface{}) {
	self.output(printf(format, v), LNotify)
}

func (self *multi) Report(v ...interface{}) {
	self.output(printf("", v), LReport)
}

func (self *multi) Reportf(format string, v ...interface{}) {
	self.output(printf(format, v), LReport)
}

func (self *multi) Error(v ...interface{}) {
	self.output(printf("", v), LError)
}

func (self *multi) Errorf(format string, v ...interface{}) {
	self.output(printf(format, v), LError)
}

func (self *multi) Alert(v ...interface{}) {
	self.output(printf("", v), LAlert)
}

func (self *multi) Alertf(format string, v ...interface{}) {
	self.output(printf(format, v), LAlert)
}

func (self *multi) Fatal(v ...interface{}) {
	self.output(printf("", v), LFatal)
}

func (self *multi) Fatalf(format string, v ...interface{}) {
	self.output(printf(format, v), LFatal)
}

func (self *multi) Panic(v ...interface{}) {
	self.output(printf("", v), LPanic)
}

func (self *multi) Panicf(format string, v ...interface{}) {
	self.output(printf(format, v), LPanic)
}
