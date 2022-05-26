package ws

import (
	"bufio"
	"io"
	"os"
	"sync"
)

var (
	bufioReaderPool = sync.Pool{New: func() interface{} {
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

func newBufioReader(r io.Reader) *bufio.Reader {
	if v := bufioReaderPool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	return bufio.NewReader(r)
}

func putBufioReader(br *bufio.Reader) {
	br.Reset(nil)
	bufioReaderPool.Put(br)
}

func putLimitReader(r *io.LimitedReader) {
	limitReaderPool.Put(r)
}

func newLimitReader(rr io.Reader, n int64) *io.LimitedReader {
	r := limitReaderPool.Get().(*io.LimitedReader)
	r.R = rr
	r.N = n
	return r
}

func newBuffWriter(ww io.Writer) *bufio.Writer {
	w := bufferWriterPool.Get().(*bufio.Writer)
	w.Reset(ww)
	return w
}

func putBuffWriter(w *bufio.Writer) {
	w.Flush()
	bufferWriterPool.Put(w)
}
