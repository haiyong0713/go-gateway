package s10

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/client"
	"strconv"
	"strings"

	grab "git.bilibili.co/bapis/bapis-go/garb/service"

	"go-common/library/log"
)

/*
	message GrantByBizReq{
		repeated int64 mids = 1; // 多个mid 最多100个
		repeated int64 ids = 2; // 装扮id（非套装）
		int64 addSecond = 3; // 必填 发放时长 单位秒 大于0
		string token = 4; // 不必填 同一个token 只发一次
	}

	message GrantSuitReq{
		repeated int64 mids = 1; // 必填 多个mid 最多100个
		int64 suitID = 2; // 必填 套装id
		int64 addSecond = 3; // 必填 发放时长 单位秒 大于0
		string token = 4; // 非必填 同一个token 只发一次
		string business = 5; // 业务方名发放目的 最长16个字符
	}
*/
func (d *Dao) GrantByBiz(ctx context.Context, mid int64, uniqueID, extra string) error {
	var err error
	strs := strings.Split(extra, ":")
	if len(strs) < 2 {
		log.Errorc(ctx, "s10 goods info error! uniqueID:%s", uniqueID)
		return fmt.Errorf("商品信息错误")
	}
	ints := make([]int64, len(strs))
	for i := 0; i < 2; i++ {
		if ints[i], err = strconv.ParseInt(strs[i], 10, 64); err != nil {
			log.Errorc(ctx, "s10 d.dao.GrantByBiz extra  to Num fail mid:%d, extra:%s error:%v", mid, extra, err)
			return err
		}
	}
	req := &grab.GrantByBizReq{Mids: []int64{mid}, Ids: []int64{ints[1]}, Token: uniqueID, AddSecond: ints[0] * 24 * 60 * 60}
	if _, err = client.GarbClient.GrantByBiz(ctx, req); err != nil {
		log.Errorc(ctx, "s10 d.dao.GrantByBiz(mid:%s) error:%v", uniqueID, err)
	}
	return err
}

func (d *Dao) GrantSuit(ctx context.Context, mid int64, uniqueID, extra string) error {
	var err error
	strs := strings.Split(extra, ":")
	if len(strs) < 2 {
		log.Errorc(ctx, "s10 goods info error! uniqueID:%s", uniqueID)
		return fmt.Errorf("商品信息错误")
	}
	ints := make([]int64, len(strs))
	for i := 0; i < 2; i++ {
		if ints[i], err = strconv.ParseInt(strs[i], 10, 64); err != nil {
			log.Errorc(ctx, "s10 d.dao.GrantSuit extra  to Num fail mid:%d, extra:%s error:%v", mid, extra, err)
			return err
		}
	}
	req := &grab.GrantSuitReq{Mids: []int64{mid}, SuitID: ints[1], Token: uniqueID, Business: "S10积分兑换商品", AddSecond: ints[0] * 24 * 60 * 60}
	if _, err = client.GarbClient.GrantSuit(ctx, req); err != nil {
		log.Errorc(ctx, "s10 d.dao.GrantSuit(mid:%s) error:%v", uniqueID, err)
	}
	return err
}
