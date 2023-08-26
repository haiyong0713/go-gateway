package region

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/interface/conf"
	tagdao "go-gateway/app/app-svr/app-feed/interface/dao/tag"
)

type Service struct {
	c *conf.Config
	// dao
	tg *tagdao.Dao
}

// New a region service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:  c,
		tg: tagdao.New(c),
	}
	return
}

func (s *Service) md5(v interface{}) string {
	bs, err := json.Marshal(v)
	if err != nil {
		log.Error("json.Marshal error(%v)", err)
		return "region_version"
	}
	hs := md5.Sum(bs)
	return hex.EncodeToString(hs[:])
}
