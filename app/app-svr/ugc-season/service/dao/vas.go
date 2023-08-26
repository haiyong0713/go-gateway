package dao

import (
	"context"
	"go-common/library/cache/redis"
	"go-common/library/log"

	vasGrpc "git.bilibili.co/bapis/bapis-go/vas/trans/service"

	"go-gateway/app/app-svr/ugc-season/service/api"
)

const (
	_seasonGoodsPrefix = "s_g_p_"
	_seasonGoodsExpire = 10
)

func seasonGoodsKey(sid string) string {
	return _seasonGoodsPrefix + sid
}

func (d *Dao) GetGoodsInfoFromCache(c context.Context, seasonId string) *api.GoodsInfo {
	conn := d.redis.Get(c)
	defer conn.Close()
	bs, err := redis.Bytes(conn.Do("GET", seasonGoodsKey(seasonId)))
	if err != nil && err != redis.ErrNil {
		log.Error("日志告警 付费合集 缓存获取商品信息失败 seasonId(%s) error(%+v)", seasonId, err)
	}

	gi := new(api.GoodsInfo)
	if bs != nil {
		if err = gi.Unmarshal(bs); err != nil {
			log.Error("日志告警 付费合集 缓存Unmarshal商品信息失败 seasonId(%s) error(%+v)", seasonId, err)
		} else {
			return gi
		}
	}

	res, err := d.vasTransItemInfo(c, seasonId)
	if err != nil || res == nil || res.Result == nil || res.Result[seasonId] == nil {
		log.Error("日志告警 付费合集 vas获取商品信息失败 seasonId(%s) err %v", seasonId, err)
		return nil
	}
	gi.GoodsId = res.Result[seasonId].ProductId
	gi.GoodsName = res.Result[seasonId].Name
	gi.GoodsPrice = res.Result[seasonId].Price
	gi.GoodsPriceFmt = res.Result[seasonId].PriceFmt

	err = d.setSeasonCache(c, seasonId, gi)
	if err != nil {
		log.Error("日志告警 付费合集 商品信息设置缓存失败 seasonId(%s) err %v", seasonId, err)
	}
	return gi
}

// AddSeasonCache set season into cache.
func (d *Dao) setSeasonCache(c context.Context, seasonId string, gi *api.GoodsInfo) error {
	bs, err := gi.Marshal()
	if err != nil {
		log.Error("goods.Marshal error(%+v)", err)
		return err
	}

	key := seasonGoodsKey(seasonId)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SETEX", key, _seasonGoodsExpire, bs); err != nil {
		log.Error("conn.Do(SET, %s) error(%+v)", key, err)
		return err
	}
	return nil
}

func (d *Dao) vasTransItemInfo(c context.Context, seasonId string) (*vasGrpc.VasTransItemInfoReply, error) {
	req := &vasGrpc.VasTransItemInfoReq{
		BizItemIds: []string{seasonId},
		Category:   vasGrpc.Category_CategorySeason,
	}
	reply, err := d.vasGRPC.VasTransItemInfo(c, req)
	if err != nil {
		log.Error("d.vasGRPC.VasTransItemInfo seasonId(%s) err %v", seasonId, err)
		return nil, err
	}
	return reply, nil
}
