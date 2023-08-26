package article

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/conf"

	articleRpc "git.bilibili.co/bapis/bapis-go/article/service"
)

// Dao .
type Dao struct {
	c                 *conf.Config
	articleHTTPClient *bm.Client
	userFeed          *conf.UserFeed
	articleRpcClient  articleRpc.ArticleGRPCClient
}

// New .
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:                 c,
		articleHTTPClient: bm.NewClient(c.HTTPClient.Read),
		userFeed:          c.UserFeed,
	}
	articleClient, err := articleRpc.NewClient(nil)
	if err != nil {
		panic("article client rpc error: " + err.Error())
	}
	d.articleRpcClient = articleClient
	return d
}
