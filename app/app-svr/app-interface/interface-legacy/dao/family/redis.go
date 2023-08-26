package family

import (
	"context"
	"fmt"
)

func qrcodeKey(ticket string) string {
	return fmt.Sprintf("fyqr_%+v", ticket)
}

func qrcodeBindKey(ticket string) string {
	return fmt.Sprintf("fyqr_bind_%+v", ticket)
}

func timelockPwdKey(mid int64) string {
	return fmt.Sprintf("fy_tl_%d", mid)
}

//go:generate kratos tool redisgen
type _redis interface {
	// redis: -struct_name=Dao -key=qrcodeKey
	CacheQrcode(ctx context.Context, ticket string) (int64, error)
	// redis: -struct_name=Dao -key=qrcodeKey -expire=d.qrcodeExpire
	AddCacheQrcode(ctx context.Context, ticket string, mid int64) error
	// redis: -struct_name=Dao -key=qrcodeKey
	DelCacheQrcode(ctx context.Context, ticket string) error
	// redis: -struct_name=Dao -key=qrcodeBindKey
	CacheQrcodeBind(ctx context.Context, ticket string) (int64, error)
	// redis: -struct_name=Dao -key=qrcodeBindKey -expire=d.qrcodeStatusExpire
	AddCacheQrcodeBind(ctx context.Context, ticket string, mid int64) error
	// redis: -struct_name=Dao -key=timelockPwdKey
	CacheTimelockPwd(ctx context.Context, mid int64) (string, error)
	// redis: -struct_name=Dao -key=timelockPwdKey -expire=d.timelockPwdExpire
	AddCacheTimelockPwd(ctx context.Context, mid int64, pwd string) error
	// redis: -struct_name=Dao -key=timelockPwdKey
	DelCacheTimelockPwd(ctx context.Context, mid int64) error
}
