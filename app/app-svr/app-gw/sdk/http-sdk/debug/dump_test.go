package debug

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDumpRequestOut(t *testing.T) {
	body := url.Values{}
	body.Set("test", "test")
	// Case 1 : http GET
	req1, _ := http.NewRequest(http.MethodGet, "http://api.bilibili.com/x/v2/view?test=test", nil)
	p1, err := DumpRequestOut(req1, true)
	assert.Nil(t, err)
	assert.Equal(t, string(p1), "GET /x/v2/view?test=test HTTP/1.1\r\nHost: api.bilibili.com\r\nUser-Agent: Go-http-client/1.1\r\nAccept-Encoding: gzip\r\n\r\n")
	// Case 2: discovery GET
	req2, _ := http.NewRequest(http.MethodGet, "discovery://main.app-svr.app-view/x/v2/view", strings.NewReader(body.Encode()))
	p2, err := DumpRequestOut(req2, true)
	assert.Nil(t, err)
	assert.Equal(t, string(p2), "GET /x/v2/view HTTP/1.1\r\nHost: main.app-svr.app-view\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 9\r\nAccept-Encoding: gzip\r\n\r\ntest=test")
	// Case 3: http POST
	req3, _ := http.NewRequest(http.MethodPost, "http://api.bilibili.com/x/v2/view", strings.NewReader(body.Encode()))
	p3, err := DumpRequestOut(req3, true)
	assert.Nil(t, err)
	assert.Equal(t, string(p3), "POST /x/v2/view HTTP/1.1\r\nHost: api.bilibili.com\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 9\r\nAccept-Encoding: gzip\r\n\r\ntest=test")
	// Case 3: discovery POST
	req4, _ := http.NewRequest(http.MethodPost, "discovery://main.app-svr.app-view/x/v2/view", strings.NewReader(body.Encode()))
	p4, err := DumpRequestOut(req4, true)
	assert.Nil(t, err)
	assert.Equal(t, string(p4), "POST /x/v2/view HTTP/1.1\r\nHost: main.app-svr.app-view\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 9\r\nAccept-Encoding: gzip\r\n\r\ntest=test")
}
