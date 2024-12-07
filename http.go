package brotli

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HTTPCompressor chooses a compression method (brotli, gzip, or none) based on
// the Accept-Encoding header, sets the Content-Encoding header, and returns a
// WriteCloser that implements that compression. The Close method must be called
// before the current HTTP handler returns.
func HTTPCompressor(w http.ResponseWriter, r *http.Request) io.WriteCloser {
	return HTTPCompressorWithLevels(w, r, DefaultCompression, gzip.DefaultCompression)
}

// Like HTTPCompressor, but allows customization of the compression levels for
// brotli and gzip. This will panic if you pass in an invalid compression level.
func HTTPCompressorWithLevels(w http.ResponseWriter, r *http.Request, brotliLevel, gzipLevel int) io.WriteCloser {
	return HTTPCompressorWithCustom(w, r, func(w http.ResponseWriter) io.WriteCloser {
		return NewWriterV2(w, brotliLevel)
	}, func(w http.ResponseWriter) io.WriteCloser {
		writer, err := gzip.NewWriterLevel(w, gzipLevel)
		if err != nil {
			panic(fmt.Sprintf("could not create gzip writer with compression level %d: %s", gzipLevel, err))
		}
		return writer
	})
}

// Like HTTPCompressor, but allows you to specify factories to create a custom
// brotli compressor or custom gzip compressor as appropriate. Use this if you
// need to set custom options on compressors.
func HTTPCompressorWithCustom(w http.ResponseWriter, r *http.Request, brotliFactory, gzipFactory func(http.ResponseWriter) io.WriteCloser) io.WriteCloser {
	if w.Header().Get("Vary") == "" {
		w.Header().Set("Vary", "Accept-Encoding")
	}

	encoding := negotiateContentEncoding(r, []string{"br", "gzip"})
	switch encoding {
	case "br":
		w.Header().Set("Content-Encoding", "br")
		return brotliFactory(w)
	case "gzip":
		w.Header().Set("Content-Encoding", "gzip")
		return gzipFactory(w)
	}
	return nopCloser{w}
}

// negotiateContentEncoding returns the best offered content encoding for the
// request's Accept-Encoding header. If two offers match with equal weight and
// then the offer earlier in the list is preferred. If no offers are
// acceptable, then "" is returned.
func negotiateContentEncoding(r *http.Request, offers []string) string {
	bestOffer := "identity"
	bestQ := -1.0
	specs := parseAccept(r.Header, "Accept-Encoding")
	for _, offer := range offers {
		for _, spec := range specs {
			if spec.Q > bestQ &&
				(spec.Value == "*" || spec.Value == offer) {
				bestQ = spec.Q
				bestOffer = offer
			}
		}
	}
	if bestQ == 0 {
		bestOffer = ""
	}
	return bestOffer
}

// acceptSpec describes an Accept* header.
type acceptSpec struct {
	Value string
	Q     float64
}

// parseAccept parses Accept* headers.
func parseAccept(header http.Header, key string) (specs []acceptSpec) {
loop:
	for _, s := range header[key] {
		for {
			var spec acceptSpec
			spec.Value, s = expectTokenSlash(s)
			if spec.Value == "" {
				continue loop
			}
			spec.Q = 1.0
			s = skipSpace(s)
			if strings.HasPrefix(s, ";") {
				s = skipSpace(s[1:])
				if !strings.HasPrefix(s, "q=") {
					continue loop
				}
				spec.Q, s = expectQuality(s[2:])
				if spec.Q < 0.0 {
					continue loop
				}
			}
			specs = append(specs, spec)
			s = skipSpace(s)
			if !strings.HasPrefix(s, ",") {
				continue loop
			}
			s = skipSpace(s[1:])
		}
	}
	return
}

func skipSpace(s string) (rest string) {
	i := 0
	for ; i < len(s); i++ {
		if octetTypes[s[i]]&isSpace == 0 {
			break
		}
	}
	return s[i:]
}

func expectTokenSlash(s string) (token, rest string) {
	i := 0
	for ; i < len(s); i++ {
		b := s[i]
		if (octetTypes[b]&isToken == 0) && b != '/' {
			break
		}
	}
	return s[:i], s[i:]
}

func expectQuality(s string) (q float64, rest string) {
	switch {
	case len(s) == 0:
		return -1, ""
	case s[0] == '0':
		q = 0
	case s[0] == '1':
		q = 1
	default:
		return -1, ""
	}
	s = s[1:]
	if !strings.HasPrefix(s, ".") {
		return q, s
	}
	s = s[1:]
	i := 0
	n := 0
	d := 1
	for ; i < len(s); i++ {
		b := s[i]
		if b < '0' || b > '9' {
			break
		}
		n = n*10 + int(b) - '0'
		d *= 10
	}
	return q + float64(n)/float64(d), s[i:]
}

// Octet types from RFC 2616.
var octetTypes [256]octetType

type octetType byte

const (
	isToken octetType = 1 << iota
	isSpace
)

func init() {
	// OCTET      = <any 8-bit sequence of data>
	// CHAR       = <any US-ASCII character (octets 0 - 127)>
	// CTL        = <any US-ASCII control character (octets 0 - 31) and DEL (127)>
	// CR         = <US-ASCII CR, carriage return (13)>
	// LF         = <US-ASCII LF, linefeed (10)>
	// SP         = <US-ASCII SP, space (32)>
	// HT         = <US-ASCII HT, horizontal-tab (9)>
	// <">        = <US-ASCII double-quote mark (34)>
	// CRLF       = CR LF
	// LWS        = [CRLF] 1*( SP | HT )
	// TEXT       = <any OCTET except CTLs, but including LWS>
	// separators = "(" | ")" | "<" | ">" | "@" | "," | ";" | ":" | "\" | <">
	//              | "/" | "[" | "]" | "?" | "=" | "{" | "}" | SP | HT
	// token      = 1*<any CHAR except CTLs or separators>
	// qdtext     = <any TEXT except <">>

	for c := 0; c < 256; c++ {
		var t octetType
		isCtl := c <= 31 || c == 127
		isChar := 0 <= c && c <= 127
		isSeparator := strings.ContainsRune(" \t\"(),/:;<=>?@[]\\{}", rune(c))
		if strings.ContainsRune(" \t\r\n", rune(c)) {
			t |= isSpace
		}
		if isChar && !isCtl && !isSeparator {
			t |= isToken
		}
		octetTypes[c] = t
	}
}
