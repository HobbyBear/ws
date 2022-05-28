package wsgateway

import (
	"encoding/binary"
	"github.com/gobwas/ws"
	"io"
	"time"
	"ws/pkg/bufferpool"
)

type packet struct {
	header ws.Header
	data   []byte
}

func protocolContent(conn *Conn, headDeadLine time.Duration, bodyDeadLine time.Duration) ([]*packet, error) {

	var (
		r       = bufferpool.NewBufioReader(conn.rawConn)
		packets []*packet
	)
	defer bufferpool.PutBufioReader(r)
	for {
		var buf []byte
	readFrame:
		conn.rawConn.SetReadDeadline(time.Now().Add(headDeadLine))
		header, err := readHeader(r)
		if err != nil {
			return packets, err
		}
		lr := bufferpool.NewLimitReader(r, header.Length)
		conn.rawConn.SetReadDeadline(time.Now().Add(bodyDeadLine))
		payload, err := io.ReadAll(lr)
		bufferpool.PutLimitReader(lr)
		if err != nil {
			return packets, err
		}

		if header.Masked {
			ws.Cipher(payload, header.Mask, 0)
		}

		if header.OpCode == ws.OpContinuation && !header.Fin {
			buf = append(buf, payload...)
			goto readFrame
		}
		if len(buf) != 0 {
			packets = append(packets, &packet{
				header: header,
				data:   append(buf, payload...),
			})
		} else {
			packets = append(packets, &packet{
				header: header,
				data:   payload,
			})
		}
		if r.Buffered() <= 0 {
			break
		}
	}
	return packets, nil
}

const (
	bit0 = 0x80
)

// readHeader reads a frame header from r.
func readHeader(r io.Reader) (h ws.Header, err error) {

	// Make slice of bytes with capacity 12 that could hold any header.
	//
	// The maximum header size is 14, but due to the 2 hop reads,
	// after first hop that reads first 2 constant bytes, we could reuse 2 bytes.
	// So 14 - 2 = 12.

	// Prepare to hold first 2 bytes to choose size of next read.
	lr := bufferpool.NewLimitReader(r, 2)
	defer bufferpool.PutLimitReader(lr)

	bts, err := io.ReadAll(lr)
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
	lr.N = int64(extra)
	lr.R = r
	bts, err = io.ReadAll(lr)
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
