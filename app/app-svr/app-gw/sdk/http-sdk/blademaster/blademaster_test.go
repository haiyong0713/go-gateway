package blademaster

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"

	"github.com/stretchr/testify/assert"
)

func RequestOperation(req *http.Request) *request.Operation {
	op := &request.Operation{
		Name:       fmt.Sprintf("proxy-for:%s", req.URL.Path),
		HTTPMethod: req.Method,
		HTTPPath:   req.URL.EscapedPath(),
	}
	return op
}

func TestRequestOperation(t *testing.T) {
	req1, _ := url.Parse("http://api.bilibili.com/x/web-interface/nav")
	req2, _ := url.Parse("http://api.bilibili.com/x/web-interface>nav")
	req3, _ := url.Parse("http://api.bilibili.com/x/web-interface<nav")

	op1 := RequestOperation(&http.Request{Method: "GET", URL: req1})
	assert.Equal(t, "/x/web-interface/nav", op1.HTTPPath)

	op2 := RequestOperation(&http.Request{Method: "GET", URL: req2})
	assert.NotEqual(t, "/x/web-interface>nav", op2.HTTPPath)

	op3 := RequestOperation(&http.Request{Method: "GET", URL: req3})
	assert.NotEqual(t, "/x/web-interface<nav", op3.HTTPPath)
}

type testMatch struct {
	testURL  *url.URL
	priority int
}

func (t testMatch) String() string {
	return t.testURL.String()
}

func (t testMatch) Match(url *url.URL) bool {
	if t.testURL != nil {
		return true
	}
	return false
}

func (t testMatch) Len() int {
	return len(t.String())
}

func (t testMatch) Priority() int {
	return t.priority
}

func TestMatchLongestPath(t *testing.T) {
	u, _ := url.Parse("http://test.bilibili.com")
	u1, _ := url.Parse("http://api.bilibili.com")
	u2, _ := url.Parse("discovery://web.interface")
	u3, _ := url.Parse("http://api.com")
	u4, _ := url.Parse("discovery://web.interface.xxxxxx")
	m1 := &testMatch{
		testURL:  u1,
		priority: 1,
	}
	m2 := &testMatch{
		testURL:  u2,
		priority: 2,
	}
	m3 := &testMatch{
		testURL:  u3,
		priority: 2,
	}
	m4 := &testMatch{
		testURL:  u4,
		priority: 1,
	}
	p1 := &PathMeta{
		matcher: m1,
	}
	p2 := &PathMeta{
		matcher: m2,
	}
	p3 := &PathMeta{
		matcher: m3,
	}
	p4 := &PathMeta{
		matcher: m4,
	}
	pp1 := &ProxyPass{
		cfg: Config{
			DynPath: []*PathMeta{p1, p2},
		},
	}
	path1, _ := pp1.MatchLongestPath(u)
	assert.Equal(t, path1.matcher.String(), u1.String())

	pp2 := &ProxyPass{
		cfg: Config{
			DynPath: []*PathMeta{p2, p3},
		},
	}
	path2, _ := pp2.MatchLongestPath(u)
	assert.Equal(t, path2.matcher.String(), u2.String())

	pp3 := &ProxyPass{
		cfg: Config{
			DynPath: []*PathMeta{p1, p2, p3},
		},
	}
	path3, _ := pp3.MatchLongestPath(u)
	assert.Equal(t, path3.matcher.String(), u1.String())

	pp4 := &ProxyPass{
		cfg: Config{
			DynPath: []*PathMeta{p1, p2, p3, p4},
		},
	}
	path4, _ := pp4.MatchLongestPath(u)
	assert.Equal(t, path4.matcher.String(), u4.String())
}

func TestMatcher(t *testing.T) {
	u1, _ := url.Parse("http://api.bilibili.com/path")
	e := &exactlyMatcher{
		path: "/path",
	}
	p1 := &prefixMatcher{
		prefix: "/p",
	}
	p2 := &prefixMatcher{
		prefix: "/pa",
	}
	r, _ := createRegexMatcher("/pat")
	pm1 := &PathMeta{
		matcher: e,
	}
	pm2 := &PathMeta{
		matcher: p1,
	}
	pm3 := &PathMeta{
		matcher: p2,
	}
	pm4 := &PathMeta{
		matcher: r,
	}

	pp1 := &ProxyPass{
		cfg: Config{
			DynPath: []*PathMeta{pm1, pm2},
		},
	}
	path1, _ := pp1.MatchLongestPath(u1)
	assert.Equal(t, path1.matcher.String(), "= /path")

	pp2 := &ProxyPass{
		cfg: Config{
			DynPath: []*PathMeta{pm2, pm3},
		},
	}
	path2, _ := pp2.MatchLongestPath(u1)
	assert.Equal(t, path2.matcher.String(), "/pa")

	pp3 := &ProxyPass{
		cfg: Config{
			DynPath: []*PathMeta{pm2, pm4},
		},
	}
	path3, _ := pp3.MatchLongestPath(u1)
	assert.Equal(t, path3.matcher.String(), "/p")

	c1 := &PathMeta{
		Pattern: "~ ^/path",
	}
	c1.InitStatic()
	c2 := &PathMeta{
		Pattern: "~ /path",
	}
	c2.InitStatic()
	c3 := &PathMeta{
		Pattern: "= /path",
	}
	c3.InitStatic()
	pp4 := &ProxyPass{
		cfg: Config{
			DynPath: []*PathMeta{c1, c2, c3},
		},
	}
	path4, _ := pp4.MatchLongestPath(u1)
	assert.Equal(t, path4.matcher.String(), "= /path")

	pp5 := &ProxyPass{
		cfg: Config{
			DynPath: []*PathMeta{c1, c2},
		},
	}
	path5, _ := pp5.MatchLongestPath(u1)
	assert.Equal(t, path5.matcher.String(), "^/path")

	u2, _ := url.Parse("http://api.bilibili.com/path/index")
	path6, _ := pp5.MatchLongestPath(u2)
	assert.Equal(t, path6.matcher.String(), "^/path")

	u3, _ := url.Parse("http://api.bilibili.com/index/path")
	path7, _ := pp5.MatchLongestPath(u3)
	assert.Equal(t, path7.matcher.String(), "/path")
}
