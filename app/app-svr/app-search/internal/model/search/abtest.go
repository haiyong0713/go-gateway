package search

import (
	"context"

	"go-common/component/metadata/device"
	"go-common/library/exp/ab"
)

var (
	ogvVipNewUserFlag = ab.Int("ogv_vip_new_user_gift", "ogv新人用户角标实验", -1)
)

func doOgvNewUserAbtest(ctx context.Context) int64 {
	d, ok := device.FromContext(ctx)
	if !ok {
		return 0
	}
	t, ok := ab.FromContext(ctx)
	if !ok {
		return 0
	}
	t.Add(ab.KVString("buvid", d.Buvid))
	exp := ogvVipNewUserFlag.Value(t)
	return exp
}
