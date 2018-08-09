package pgsrv

import (
	"encoding/binary"
	"fmt"
)

// Msg is just an alias for a slice of bytes that exposes common operations on
// Postgres' client-server protocol messages.
// see: https://www.postgresql.org/docs/9.2/static/protocol-message-formats.html
// for postgres specific list of message formats
type msg []byte

// Type returns a string (single-char) representing the message type. The full
// list of available types is available in the aforementioned documentation.
func (m msg) Type() byte {
	var b byte
	if len(m) > 0 {
		b = m[0]
	}
	return b
}

func newMsg(b []byte) msg {
	return msg(b)
}

func errMsg(err error) msg {
	msg := []byte{'E', 0, 0, 0, 0}

	// https://www.postgresql.org/docs/9.3/static/protocol-error-fields.html
	fields := map[string]string{
		"S": "ERROR", // Severity
		"C": "XX000",
		"M": err.Error(),
	}

	// error code
	errCode, ok := err.(interface {
		Code() string
	})
	if ok && errCode.Code() != "" {
		fields["C"] = errCode.Code()
	}

	// detail
	errDetail, ok := err.(interface {
		Detail() string
	})
	if ok && errDetail.Detail() != "" {
		fields["D"] = errDetail.Detail()
	}

	// hint
	errHint, ok := err.(interface {
		Hint() string
	})
	if ok && errHint.Hint() != "" {
		fields["H"] = errHint.Hint()
	}

	// cursor position
	errPosition, ok := err.(interface {
		Position() int
	})
	if ok && errPosition.Position() >= 0 {
		fields["P"] = fmt.Sprintf("%d", errPosition.Position())
	}

	for k, v := range fields {
		msg = append(msg, byte(k[0]))
		msg = append(msg, []byte(v)...)
		msg = append(msg, 0) // NULL TERMINATED
	}

	msg = append(msg, 0) // NULL TERMINATED

	// write the length
	binary.BigEndian.PutUint32(msg[1:5], uint32(len(msg)-1))
	return msg
}
