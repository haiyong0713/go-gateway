package summer_camp

import (
	"crypto/md5"
	"encoding/hex"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/bwsonline"
	"go-gateway/app/web-svr/activity/interface/dao/cost"
	"go-gateway/app/web-svr/activity/interface/dao/favorite"
	"go-gateway/app/web-svr/activity/interface/dao/reward_conf"
	"go-gateway/app/web-svr/activity/interface/dao/summer_camp"
	"go-gateway/app/web-svr/activity/interface/service/archive"
	"go-gateway/app/web-svr/activity/interface/service/like"
)

// Service ...
type Service struct {
	c             *conf.Config
	summerCampDao summer_camp.Dao
	costPointDao  cost.Dao
	rewardConfDao reward_conf.Dao
	stockDao      *bwsonline.Dao
	favDao        *favorite.Dao
	archive       *archive.Service
	likeSrv       *like.Service
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:             c,
		costPointDao:  cost.New(c),
		summerCampDao: summer_camp.New(c),
		rewardConfDao: reward_conf.New(c),
		stockDao:      bwsonline.New(c),
		favDao:        favorite.New(c),
		archive:       archive.New(c),
		likeSrv:       like.New(c),
	}
	return s
}

func (s *Service) Close() {

}

func (s *Service) md5(source string) string {
	md5Str := md5.New()
	md5Str.Write([]byte(source))
	return hex.EncodeToString(md5Str.Sum(nil))
}
