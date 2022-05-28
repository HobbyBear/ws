package bufferpool

import (
	"bufio"
	"io"
	"os"
	"sync"
)

var (
	readerPool = sync.Pool{New: func() interface{} {
		return bufio.NewReader(os.Stdin)
	}}
	writerPool = sync.Pool{New: func() interface{} {
		return bufio.NewWriter(io.Discard)
	}}
	limitReaderPool = sync.Pool{New: func() interface{} {
		return &io.LimitedReader{
			R: nil,
			N: 0,
		}
	}}
)

func NewBufioReader(r io.Reader) *bufio.Reader {
	if v := readerPool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	return bufio.NewReader(r)
}

func PutBufioReader(br *bufio.Reader) {
	br.Reset(nil)
	readerPool.Put(br)
}

func PutLimitReader(r *io.LimitedReader) {
	limitReaderPool.Put(r)
}

func NewLimitReader(rr io.Reader, n int64) *io.LimitedReader {
	r := limitReaderPool.Get().(*io.LimitedReader)
	r.R = rr
	r.N = n
	return r
}

func NewBuffWriter(ww io.Writer) *bufio.Writer {
	w := writerPool.Get().(*bufio.Writer)
	w.Reset(ww)
	return w
}

func PutBuffWriter(w *bufio.Writer) {
	w.Flush()
	writerPool.Put(w)
}
