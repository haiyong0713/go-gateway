package binlog

import (
	"context"
	"fmt"

	"go-gateway/app/web-svr/native-page/job/internal/model"
)

//go:generate kratos tool redisgen
type _redis interface {
	// redis: -key=whiteListByMidKey -expire=d.cfg.WhiteListByMidExpire -check_null_code=$!=nil&&$.ID==-1 -null_expire=d.cfg.WhiteListByMidNullExpire
	AddCacheWhiteListByMid(c context.Context, mid int64, data *model.WhiteList) error
	// redis: -key=whiteListByMidKey
	DelCacheWhiteListByMid(c context.Context, mid int64) error
	// redis: -key=userSpaceByMidKey
	DelCacheUserSpaceByMid(c context.Context, mid int64) error
}

func whiteListByMidKey(mid int64) string {
	return fmt.Sprintf("white_list_mid_%d", mid)
}

func userSpaceByMidKey(mid int64) string {
	return fmt.Sprintf("nat_user_space_%d", mid)
}
