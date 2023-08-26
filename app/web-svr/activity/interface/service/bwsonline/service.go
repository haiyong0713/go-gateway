package bwsonline

import (
	"crypto/md5"
	"encoding/hex"

	http "go-common/library/net/http/blademaster"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/bws"
	"go-gateway/app/web-svr/activity/interface/dao/bwsonline"
	"go-gateway/app/web-svr/activity/interface/dao/like"
	"go-gateway/app/web-svr/activity/interface/dao/lottery"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
)

// Service struct
type Service struct {
	c          *conf.Config
	accClient  accapi.AccountClient
	dao        *bwsonline.Dao
	bwsdao     *bws.Dao
	likeDao    *like.Dao
	lottDao    *lottery.Dao
	cache      *fanout.Fanout
	httpClient *http.Client
}

// New Service
func New(c *conf.Config) *Service {
	s := &Service{
		c:          c,
		dao:        bwsonline.New(c),
		likeDao:    like.New(c),
		bwsdao:     bws.New(c),
		lottDao:    lottery.New(c),
		cache:      fanout.New("cache", fanout.Worker(1), fanout.Buffer(1024)),
		httpClient: http.NewClient(c.HTTPClient),
	}
	var err error
	if s.accClient, err = accapi.NewClient(s.c.AccClient); err != nil {
		panic(err)
	}
	return s
}

func filterIDs(ids []int64) []int64 {
	idMap := make(map[int64]struct{})
	var res []int64
	for _, id := range ids {
		if _, ok := idMap[id]; ok || id <= 0 {
			continue
		}
		idMap[id] = struct{}{}
		res = append(res, id)
	}
	return res
}

func (s *Service) md5(source string) string {
	md5Str := md5.New()
	md5Str.Write([]byte(source))
	return hex.EncodeToString(md5Str.Sum(nil))
}

func (s *Service) DefaultBid() int64 {
	return s.c.BwsOnline.DefaultBid
}
