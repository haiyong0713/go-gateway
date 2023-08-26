package web

import (
	"context"
	"encoding/json"
	"net/url"

	"go-common/library/log"
	"go-gateway/app/web-svr/web-goblin/interface/model/web"
)

// Recruit .
func (s *Service) Recruit(ctx context.Context, param url.Values, route *web.Params) (res json.RawMessage, err error) {
	if res, err = s.dao.Recruit(ctx, param, route); err != nil {
		log.Error("s.dao.Recruit route(%s) error(%v)", route.Route, err)
	}
	return
}
