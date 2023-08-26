package system

import (
	"context"
	"go-common/library/log"
	"go-common/library/net/metadata"
	model "go-gateway/app/web-svr/activity/admin/model/system"
	"net/url"
)

const getSystemUserInfoByUID = "/x/internal/activity/system/users/info/v1"

func (d *Dao) GetUsersInfo(ctx context.Context, uids []string) (res map[string]*model.UsersInfoDetail, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(ctx, metadata.RemoteIP)
	)
	res = make(map[string]*model.UsersInfoDetail)

	response := new(model.GetUsersInfo)
	for _, uid := range uids {
		params.Add("uids", uid)
	}
	if err = d.client.Get(ctx, d.c.Host.API+getSystemUserInfoByUID, ip, params, response); err != nil || response.Code != 0 {
		log.Errorc(ctx, "GetUsersInfo d.client.Get Err url(%v) params(%v) response(%v)", d.c.Host.API+getSystemUserInfoByUID, params, response)
		return
	}

	for _, v := range response.Data {
		res[v.UID] = v
	}
	return
}
