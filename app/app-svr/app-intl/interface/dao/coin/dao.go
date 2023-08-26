package coin

import (
	"context"

	api "git.bilibili.co/bapis/bapis-go/community/service/coin"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-intl/interface/conf"

	"github.com/pkg/errors"
)

const (
	coinAv         = 1
	coinArticle    = 2
	coinBizAv      = "archive"
	coinBizArticle = "article"
)

func coinBusiness(avtype int64) string {
	switch avtype {
	case coinAv:
		return coinBizAv
	case coinArticle:
		return coinBizArticle
	}
	return ""
}

// Dao is coin dao
type Dao struct {
	coinClient api.CoinClient
}

// New initial coin dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.coinClient, err = api.NewClient(c.CoinClient); err != nil {
		panic(err)
	}
	return
}

// AddCoins add coin to upper.
func (d *Dao) AddCoins(c context.Context, aid, mid, upID, maxCoin, avtype, multiply int64, typeID int16, pubTime int64) (err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	// arg := &model.ArgAddCoin{Mid: mid, UpMid: upID, MaxCoin: maxCoin, Aid: aid, AvType: avtype, Multiply: multiply, RealIP: ip, TypeID: typeID, PubTime: pubTime}
	arg := &api.AddCoinReq{
		Mid:      mid,
		IP:       ip,
		Upmid:    upID,
		MaxCoin:  maxCoin,
		Aid:      aid,
		Typeid:   int32(typeID),
		PubTime:  pubTime,
		Number:   multiply,
		Business: coinBusiness(avtype),
	}
	_, err = d.coinClient.AddCoin(c, arg)
	return
}

// ArchiveUserCoins .
func (d *Dao) ArchiveUserCoins(c context.Context, aid, mid, avtype int64) (num int64, err error) {
	// arg := &model.ArgCoinInfo{Mid: mid, Aid: aid, AvType: avType, RealIP: ip}
	var reply *api.ItemUserCoinsReply
	arg := &api.ItemUserCoinsReq{
		Mid:      mid,
		Aid:      aid,
		Business: coinBusiness(avtype),
	}
	if reply, err = d.coinClient.ItemUserCoins(c, arg); err != nil || reply == nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	num = reply.Number
	return
}

// UserCoins get user coins
func (d *Dao) UserCoins(c context.Context, mid int64) (count float64, err error) {
	var reply *api.UserCoinsReply
	// arg := &model.ArgCoinInfo{Mid: mid, RealIP: ip}
	arg := &api.UserCoinsReq{
		Mid: mid,
	}
	if reply, err = d.coinClient.UserCoins(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	count = reply.Count
	return
}
