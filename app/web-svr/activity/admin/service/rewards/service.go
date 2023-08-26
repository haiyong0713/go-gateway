package rewards

import (
	"context"

	"gopkg.in/go-playground/validator.v9"

	"go-gateway/app/web-svr/activity/admin/conf"
	dao "go-gateway/app/web-svr/activity/admin/dao/rewards"
)

var Client *service

// service ...
type service struct {
	c   *conf.Config
	dao *dao.Dao
	//校验工具
	v *validator.Validate
}

// New ...
func Init(c *conf.Config) *service {
	s := &service{
		c:   c,
		dao: dao.New(c),
		v:   validator.New(),
	}
	Client = s
	return s
}

func (s *service) UploadCdKey(ctx context.Context, userName string, awardId int64, keys []string) error {
	return s.dao.UploadCdKey(ctx, userName, 1000, awardId, keys)
}

func (s *service) GetCdkeyCount(ctx context.Context, awardId int64) (int64, error) {
	return s.dao.GetCdkeyCount(ctx, awardId)
}
