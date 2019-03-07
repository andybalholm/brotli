package brotli

import (
	"errors"
	"io"
)

// WriterOptions configures Writer.
type WriterOptions struct {
	// Quality controls the compression-speed vs compression-density trade-offs.
	// The higher the quality, the slower the compression. Range is 0 to 11.
	Quality int
	// LGWin is the base 2 logarithm of the sliding window size.
	// Range is 10 to 24. 0 indicates automatic configuration based on Quality.
	LGWin int
}

var (
	errEncode       = errors.New("brotli: encode error")
	errWriterClosed = errors.New("brotli: Writer is closed")
)

// NewWriter initializes new Writer instance.
func NewWriter(dst io.Writer, options WriterOptions) *Writer {
	w := new(Writer)
	BrotliEncoderInitState(w)
	BrotliEncoderSetParameter(w, BROTLI_PARAM_QUALITY, uint32(options.Quality))
	if options.LGWin > 0 {
		BrotliEncoderSetParameter(w, BROTLI_PARAM_LGWIN, uint32(options.LGWin))
	}
	w.dst = dst
	return w
}

func (w *Writer) writeChunk(p []byte, op int) (n int, err error) {
	if w.dst == nil {
		return 0, errWriterClosed
	}

	for {
		availableIn := uint(len(p))
		nextIn := p
		availableOut := uint(0)
		success := BrotliEncoderCompressStream(w, op, &availableIn, &nextIn, &availableOut, nil, nil)
		bytesConsumed := len(p) - int(availableIn)
		p = p[bytesConsumed:]
		n += bytesConsumed
		if !success {
			return n, errEncode
		}

		var outputDataSize uint
		outputData := BrotliEncoderTakeOutput(w, &outputDataSize)
		outputData = outputData[:outputDataSize]

		if outputDataSize > 0 {
			_, err = w.dst.Write(outputData)
			if err != nil {
				return n, err
			}
		}
		if len(p) == 0 && !BrotliEncoderHasMoreOutput(w) {
			return n, nil
		}
	}
}

// Flush outputs encoded data for all input provided to Write. The resulting
// output can be decoded to match all input before Flush, but the stream is
// not yet complete until after Close.
// Flush has a negative impact on compression.
func (w *Writer) Flush() error {
	_, err := w.writeChunk(nil, BROTLI_OPERATION_FLUSH)
	return err
}

// Close flushes remaining data to the decorated writer.
func (w *Writer) Close() error {
	// If stream is already closed, it is reported by `writeChunk`.
	_, err := w.writeChunk(nil, BROTLI_OPERATION_FINISH)
	w.dst = nil
	return err
}

// Write implements io.Writer. Flush or Close must be called to ensure that the
// encoded bytes are actually flushed to the underlying Writer.
func (w *Writer) Write(p []byte) (n int, err error) {
	return w.writeChunk(p, BROTLI_OPERATION_PROCESS)
}
