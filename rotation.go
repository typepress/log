package log

import (
	"io"
	"time"
)

// RotateWriter interface for rotation logger.
type RotateWriter interface {
	io.Writer
	Rotate(begin, now time.Time)
}

type rotate struct {
	w                   RotateWriter
	maxSize, maxRecodes int
	minutes             int64
	size, recodes       int
	seconds             int64
	begin               time.Time
}

func (r *rotate) Close() {
	c, ok := r.w.(io.Closer)
	if ok {
		c.Close()
	}
}

func (r *rotate) Write(p []byte) (n int, err error) {
	var now time.Time
	var do bool

	// 特别的 p==nil 也执行 Write, 可以满足一些特殊需求
	if p == nil || len(p) != 0 {
		n, err = r.w.Write(p)
	}

	if len(p) == 0 && r.maxRecodes > 0 {
		r.recodes++
		if r.recodes >= r.maxRecodes {
			do = true
		}
	}
	if !do && r.maxSize > 0 {
		r.size += n
		if r.size >= r.maxSize {
			do = true
		}
	}
	if do || r.minutes > 0 {
		now = time.Now()
	}
	if !do && r.minutes > 0 && now.Unix() >= r.seconds {
		do = true
	}

	if do {
		begin, to := r.begin, now
		r.begin = now
		r.size = 0
		r.recodes = 0
		if r.minutes > 0 {
			r.seconds = now.Add(time.Duration(r.minutes * int64(time.Minute))).Unix()
		}
		r.w.Rotate(begin, to)
	}
	return
}

// +dl zh-cn
// Rotate 包装 RotateWriter 对象, 返回 io.Writer. 当达到分割条件 RotateWriter.Rotate 被调用.
// 具体分割行为由 RotateWriter 对象自己完成.
//
// 当 RotateSets 属性值为 0 时, 采用下述缺省值
//
//   Size    1 << 28 256M
//   Recodes 1000000
//   Minutes 60*24*7 7days
//
// 当 RotateSets 属性值小于 0 时, 表示忽略此属性.
// +dl

// Rotate wrapper RotateWriter, returns io.Writer. invoke RotateWriter.Rotate method by the time.
func Rotate(w RotateWriter, sets RotateSets) io.Writer {
	var seconds int64

	if w == nil {
		return nil
	}

	size, recodes, minutes := sets.Size, sets.Recodes, sets.Minutes

	if size == 0 {
		size = 1 << 28
	}
	if recodes == 0 {
		recodes = 1000000
	}
	if minutes == 0 {
		minutes = 60 * 24 * 7
	}

	now := time.Now()
	if minutes > 0 {
		seconds = now.Add(time.Duration(int64(minutes) * int64(time.Minute))).Unix()
	}

	return &rotate{w, size, recodes, int64(minutes), 0, 0, seconds, now}
}

// +dl zh-cn
// RotateSets 把 Rotate 的配置参数包装成 struct,
// 这样有利于从 toml 配置文件中直接进行 Apply
// +dl

// RotateSets for Rotate
type RotateSets struct {
	Size, Recodes, Minutes int
}
