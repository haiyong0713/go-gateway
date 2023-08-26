package client

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	sdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client/metadata"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/corehandlers"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"
)

type mockCloser struct {
	closed bool
}

func (closer *mockCloser) Read(b []byte) (int, error) {
	return 0, io.EOF
}

func (closer *mockCloser) Close() error {
	closer.closed = true
	return nil
}

func TestTeeReaderCloser(t *testing.T) {
	expected := "TEST"
	buf := bytes.NewBuffer([]byte(expected))
	lw := bytes.NewBuffer(nil)
	c := &mockCloser{}
	closer := teeReaderCloser{
		io.TeeReader(buf, lw),
		c,
	}

	b := make([]byte, len(expected))
	_, err := closer.Read(b)
	closer.Close()
	if expected != lw.String() {
		t.Errorf("Expected %q, but received %q", expected, lw.String())
	}
	if err != nil {
		t.Errorf("Expected 'nil', but received %v", err)
	}
	if !c.closed {
		t.Error("Expected 'true', but received 'false'")
	}
}

func TestLogRequest(t *testing.T) {
	cases := []struct {
		Body       io.ReadSeeker
		ExpectBody []byte
	}{
		{
			Body:       sdk.ReadSeekCloser(bytes.NewBuffer([]byte("body content"))),
			ExpectBody: []byte("body content"),
		},
		{
			Body:       bytes.NewReader([]byte("body content")),
			ExpectBody: []byte("body content"),
		},
	}

	for i, c := range cases {
		req := request.New(
			sdk.Config{},
			metadata.ClientInfo{
				Endpoint: "https://www.bilibili.com",
			},
			testHandlers(),
			nil,
			&request.Operation{
				Name:       "bilibili",
				HTTPMethod: "GET",
				HTTPPath:   "/",
			},
			struct{}{}, nil,
		)
		req.SetReaderBody(c.Body)
		req.Build()

		logRequest(req)

		b, err := ioutil.ReadAll(req.HTTPRequest.Body)
		if err != nil {
			t.Fatalf("%d, expect to read SDK request Body", i)
		}

		if e, a := c.ExpectBody, b; !reflect.DeepEqual(e, a) {
			t.Errorf("%d, expect %v body, got %v", i, e, a)
		}
	}
}

func TestLogResponse(t *testing.T) {
	cases := []struct {
		Body       *bytes.Buffer
		ExpectBody []byte
		ReadBody   bool
	}{
		{
			Body:       bytes.NewBuffer([]byte("body content")),
			ExpectBody: []byte("body content"),
		},
		{
			Body:       bytes.NewBuffer([]byte("body content")),
			ReadBody:   true,
			ExpectBody: []byte("body content"),
		},
	}

	for i, c := range cases {
		req := request.New(
			sdk.Config{},
			metadata.ClientInfo{
				Endpoint: "https://www.bilibili.com",
			},
			testHandlers(),
			nil,
			&request.Operation{
				Name:       "bilibili",
				HTTPMethod: "GET",
				HTTPPath:   "/",
			},
			struct{}{}, nil,
		)
		req.HTTPResponse = &http.Response{
			StatusCode: 200,
			Status:     "OK",
			Header: http.Header{
				"TEST": []string{"hello,world"},
			},
			Body: ioutil.NopCloser(c.Body),
		}

		logResponse(req)
		req.Handlers.Unmarshal.Run(req)
		if c.ReadBody {
			if e, a := len(c.ExpectBody), c.Body.Len(); e != a {
				t.Errorf("%d, expect original body not to of been read", i)
			}
		}

		b, err := ioutil.ReadAll(req.HTTPResponse.Body)
		if err != nil {
			t.Fatalf("%d, expect to read SDK request Body", i)
		}
		if e, a := c.ExpectBody, b; !bytes.Equal(e, a) {
			t.Errorf("%d, expect %v body, got %v", i, e, a)
		}
	}
}

func testHandlers() request.Handlers {
	var handlers request.Handlers

	handlers.Build.PushBackNamed(corehandlers.SDKVersionUserAgentHandler)

	return handlers
}
