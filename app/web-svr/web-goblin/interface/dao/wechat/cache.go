package wechat

import (
	"context"

	"go-gateway/app/web-svr/web-goblin/interface/model/wechat"
)

//go:generate kratos tool btsgen
type _bts interface {
	// cache
	AccessToken(c context.Context) (*wechat.AccessToken, error)
}
