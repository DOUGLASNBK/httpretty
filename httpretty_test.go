package httpretty

import (
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestPrintRequest(t *testing.T) {
	t.Parallel()

	var req, err = http.NewRequest(http.MethodPost, "http://wxww.example.com/", nil)

	if err != nil {
		panic(err)
	}

	logger := &Logger{
		TLS:            true,
		RequestHeader:  true,
		RequestBody:    true,
		ResponseHeader: true,
		ResponseBody:   true,
	}

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.PrintRequest(req)

	want := `> POST / HTTP/1.1
> Host: wxww.example.com

`

	if got := buf.String(); got != want {
		t.Errorf("PrintRequest(req) = %v, wanted %v", got, want)
	}
}

func TestPrintRequestWithColors(t *testing.T) {
	t.Parallel()

	var req, err = http.NewRequest(http.MethodPost, "http://wxww.example.com/", nil)

	if err != nil {
		panic(err)
	}

	logger := &Logger{
		TLS:            true,
		RequestHeader:  true,
		RequestBody:    true,
		ResponseHeader: true,
		ResponseBody:   true,
		Colors:         true,
	}

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.PrintRequest(req)

	want := "> \x1b[34;1mPOST\x1b[0m \x1b[33m/\x1b[0m \x1b[34mHTTP/1.1\x1b[0m" +
		"\n> \x1b[34;1mHost\x1b[0m\x1b[31m:\x1b[0m \x1b[33mwxww.example.com\x1b[0m\n\n"

	if got := buf.String(); got != want {
		t.Errorf("PrintRequest(req) = %v, wanted %v", got, want)
	}
}

func TestPrintRequestFiltered(t *testing.T) {
	t.Parallel()

	var req, err = http.NewRequest(http.MethodPost, "http://wxww.example.com/", nil)

	if err != nil {
		panic(err)
	}

	logger := &Logger{
		TLS:            true,
		RequestHeader:  true,
		RequestBody:    true,
		ResponseHeader: true,
		ResponseBody:   true,
	}

	var buf bytes.Buffer
	logger.SetOutput(&buf)
	logger.SetFilter(func(req *http.Request) (skip bool, err error) {
		return true, nil
	})

	logger.PrintRequest(req)

	if got := buf.Len(); got != 0 {
		t.Errorf("got %v from logger, wanted nothing (everything should be filtered)", got)
	}
}

func TestPrintRequestNil(t *testing.T) {
	t.Parallel()

	logger := &Logger{
		TLS:            true,
		RequestHeader:  true,
		RequestBody:    true,
		ResponseHeader: true,
		ResponseBody:   true,
	}

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.PrintRequest(nil)

	want := "> error: null request\n"

	if got := buf.String(); got != want {
		t.Errorf("PrintRequest(req) = %v, wanted %v", got, want)
	}
}

func TestPrintResponseNil(t *testing.T) {
	t.Parallel()

	logger := &Logger{
		TLS:            true,
		RequestHeader:  true,
		RequestBody:    true,
		ResponseHeader: true,
		ResponseBody:   true,
	}

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.PrintResponse(nil)

	want := "< error: null response\n"

	if got := buf.String(); got != want {
		t.Errorf("PrintResponse(req) = %v, wanted %v", got, want)
	}
}

func testBody(t *testing.T, r io.Reader, want []byte) {
	t.Helper()

	got, err := ioutil.ReadAll(r)

	if err != nil {
		t.Errorf("expected no error reading response body, got %v instead", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf(`got body = %v, wanted %v`, string(got), string(want))
	}
}

func TestJSONFormatterWriterError(t *testing.T) {
	// verifies if function doesn't panic if passed writer isn't *bytes.Buffer
	f := &JSONFormatter{}
	want := "underlying writer for JSONFormatter must be *bytes.Buffer"
	if err := f.Format(os.Stdout, []byte(`{}`)); err == nil || err.Error() != want {
		t.Errorf("got format error = %v, wanted %v", err, want)
	}
}

// newTransport creates a new HTTP Transport.
//
// BUG(henvic): this function is mostly used at this moment because of a data race condition on the standard library.
// See https://github.com/golang/go/issues/30597 for details.
func newTransport() *http.Transport {
	// values copied from Go 1.13.7 http.DefaultTransport variable.
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
