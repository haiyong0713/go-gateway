package bws

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"

	"go-common/library/sync/errgroup.v2"
)

const (
	_bwsBluetoothUpMid = "blue_mid_%d_%d"
	_bwsBluetoothUpKey = "blue_key_%d_%s"
)

func bwsBluetoothUpMid(bid, mid int64) string {
	return fmt.Sprintf(_bwsBluetoothUpMid, bid, mid)
}

func bwsBluetoothUpKey(bid int64, key string) string {
	return fmt.Sprintf(_bwsBluetoothUpKey, bid, key)
}

func (s *Service) InCatchUp(c context.Context, mid int64, p *bwsmdl.CatchUpper) ([]*bwsmdl.BluetoothUpInfo, error) {
	var (
		key = p.Key
		err error
	)
	if p.Key == "" {
		if key, err = s.midToKey(c, p.Bid, mid); err != nil {
			return nil, err
		}
	} else {
		if mid, _, err = s.keyToMid(c, p.Bid, p.Key); err != nil {
			return nil, err
		}
	}
	// 蓝牙设备mac地址
	bluetoothMacs := strings.Split(p.UpKeys, ",")
	// 最多10个
	if len(bluetoothMacs) > 10 || len(bluetoothMacs) == 0 {
		return nil, ecode.RequestErr
	}
	// 查询蓝牙mac地址对应的up信息
	var (
		cus    []*bwsmdl.CatchUser
		upMids []int64
		acs    map[int64]*accapi.Card
		upsmac = map[string]struct{}{}
	)
	// 查询当前捕获的所有，用于去重
	ups, _, _ := s.dao.CatchUps(c, p.Bid, key, 0, -1)
	for _, up := range ups {
		v, ok := s.blueUpMidCahce[bwsBluetoothUpMid(p.Bid, up.Mid)]
		if !ok {
			continue
		}
		upsmac[v.Key] = struct{}{}
	}
	for _, mac := range bluetoothMacs {
		// 当面蓝牙mac地址是否存在
		v, ok := s.blueUpKeyCahce[bwsBluetoothUpKey(p.Bid, mac)]
		if !ok {
			continue
		}
		// 去重复
		if _, ok := upsmac[mac]; ok {
			continue
		}
		cu := &bwsmdl.CatchUser{
			Mid: v.Mid,
			Key: key,
		}
		cus = append(cus, cu)
		upMids = append(upMids, v.Mid)
	}
	if len(upMids) == 0 || len(cus) == 0 {
		return []*bwsmdl.BluetoothUpInfo{}, nil
	}
	// 获取用户情况
	_, dayStr := todayDate()
	_, err = s.getUserDetail(c, p.Bid, mid, key, dayStr, false)
	if err != nil {
		log.Errorc(c, "s.getUserDetail(%v)", err)
		return nil, err
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		if err := s.dao.AddCatchUps(ctx, p.Bid, key, cus); err != nil {
			log.Error("s.dao.AddCatchUps mid(%d) error(%v)", mid, err)
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if acs, err = s.accCardms(ctx, upMids); err != nil {
			log.Error("s.accCardms error(%v)", err)
			return
		}
		return
	})
	dayInt, dayStr := todayDate()

	// 保存今日捕获up主个数
	eg.Go(func(ctx context.Context) (err error) {
		reason := &bwsmdl.LogReason{
			Reason: bwsmdl.ReasonUps,
			Params: fmt.Sprintf("%d", len(upMids)),
		}
		now := time.Now().Unix()
		orderNo := fmt.Sprintf("%d_%d_%d_%d", mid, p.Bid, bwsmdl.StarMapUps, now)
		if err = s.UpdateUserDetail(ctx, mid, p.Bid, 0, int64(len(upMids)), dayStr, 0, nil, false, reason, orderNo, key); err != nil {
			log.Errorc(ctx, "s.UpdateUserDetail error(%v)", err)
			return
		}
		return
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	s.cache.Do(c, func(c context.Context) {
		if taskErr := s.doPointTask(c, key, bwsmdl.TaskCateCatch, 0, dayInt, len(cus)); taskErr != nil {
			log.Error("InCatchUp doPointTask userToken:%s error:%v", key, err)
		}
	})
	var res []*bwsmdl.BluetoothUpInfo
	for _, mac := range bluetoothMacs {
		v, ok := s.blueUpKeyCahce[bwsBluetoothUpKey(p.Bid, mac)]
		if !ok {
			continue
		}
		up := &bwsmdl.BluetoothUpInfo{}
		if err := up.BluetoothUpInfoChange(v, acs[v.Mid]); err != nil {
			continue
		}
		res = append(res, up)
	}
	return res, nil
}

func (s *Service) CatchList(c context.Context, mid int64, p *bwsmdl.CatchUpper) ([]*bwsmdl.BluetoothUpInfo, int, error) {
	var (
		key = p.Key
		err error
	)
	if p.Key == "" {
		if key, err = s.midToKey(c, p.Bid, mid); err != nil {
			return nil, 0, err
		}
	} else {
		if mid, _, err = s.keyToMid(c, p.Bid, p.Key); err != nil {
			return nil, 0, err
		}
	}
	start := (p.Pn - 1) * p.Ps
	end := start + (p.Ps - 1)
	ups, count, err := s.dao.CatchUps(c, p.Bid, key, start, end)
	if err != nil {
		log.Error("%+v", err)
		return []*bwsmdl.BluetoothUpInfo{}, 0, nil
	}
	var (
		upMids []int64
	)
	for _, up := range ups {
		upMids = append(upMids, up.Mid)
	}
	acs, err := s.accCardms(c, upMids)
	if err != nil {
		log.Error("s.accCardms error(%v)", err)
		return nil, 0, err
	}
	var res []*bwsmdl.BluetoothUpInfo
	for _, up := range ups {
		u := &bwsmdl.BluetoothUpInfo{}
		if err := u.BluetoothUpChange(acs[up.Mid]); err != nil {
			continue
		}
		res = append(res, u)
	}
	return res, count, nil
}

func (s *Service) CatchBluetoothList(c context.Context, mid int64, p *bwsmdl.CatchUpper) ([]*bwsmdl.BluetoothUpInfo, error) {
	var (
		key  = p.Key
		err  error
		upsm = map[string]struct{}{}
	)
	if p.Key == "" {
		if key, err = s.midToKey(c, p.Bid, mid); err != nil {
			return nil, err
		}
	} else {
		if mid, _, err = s.keyToMid(c, p.Bid, p.Key); err != nil {
			return nil, err
		}
	}
	ups, _, err := s.dao.CatchUps(c, p.Bid, key, 0, -1)
	if err != nil {
		log.Error("%+v", err)
		return []*bwsmdl.BluetoothUpInfo{}, nil
	}
	var upMid []int64
	for _, up := range ups {
		upMid = append(upMid, up.Mid)
	}
	var res []*bwsmdl.BluetoothUpInfo
	for _, up := range ups {
		if _, ok := upsm[bwsBluetoothUpMid(p.Bid, up.Mid)]; ok {
			continue
		}
		upsm[bwsBluetoothUpMid(p.Bid, up.Mid)] = struct{}{}
		b, ok := s.blueUpMidCahce[bwsBluetoothUpMid(p.Bid, up.Mid)]
		if !ok {
			continue
		}
		u := &bwsmdl.BluetoothUpInfo{Key: b.Key}
		res = append(res, u)
	}
	return res, nil
}

func (s *Service) BluetoothUpsAll(c context.Context, p *bwsmdl.CatchUpper) []*bwsmdl.BluetoothUpInfo {
	upsAll, ok := s.bluemCache[p.Bid]
	if !ok {
		return []*bwsmdl.BluetoothUpInfo{}
	}
	var res []*bwsmdl.BluetoothUpInfo
	for _, v := range upsAll {
		u := &bwsmdl.BluetoothUpInfo{Key: v.Key}
		res = append(res, u)
	}
	return res
}

func (s *Service) loadBluetoothUpsCache() {
	tmp := map[int64][]*bwsmdl.BluetoothUp{}
	tmpMid := map[string]*bwsmdl.BluetoothUp{}
	tmpKey := map[string]*bwsmdl.BluetoothUp{}
	data, err := s.dao.BluetoothUps(context.Background(), s.c.Bws.Bws202012Bid)
	if err != nil {
		log.Error("s.dao.BluetoothUps error(%v)", err)
		return
	}
	for _, d := range data {
		tmp[d.Bid] = append(tmp[d.Bid], d)
		tmpMid[bwsBluetoothUpMid(d.Bid, d.Mid)] = d
		tmpKey[bwsBluetoothUpKey(d.Bid, d.Key)] = d
	}
	s.bluemCache = tmp
	s.blueUpMidCahce = tmpMid
	s.blueUpKeyCahce = tmpKey
}

// accCardms .
func (s *Service) accCardms(c context.Context, mids []int64) (map[int64]*accapi.Card, error) {
	var (
		arg       = &accapi.MidsReq{Mids: mids}
		tempReply *accapi.CardsReply
	)
	if len(mids) == 0 {
		return nil, ecode.RequestErr
	}
	tempReply, err := s.accClient.Cards3(c, arg)
	if err != nil {
		log.Error("s.accRPC.Cards3(%d) error(%v)", mids, err)
		return nil, err
	}
	return tempReply.Cards, nil
}
