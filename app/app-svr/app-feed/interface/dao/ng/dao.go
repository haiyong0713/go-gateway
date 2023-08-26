package ng

import (
	"net/http"
	"net/url"

	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-gateway/app/app-svr/app-feed/interface/conf"
	"go-gateway/app/app-svr/app-feed/interface/model/ng"

	"github.com/pkg/errors"
)

// Dao is a dao.
type Dao struct {
	// http client
	client *bm.Client
}

// New new a dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// http client
		client: bm.NewClient(c.HTTPNg, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
	}
	return
}

func (d *Dao) ToNgDispatch(ctx *bm.Context) (*ng.ToNgDispatchReply, error) {
	request := ctx.Request
	ngIndexURL, err := url.Parse("discovery://main.app-svr.app-feed-ng/x/v2/feed-ng/index")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ngIndexURL.RawQuery = request.URL.Query().Encode()
	clientReq, err := http.NewRequest(request.Method, ngIndexURL.String(), nil)
	if err != nil {
		return nil, err
	}
	copyRequestHeader(clientReq, request)
	resp, body, err := d.client.RawResponse(ctx, clientReq)
	if err != nil {
		return nil, err
	}
	header := make(map[string]ng.Header)
	for k, v := range resp.Header {
		val := ng.Header{Values: v}
		header[k] = val
	}
	reply := &ng.ToNgDispatchReply{
		Response:   body,
		StatusCode: int32(resp.StatusCode),
		Header:     header,
	}
	return reply, nil
}

func copyRequestHeader(src, dst *http.Request) {
	for k, v := range src.Header {
		dst.Header[k] = v
	}
	dst.Host = src.Host // host header is seperated in field
}
