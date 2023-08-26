package push

import (
	"context"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	dao "go-gateway/app/app-svr/app-feed/admin/dao/show"
	"go-gateway/app/app-svr/app-feed/admin/model/push"
	"go-gateway/app/app-svr/app-feed/admin/util"

	"go-common/library/log"

	"github.com/pkg/errors"
)

// Service is
type Service struct {
	dao *dao.Dao
}

// New is
func New(c *conf.Config) *Service {
	s := &Service{
		dao: dao.New(c),
	}
	return s
}

func (s *Service) PushList(ctx context.Context) ([]*push.PushDetail, error) {
	pushlist, err := s.dao.PushList(ctx)
	if err != nil {
		log.Error("s.dao.PushList error(%+v)", err)
		return nil, err
	}
	return pushlist, nil
}

func (s *Service) PushDetail(ctx context.Context, id int64) (*push.PushDetail, error) {
	pushDetail, err := s.dao.PushDetail(ctx, id)
	if err != nil {
		log.Error("s.dao.PushDetail error(%+v), id(%d)", err, id)
		return nil, err
	}
	return pushDetail, nil
}

func (s *Service) PushSave(ctx context.Context, detail *push.PushDetail, username string, uid int64) error {
	var (
		action string
		err    error
	)
	if action, err = func() (string, error) {
		if detail.IsUpdateOp() {
			if err := s.dao.PushUpdate(ctx, detail); err != nil {
				return "", errors.Wrapf(err, "s.dao.PushUpdate error, detail(%v)", detail)
			}
			return "update", nil
		}
		if err := s.dao.PushCreate(ctx, detail); err != nil {
			return "", errors.Wrapf(err, "s.dao.PushCreate error, detail(%v)", detail)
		}
		return "create", nil
	}(); err != nil {
		log.Error("PushSave error(%+v)", err)
		return err
	}
	if err = util.AddPackagePushLog(username, uid, detail.ID, action, 0, nil, detail); err != nil {
		//just log
		log.Error("PushSave util.AddPackagePushLog error(%+v), action(%s), uname(%s)", err, action, username)
	}
	return nil
}

func (s *Service) PushDelete(ctx context.Context, id, uid int64, username string) error {
	if err := s.dao.PushDelete(ctx, id); err != nil {
		log.Error("s.dao.PushDelete error(%+v)", err)
		return err
	}
	if err := util.AddPackagePushLog(username, uid, id, "delete", 0, nil, id); err != nil {
		//just log
		log.Error("PushDelete util.AddPackagePushLog error(%+v), action(%s), uname(%s)", err, "delete", username)
	}
	return nil
}
