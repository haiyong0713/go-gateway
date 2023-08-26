package adapters

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/dao/vote"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
)

// up: Up主
type up struct {
	Mid    int64  `json:"mid"`
	Avatar string `json:"avatar"`
	Name   string `json:"name"`
}

func (i *up) GetName() string {
	return i.Name
}

func (i *up) GetId() int64 {
	return i.Mid
}

func (i *up) GetSearchField1() string {
	return i.Name
}

func (i *up) GetSearchField2() string {
	return ""
}

func (i *up) GetSearchField3() string {
	return ""
}

func getVoteUpInfoByMids(ctx context.Context, mids []int64) (res []vote.DataSourceItem, err error) {
	res = make([]vote.DataSourceItem, 0, len(mids))
	midInfos, err := client.AccountClient.Infos3(ctx, &accapi.MidsReq{
		Mids:   mids,
		RealIp: "",
	})
	if err != nil {
		return
	}

	for _, mid := range mids {
		info, ok := midInfos.Infos[mid]
		if !ok {
			continue
		}
		res = append(res, &up{
			Mid:    mid,
			Avatar: info.Face,
			Name:   info.Name,
		})
	}
	return
}
