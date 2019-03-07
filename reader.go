package brotli

import (
	"errors"
	"io"
)

type decodeError int

func (err decodeError) Error() string {
	return "brotli: " + string(BrotliDecoderErrorString(int(err)))
}

var errExcessiveInput = errors.New("brotli: excessive input")
var errInvalidState = errors.New("brotli: invalid state")
var errReaderClosed = errors.New("brotli: Reader is closed")

// readBufSize is a "good" buffer size that avoids excessive round-trips
// between C and Go but doesn't waste too much memory on buffering.
// It is arbitrarily chosen to be equal to the constant used in io.Copy.
const readBufSize = 32 * 1024

// NewReader initializes new Reader instance.
func NewReader(src io.Reader) *Reader {
	r := new(Reader)
	BrotliDecoderStateInit(r)
	r.src = src
	r.buf = make([]byte, readBufSize)
	return r
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if !BrotliDecoderHasMoreOutput(r) && len(r.in) == 0 {
		m, readErr := r.src.Read(r.buf)
		if m == 0 {
			// If readErr is `nil`, we just proxy underlying stream behavior.
			return 0, readErr
		}
		r.in = r.buf[:m]
	}

	if len(p) == 0 {
		return 0, nil
	}

	for {
		var written uint
		in_len := uint(len(r.in))
		out_len := uint(len(p))
		in_remaining := in_len
		out_remaining := out_len
		result := BrotliDecoderDecompressStream(r, &in_remaining, &r.in, &out_remaining, &p, nil)
		written = out_len - out_remaining
		n = int(written)

		switch result {
		case BROTLI_DECODER_RESULT_SUCCESS:
			if len(r.in) > 0 {
				return n, errExcessiveInput
			}
			return n, nil
		case BROTLI_DECODER_RESULT_ERROR:
			return n, decodeError(BrotliDecoderGetErrorCode(r))
		case BROTLI_DECODER_RESULT_NEEDS_MORE_OUTPUT:
			if n == 0 {
				return 0, io.ErrShortBuffer
			}
			return n, nil
		case BROTLI_DECODER_NEEDS_MORE_INPUT:
		}

		if len(r.in) != 0 {
			return 0, errInvalidState
		}

		// Calling r.src.Read may block. Don't block if we have data to return.
		if n > 0 {
			return n, nil
		}

		// Top off the buffer.
		encN, err := r.src.Read(r.buf)
		if encN == 0 {
			// Not enough data to complete decoding.
			if err == io.EOF {
				return 0, io.ErrUnexpectedEOF
			}
			return 0, err
		}
		r.in = r.buf[:encN]
	}

	return n, nil
}
