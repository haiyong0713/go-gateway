package debug

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type neverEnding byte

func (b neverEnding) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(b)
	}
	return len(p), nil
}

// delegateReader is a reader that delegates to another reader,
// once it arrives on a channel.
type delegateReader struct {
	c chan io.Reader
	r io.Reader // nil until received from c
}

func (r *delegateReader) Read(p []byte) (int, error) {
	if r.r == nil {
		r.r = <-r.c
	}
	return r.r.Read(p)
}

// dumpConn is a net.Conn which writes to Writer and reads from Reader
type dumpConn struct {
	io.Writer
	io.Reader
}

func (c *dumpConn) Close() error                       { return nil }
func (c *dumpConn) LocalAddr() net.Addr                { return nil }
func (c *dumpConn) RemoteAddr() net.Addr               { return nil }
func (c *dumpConn) SetDeadline(t time.Time) error      { return nil }
func (c *dumpConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *dumpConn) SetWriteDeadline(t time.Time) error { return nil }

func DumpRequestOut(req *http.Request, body bool) ([]byte, error) {
	save := req.Body
	dummyBody := false
	if !body || req.Body == nil {
		req.Body = nil
		if req.ContentLength != 0 {
			req.Body = ioutil.NopCloser(io.LimitReader(neverEnding('x'), req.ContentLength))
			dummyBody = true
		}
	} else {
		var err error
		save, req.Body, err = drainBody(req.Body)
		if err != nil {
			return nil, err
		}
	}

	// Since we're using the actual Transport code to write the request,
	// switch to http so the Transport doesn't try to do an SSL
	// negotiation with our dumpConn and its bytes.Buffer & pipe.
	// The wire format for https and http are the same, anyway.
	reqSend := req
	if req.URL.Scheme == "https" || req.URL.Scheme == "discovery" {
		reqSend = new(http.Request)
		*reqSend = *req
		reqSend.URL = new(url.URL)
		*reqSend.URL = *req.URL
		reqSend.URL.Scheme = "http"
	}

	// Use the actual Transport code to record what we would send
	// on the wire, but not using TCP.  Use a Transport with a
	// custom dialer that returns a fake net.Conn that waits
	// for the full input (and recording it), and then responds
	// with a dummy response.
	var buf bytes.Buffer // records the output
	pr, pw := io.Pipe()
	defer pr.Close()
	defer pw.Close()
	dr := &delegateReader{c: make(chan io.Reader)}

	t := &http.Transport{
		Dial: func(net, addr string) (net.Conn, error) {
			return &dumpConn{io.MultiWriter(&buf, pw), dr}, nil
		},
	}
	defer t.CloseIdleConnections()

	// Wait for the request before replying with a dummy response:
	//nolint:biligowordcheck
	go func() {
		req, err := http.ReadRequest(bufio.NewReader(pr))
		if err == nil {
			// Ensure all the body is read; otherwise
			// we'll get a partial dump.
			//nolint:errcheck
			io.Copy(ioutil.Discard, req.Body)
			req.Body.Close()
		}
		dr.c <- strings.NewReader("HTTP/1.1 204 No Content\r\nConnection: close\r\n\r\n")
	}()

	//nolint
	_, err := t.RoundTrip(reqSend)

	req.Body = save
	if err != nil {
		return nil, err
	}
	dump := buf.Bytes()

	// If we used a dummy body above, remove it now.
	// TODO: if the req.ContentLength is large, we allocate memory
	// unnecessarily just to slice it off here. But this is just
	// a debug function, so this is acceptable for now. We could
	// discard the body earlier if this matters.
	if dummyBody {
		if i := bytes.Index(dump, []byte("\r\n\r\n")); i >= 0 {
			dump = dump[:i+4]
		}
	}
	return dump, nil
}

func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
