package exporttask

import (
	"bytes"
	"encoding/csv"
	"io"
)

type readerWriter struct {
	c chan [][]string
	b *bytes.Buffer
	w *csv.Writer
}

func NewReader() *readerWriter {
	b := bytes.NewBuffer([]byte{})
	b.Write([]byte("\xEF\xBB\xBF"))
	return &readerWriter{
		c: make(chan [][]string, 10),
		b: b,
		w: csv.NewWriter(b),
	}
}

func (r *readerWriter) Put(s [][]string) {
	r.c <- s
}

func (r *readerWriter) Read(p []byte) (n int, err error) {
	// 长度足够，直接读取返回
	if r.b.Len() >= len(p) {
		return r.b.Read(p)
	}
	// 先重置buffer，避免长度过长
	if r.b.Len() > 0 {
		b := r.b.Bytes()
		r.b.Reset()
		r.b.Write(b)
	}
	// 尝试从channel读取新的数据，然后返回
	for {
		s, ok := <-r.c
		if ok {
			e := r.w.WriteAll(s)
			if e != nil {
				return 0, e
			}
		} else {
			// channel close
			n, e := r.b.Read(p)
			if e != nil {
				return n, e
			}
			return n, io.EOF
		}
		if r.b.Len() >= len(p) {
			break
		}
	}
	return r.b.Read(p)
}

func (r *readerWriter) Close() {
	close(r.c)
}
