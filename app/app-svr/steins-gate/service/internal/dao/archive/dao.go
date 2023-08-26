package archive

import (
	bm "go-common/library/net/http/blademaster"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	playurlgrpc "go-gateway/app/app-svr/playurl/service/api"
	"go-gateway/app/app-svr/steins-gate/service/conf"
)

const (
	_videoUpViewURI = "/videoup/view"
)

// Dao dao.
type Dao struct {
	c               *conf.Config
	arcClient       arcgrpc.ArchiveClient
	httpVideoClient *bm.Client
	playurlClient   playurlgrpc.PlayURLClient
	videoUpViewURL  string
}

// New new a dao and return.
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c:               c,
		videoUpViewURL:  c.Host.Videoup + _videoUpViewURI,
		httpVideoClient: bm.NewClient(c.VideoClient),
	}
	var err error
	if dao.arcClient, err = arcgrpc.NewClient(c.Archive); err != nil {
		panic(err)
	}
	if dao.playurlClient, err = playurlgrpc.NewClient(c.Playurl); err != nil {
		panic(err)
	}
	return

}
