package common

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/game"

	eg "go-common/library/sync/errgroup.v2"
)

const (
	_android = 1
	_ios     = 2
)

// AppGameInfo .
func (s *Service) AppGameInfo(c context.Context, id int64) (res *game.Info, err error) {
	var (
		androidGame, iosGame *game.Info
		androidErr, iosErr   error
	)
	eg := eg.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		if androidGame, androidErr = s.GameDao.GameInfo(ctx, id, _android); androidErr != nil {
			log.Error("AppGameInfo id(%d),error(%v)", id, androidErr)
			return nil
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if iosGame, iosErr = s.GameDao.GameInfo(ctx, id, _ios); iosErr != nil {
			log.Error("AppGameInfo id(%d),error(%v)", id, iosErr)
			return nil
		}
		return nil
	})
	//nolint:errcheck
	eg.Wait()
	if androidErr == nil && androidGame != nil {
		return androidGame, nil
	}
	if iosErr == nil && iosGame != nil {
		return iosGame, nil
	}
	if androidErr != nil {
		return nil, androidErr
	}
	if iosErr != nil {
		return nil, iosErr
	}
	return nil, fmt.Errorf("找不到游戏数据!")
}
