package model

import (
	"context"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/library/exp/ab"
)

var (
	topicNewVersionPubLayerFlag = ab.Int("topic_new_version_pub_layer_exp", "单列话题页参与按钮样式调整", 1)
)

func CanUseNewVersionPubLayerWithAvatar(ctx context.Context) int64 {
	return doAbtestOnIntFlag(ctx, topicNewVersionPubLayerFlag)
}

func doAbtestOnIntFlag(ctx context.Context, intFlag *ab.IntFlag) int64 {
	au, ok := auth.FromContext(ctx)
	if !ok {
		return -1
	}
	d, ok := device.FromContext(ctx)
	if !ok {
		return -1
	}
	t, ok := ab.FromContext(ctx)
	if !ok {
		return -1
	}
	t.Add(ab.KVString("buvid", d.Buvid))
	t.Add(ab.KVInt("mid", au.Mid))
	return intFlag.Value(t)
}
