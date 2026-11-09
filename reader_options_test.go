package brotli

import (
	"bytes"
	"encoding/hex"
	"io"
	"testing"
)

func TestReaderLargeWindowOptions(t *testing.T) {
	// Reference Brotli CLI, quality 11, windows 25 and 30. Contains backreferences.
	for _, encoded := range []string{
		"1159fd7f047e67821d3bc54a87d99742506f452497204674a990c656043d56216084fc931cf52200",
		"115efd7f047e67821d3bc54a87d99742506f452497204674a990c656043d56216084fc931cf52200",
	} {
		data, err := hex.DecodeString(encoded)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.ReadAll(NewReader(bytes.NewReader(data))); err == nil {
			t.Fatal("ordinary reader accepted a large-window stream")
		}
		r := NewReaderOptions(bytes.NewReader(data), ReaderOptions{LargeWindow: true})
		want := bytes.Repeat([]byte("Brotli PDF stream\n"), 4096)
		for _, input := range []io.Reader{bytes.NewReader(data), &oneByteReader{data: string(data)}} {
			r.Reset(input)
			got, err := io.ReadAll(r)
			if err != nil || !bytes.Equal(got, want) {
				t.Fatalf("decode: len=%d err=%v", len(got), err)
			}
		}
		// Reset must retain options even after an unrecoverable decode error.
		r.Reset(bytes.NewReader([]byte{0x11, 0x9f}))
		if _, err := io.ReadAll(r); err == nil {
			t.Fatal("invalid window accepted")
		}
		r.Reset(bytes.NewReader(data))
		got, err := io.ReadAll(r)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("reset after error: %v", err)
		}
		for i := 0; i < len(data); i++ {
			r.Reset(bytes.NewReader(data[:i]))
			if _, err := io.ReadAll(r); err == nil {
				t.Fatalf("truncation at %d accepted", i)
			}
		}
	}
}

func TestReaderLargeWindowMinimumAndOrdinaryStreams(t *testing.T) {
	minimum, _ := hex.DecodeString("110a22000242726f746c69205044462073747265616d0a03")
	r := NewReaderOptions(bytes.NewReader(minimum), ReaderOptions{LargeWindow: true})
	got, err := io.ReadAll(r)
	if err != nil || string(got) != "Brotli PDF stream\n" {
		t.Fatalf("minimum window: %q %v", got, err)
	}
	var compressed bytes.Buffer
	w := NewWriter(&compressed)
	if _, err := w.Write([]byte("ordinary")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	r.Reset(bytes.NewReader(compressed.Bytes()))
	got, err = io.ReadAll(r)
	if err != nil || string(got) != "ordinary" {
		t.Fatalf("ordinary: %q %v", got, err)
	}
	for _, invalid := range [][]byte{{0x11, 0x09}, {0x11, 0x1f}, {0x91, 0x19}} {
		r.Reset(bytes.NewReader(invalid))
		if _, err := io.ReadAll(r); err == nil {
			t.Fatalf("invalid header accepted: %x", invalid)
		}
	}
}
