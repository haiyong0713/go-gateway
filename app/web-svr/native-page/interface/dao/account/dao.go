package account

import (
	"context"

	api "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"go-gateway/app/web-svr/native-page/interface/conf"
)

type Dao struct {
	accClient api.AccountClient
}

func New(cfg *conf.Config) *Dao {
	accClient, err := api.NewClient(cfg.AccClient)
	if err != nil {
		panic(err)
	}
	return &Dao{accClient: accClient}
}

// Cards3GRPC card grpc
func (d *Dao) Cards3GRPC(c context.Context, rawMids []int64) (map[int64]*api.Card, error) {
	var (
		cardsReply *api.CardsReply
		err        error
	)
	mids := midFilter(rawMids)
	if len(mids) == 0 {
		return make(map[int64]*api.Card), nil
	}
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &api.MidsReq{Mids: mids, RealIp: ip}
	if cardsReply, err = d.accClient.Cards3(c, arg); err != nil {
		return nil, err
	}
	return cardsReply.GetCards(), nil
}

func (d *Dao) Card3(c context.Context, mid int64) (*api.CardReply, error) {
	req := &api.MidReq{
		Mid:    mid,
		RealIp: metadata.String(c, metadata.RemoteIP),
	}
	rly, err := d.accClient.Card3(c, req)
	if err != nil {
		log.Error("Fail to request account.Card3(), req=%+v error=%+v", req, err)
		return nil, err
	}
	return rly, nil
}

func midFilter(rawMids []int64) []int64 {
	//去重+剔除为0的值
	midsSet := make(map[int64]struct{})
	for _, v := range rawMids {
		midsSet[v] = struct{}{}
	}
	var mids []int64
	for mid := range midsSet {
		if mid > 0 {
			mids = append(mids, mid)
		}
	}
	return mids
}
