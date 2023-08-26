package model

import "context"

type BkArchiveArg struct {
	// 开启服务端稿件过滤 开启后被过滤的稿件不会被返回
	EnableServerFilter bool
}

type _bkArgKey struct{}

func NewBkArchiveArgs(ctx context.Context, arg BkArchiveArg) context.Context {
	return context.WithValue(ctx, _bkArgKey{}, arg)
}

func BkArchiveArgs(ctx context.Context) BkArchiveArg {
	if arg, ok := ctx.Value(_bkArgKey{}).(BkArchiveArg); ok {
		return arg
	}
	return BkArchiveArg{}
}
