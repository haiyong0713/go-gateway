package service

import (
	"context"
	"encoding/json"

	featureAdminMdl "go-gateway/app/app-svr/app-feed/admin/model/feature"
	buildLimitmdl "go-gateway/app/app-svr/feature/service/model/buildLimit"

	"go-common/library/log"

	"go-gateway/app/app-svr/feature/service/api"
)

func (s *Service) BuildLimit(c context.Context, args *api.BuildLimitReq) (res *api.BuildLimitReply, err error) {
	res = new(api.BuildLimitReply)
	var tmp = make(map[string]*buildLimitmdl.BuildLimit)
	for k, v := range s.buildLimitCache[args.TreeId] {
		tmp[k] = v
	}
	for k, v := range s.buildLimitCache[featureAdminMdl.Common] {
		tmp[k] = v
	}
	for keyname, tmp1 := range tmp {
		if tmp1 == nil {
			continue
		}
		var resKeyName = new(api.BuildLimitkeys)
		resKeyName.KeyName = keyname
		if err = json.Unmarshal(tmp1.Conditions, &resKeyName.Plats); err != nil {
			log.Warn("%v", err)
			err = nil
			continue
		}
		res.Keys = append(res.Keys, resKeyName)
	}
	return
}
