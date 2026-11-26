package matchfinder

import (
	"bytes"
	"errors"
	"testing"
)

type echoEncoder struct{}

func (echoEncoder) Encode(dst, src []byte, _ []Match, lastBlock bool) []byte {
	dst = append(dst, src...)
	if lastBlock {
		dst = append(dst, '!')
	}
	return dst
}

func (echoEncoder) Reset() {}

type failOnceWriter struct {
	writes int
	failAt int
	err    error
}

func (w *failOnceWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes == w.failAt {
		return 0, w.err
	}
	return len(p), nil
}

func TestWriterCloseAndStickyError(t *testing.T) {
	for _, blockSize := range []int{0, 4} {
		t.Run("after_close", func(t *testing.T) {
			var buf bytes.Buffer
			w := &Writer{
				Dest:        &buf,
				MatchFinder: NoMatchFinder{},
				Encoder:     echoEncoder{},
				BlockSize:   blockSize,
			}
			if _, err := w.Write([]byte("abc")); err != nil {
				t.Fatalf("Write: %v", err)
			}
			if err := w.Close(); err != nil {
				t.Fatalf("Close: %v", err)
			}
			want := buf.String()

			if n, err := w.Write([]byte("def")); err == nil || n != 0 {
				t.Fatalf("Write after Close (blockSize=%d): n=%d err=%v, want 0 and error", blockSize, n, err)
			}
			if err := w.Close(); err == nil {
				t.Fatalf("second Close (blockSize=%d): got nil, want error", blockSize)
			}
			if got := buf.String(); got != want {
				t.Fatalf("closed writer mutated output: got %q, want %q", got, want)
			}

			buf.Reset()
			w.Reset(&buf)
			if _, err := w.Write([]byte("xyz")); err != nil {
				t.Fatalf("Write after Reset: %v", err)
			}
			if err := w.Close(); err != nil {
				t.Fatalf("Close after Reset: %v", err)
			}
			if got := buf.String(); got != "xyz!" {
				t.Fatalf("after Reset: got %q, want %q", got, "xyz!")
			}
		})

		t.Run("sticky_error", func(t *testing.T) {
			errBoom := errors.New("boom")
			dst := &failOnceWriter{failAt: 1, err: errBoom}
			w := &Writer{
				Dest:        dst,
				MatchFinder: NoMatchFinder{},
				Encoder:     echoEncoder{},
				BlockSize:   blockSize,
			}
			if _, err := w.Write([]byte("abcdef")); err != errBoom {
				t.Fatalf("Write (blockSize=%d): got %v, want %v", blockSize, err, errBoom)
			}
			writesAfterFail := dst.writes
			if err := w.Close(); err != errBoom {
				t.Fatalf("Close after failed Write (blockSize=%d): got %v, want %v", blockSize, err, errBoom)
			}
			if dst.writes != writesAfterFail {
				t.Fatalf("Close called Dest.Write %d extra times after error", dst.writes-writesAfterFail)
			}
			if err := w.Close(); err == nil {
				t.Fatalf("second Close after error (blockSize=%d): got nil, want error", blockSize)
			}
		})
	}
}
