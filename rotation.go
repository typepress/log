package log

import (
	"io"
	"time"
)

// RotateWrite interface for rotation logger.
type RotateWrite interface {
	io.Writer
	Rotate(begin, now time.Time)
}

type rotate struct {
	w               RotateWrite
	msize, mrecodes int
	minutes         int64
	size, recodes   int
	seconds         int64
	begin           time.Time
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
	if p == nil || len(p) != 0 {
		n, err = r.w.Write(p)
	}

	if len(p) == 0 && r.mrecodes != 0 {
		r.recodes++
		if r.recodes >= r.mrecodes {
			do = true
		}
	}
	if r.msize != 0 {
		r.size += n
		if r.size >= r.msize {
			do = true
		}
	}
	if do || r.minutes != 0 {
		now = time.Now()
	}
	if r.minutes != 0 {
		if now.Unix() >= r.seconds {
			do = true
		}
	}
	if do {
		begin, to := r.begin, now
		r.begin = now
		r.size = 0
		r.recodes = 0
		r.seconds = now.Add(time.Duration(r.minutes * int64(time.Minute))).Unix()
		r.w.Rotate(begin, to)
	}
	return
}

// +dl zh-cn
// Rotate 包装 RotateWrite 对象, 返回 io.Writer. 当达到分割条件 RotateWrite.Rotate 被调用.
// 具体分割行为由 RotateWrite 对象自己完成.
// +dl

// Rotate wrapper RotateWrite, returns io.Writer. invoke RotateWrite.Rotate method by the time.
func Rotate(w RotateWrite, size, recodes, minutes int) io.Writer {
	var s int64

	if w == nil {
		return nil
	}
	now := time.Now()
	if minutes != 0 {
		s = now.Add(time.Duration(int64(minutes) * int64(time.Minute))).Unix()
	}

	return &rotate{w, size, recodes, int64(minutes), 0, 0, s, now}
}
