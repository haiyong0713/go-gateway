package tag

import (
	"go-gateway/app/web-svr/activity/admin/conf"

	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

// Service struct
type Service struct {
	c      *conf.Config
	tagRPC tagrpc.TagRPCClient
}

// Close service
func (s *Service) Close() {

}

// New Service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c: c,
	}
	var err error

	if s.tagRPC, err = tagrpc.NewClient(c.TagClient); err != nil {
		panic(err)
	}

	return
}
