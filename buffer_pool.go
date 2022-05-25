package ws

import (
	"bufio"
	"io"
	"os"
	"sync"
)

var (
	bufferReaderPool = sync.Pool{New: func() interface{} {
		return bufio.NewReader(os.Stdin)
	}}
	bufferWriterPool = sync.Pool{New: func() interface{} {
		return bufio.NewWriter(io.Discard)
	}}
	limitReaderPool = sync.Pool{New: func() interface{} {
		return &io.LimitedReader{
			R: nil,
			N: 0,
		}
	}}
)

func returnBuffReaderPoll(r *bufio.Reader) {
	bufferReaderPool.Put(r)
}

func newBufferReader(rr io.Reader) *bufio.Reader {
	r := bufferReaderPool.Get().(*bufio.Reader)
	r.Reset(rr)
	return r
}

func returnLimitReaderPoll(r *io.LimitedReader) {
	limitReaderPool.Put(r)
}

func newLimitReader(rr io.Reader, n int64) *io.LimitedReader {
	r := limitReaderPool.Get().(*io.LimitedReader)
	r.R = rr
	r.N = n
	return r
}

func newBuffWriter(ww io.Writer) *bufio.Writer {
	w := bufferReaderPool.Get().(*bufio.Writer)
	w.Reset(ww)
	return w
}

func returnBuffWriterPoll(w *bufio.Writer) {
	w.Flush()
	bufferWriterPool.Put(w)
}
