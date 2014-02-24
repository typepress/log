package file

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

func New(name string) (*File, error) {
	name = filepath.ToSlash(name)
	isDir := strings.HasSuffix(name, "/")
	name, err := filepath.Abs(name)

	if err != nil {
		return nil, err
	}

	if !isDir {
		fi, _ := os.Stat(name)
		isDir = fi != nil && fi.IsDir()
	}
	if isDir {
		name = filepath.Join(name, `log.txt`)
	}

	dir, name := filepath.Split(name)
	fi, _ := os.Stat(name)

	if fi == nil {
		os.MkdirAll(dir, os.ModePerm)
	}

	ext := filepath.Ext(name)
	if len(ext) == 0 {
		ext = `.txt`
	} else {
		name = name[:len(name)-len(ext)]
	}
	prefix := strings.SplitN(name, `-`, 2)[0]

	f := &File{nil, dir, prefix, ext, name}
	t := time.Now()
	f.Rotate(t, t)
	return f, err
}

var layout = "20060102150405.00000"

type File struct {
	fd                     *os.File
	dir, prefix, ext, Name string
}

func (f *File) Rotate(begin, now time.Time) {
	f.Close()
	for {
		name := filepath.Join(
			f.dir, f.prefix+`-`+now.Format(layout)+f.ext,
		)
		fi, _ := os.Stat(name)
		if fi == nil && name != f.Name {
			f.Name = name
			break
		}
		now.Add(time.Now().Sub(now))
	}
	f.fd, _ = os.OpenFile(f.Name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0664)
	return
}

func (f *File) Write(b []byte) (n int, err error) {
	if f.fd == nil {
		return 0, os.ErrNotExist
	}
	return f.fd.Write(b)
}

func (f *File) Close() (err error) {
	if f.fd == nil {
		return nil
	}
	err = f.fd.Sync()
	if err == nil {
		err = f.fd.Close()
	} else {
		f.fd.Close()
	}
	f.fd = nil
	return
}
