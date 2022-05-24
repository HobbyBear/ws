package ws

import (
	"bufio"
	"encoding/binary"
	"github.com/gobwas/ws"
	"io"
)

func getProtocolContent(r *bufio.Reader) (ws.Header, []byte, error) {
	header, err := ReadHeader(r)
	if err != nil {
		// handle error
		return header, nil, err
	}

	payload, err := io.ReadAll(io.LimitReader(r, header.Length))
	if err != nil {
		return header, nil, err
	}
	return header, payload, nil
}

const (
	bit0 = 0x80
)

// ReadHeader reads a frame header from r.
func ReadHeader(r *bufio.Reader) (h ws.Header, err error) {

	// Make slice of bytes with capacity 12 that could hold any header.
	//
	// The maximum header size is 14, but due to the 2 hop reads,
	// after first hop that reads first 2 constant bytes, we could reuse 2 bytes.
	// So 14 - 2 = 12.

	// Prepare to hold first 2 bytes to choose size of next read.
	bts, err := io.ReadAll(io.LimitReader(r, 2))
	if err != nil || len(bts) == 0 {
		return
	}

	h.Fin = bts[0]&bit0 != 0
	h.Rsv = (bts[0] & 0x70) >> 4
	h.OpCode = ws.OpCode(bts[0] & 0x0f)

	var extra int

	if bts[1]&bit0 != 0 {
		h.Masked = true
		extra += 4
	}

	length := bts[1] & 0x7f
	switch {
	case length < 126:
		h.Length = int64(length)

	case length == 126:
		extra += 2

	case length == 127:
		extra += 8

	default:
		err = ws.ErrHeaderLengthUnexpected
		return
	}

	if extra == 0 {
		return
	}

	// Increase len of bts to extra bytes need to read.
	// Overwrite first 2 bytes that was read before.
	bts, err = io.ReadAll(io.LimitReader(r, int64(extra)))
	if err != nil {
		return
	}

	switch {
	case length == 126:
		h.Length = int64(binary.BigEndian.Uint16(bts[:2]))
		bts = bts[2:]

	case length == 127:
		if bts[0]&0x80 != 0 {
			err = ws.ErrHeaderLengthMSB
			return
		}
		h.Length = int64(binary.BigEndian.Uint64(bts[:8]))
		bts = bts[8:]
	}

	if h.Masked {
		copy(h.Mask[:], bts)
	}

	return
}