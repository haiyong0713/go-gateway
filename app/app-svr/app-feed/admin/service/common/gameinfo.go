package common

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-feed/admin/model/game"

	eg "go-common/library/sync/errgroup.v2"
)

const (
	_androidPlat = 1
	_iosPlat     = 2
)

func (s *Service) Game(c context.Context, id int64, platForm int) (res *game.Info, err error) {
	var (
		androidGame, iosGame *game.Info
		androidErr, iosErr   error
	)
	if platForm != 0 {
		return s.GameDao.GameInfoApp(c, id, platForm)
	}
	eg := eg.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		if androidGame, androidErr = s.GameDao.GameInfoApp(ctx, id, _androidPlat); androidErr != nil {
			return nil
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if iosGame, iosErr = s.GameDao.GameInfoApp(ctx, id, _iosPlat); iosErr != nil {
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
