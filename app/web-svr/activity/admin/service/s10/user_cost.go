package s10

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"go-gateway/app/web-svr/activity/admin/model/s10"

	"go-common/library/log"
)

func (s *Service) UserCostInfo(ctx context.Context, mid int64) ([]*s10.UserCostRecord, error) {
	var (
		err error
		res []*s10.UserCostRecord
	)
	if s.subTabSwitch {
		res, err = s.dao.UserCostExceptSub(ctx, mid)
	} else {
		res, err = s.dao.UserCostExcept(ctx, mid)
	}
	if err != nil {
		return nil, err
	}
	exceptRecord := make([]*s10.UserCostRecord, 0, len(res))
	normalRecord := make([]*s10.UserCostRecord, 0, len(res))
	for _, v := range res {
		v.UniqueID = fmt.Sprintf("s10:0:%d:%d", mid, v.ID)
		if v.Ack != 0 || v.Gid < 10 {
			normalRecord = append(normalRecord, v)
			continue
		}
		exceptRecord = append(exceptRecord, v)
	}
	sort.Slice(exceptRecord, func(i, j int) bool {
		return exceptRecord[i].Ctime > exceptRecord[j].Ctime
	})
	sort.Slice(normalRecord, func(i, j int) bool {
		return normalRecord[i].Ctime > normalRecord[j].Ctime
	})
	copy(res, exceptRecord)
	copy(res[len(exceptRecord):], normalRecord)
	return res, nil
}

func (s *Service) SuperLotteryUser(ctx context.Context, users []*s10.SuperLotteryUserInfo) error {
	gids := make([]int64, 0, 10)
	var robin int32
	for _, v := range users {
		gids = append(gids, int64(v.Gid))
	}
	if len(users) != 0 {
		robin = users[0].Robin
	}
	if robin == 0 {
		return nil
	}
	_, err := s.dao.DelSuperLuckyUser(ctx, robin)
	if err != nil {
		return err
	}
	mids, err := s.dao.LotteryUserByGids(ctx, gids)
	if err != nil {
		return err
	}
	for _, mid := range mids {
		if err = s.dao.DelLotteryCache(ctx, mid, robin); err != nil {
			return err
		}
	}
	if _, err = s.dao.UpdateLotteryUserStateByGids(ctx, gids); err != nil {
		return err
	}
	batchsize := 50
	allUsers := users
	for len(users) > 0 {
		if batchsize > len(users) {
			batchsize = len(users)
		}
		_, err := s.dao.BatchAddSuperGiftUser(ctx, users[:batchsize])
		if err != nil {
			return err
		}
		_, err = s.dao.BatchAddSuperLuckyUser(ctx, users[:batchsize])
		if err != nil {
			return err
		}
		users = users[batchsize:]
	}
	for _, v := range allUsers {
		gift := &s10.MatchUser{IsLottery: true}
		gift.Lucky = &s10.Lucky{Gid: v.Gid}
		_, err := s.retry(func() (i int64, err error) {
			err = s.dao.AddLotteryFieldCache(ctx, v.Mid, robin, gift)
			return
		})
		if err != nil {
			log.Errorc(ctx, "s10 luckyUser error:%v", err)
			return err
		}
	}
	return nil
}

func (s *Service) NotExistLotteryUserByRobin(ctx context.Context, mids []int64, robin int32) ([]int64, error) {
	user, err := s.dao.ExistUsersByRobin(ctx, mids, robin)
	if err != nil {
		return nil, err
	}
	userMap := make(map[int64]struct{}, len(user))
	res := make([]int64, 0, len(mids))
	for _, v := range user {
		userMap[v] = struct{}{}
	}
	for _, mid := range mids {
		if _, ok := userMap[mid]; !ok {
			res = append(res, mid)
		}
	}
	return res, nil
}

func (s *Service) UserGiftInfo(ctx context.Context, mid int64) ([]*s10.UserGiftRecord, error) {
	res, gids, err := s.dao.UserGift(ctx, mid)
	if err != nil {
		return nil, err
	}
	if len(gids) == 0 {
		return nil, nil
	}
	goodsMap, err := s.dao.Goodses(ctx, gids)
	normalRes := make([]*s10.UserGiftRecord, 0, 5)
	exceptRes := make([]*s10.UserGiftRecord, 0, 5)
	unReceiveRes := make([]*s10.UserGiftRecord, 0, 5)
	for _, v := range res {
		v.UniqueID = fmt.Sprintf("s10:1:%d:%d", mid, v.ID)
		v1, ok := goodsMap[v.Gid]
		if !ok {
			fmt.Errorf("商品出现错误")
		}
		v.Name = v1.Gname
		if v.State != 0 {
			if v1.Type <= 1 {
				normalRes = append(normalRes, v)
				continue
			}
			if v.Ack == 0 {
				exceptRes = append(exceptRes, v)
				continue
			}
			normalRes = append(normalRes, v)
			continue
		}
		if v1.Type == 1 {
			normalRes = append(normalRes, v)
			continue
		}
		unReceiveRes = append(unReceiveRes, v)
	}
	copy(res, exceptRes)
	copy(res[len(exceptRes):], unReceiveRes)
	copy(res[len(exceptRes)+len(unReceiveRes):], normalRes)
	return res, nil
}

func (s *Service) RealGoodsList(ctx context.Context, robin int32) ([][]string, error) {
	goods, err := s.dao.GoodsByRobin(ctx, robin)
	goodsMap := make(map[int32]*s10.Goods, len(goods))
	for _, v := range goods {
		goodsMap[v.Gid] = v
	}
	if err != nil {
		return nil, err
	}
	gids := make([]int64, 0, len(goodsMap))
	for k, v := range goodsMap {
		if v.Type == 0 {
			gids = append(gids, int64(k))
		}
	}
	res, err := s.dao.RealGoodsUser(ctx, gids)
	if err != nil {
		return nil, err
	}
	result := make([][]string, 0, len(res))
	var name string
	for _, v := range res {
		if v1, ok := goodsMap[v.Gid]; ok {
			name = v1.Gname
		}
		result = append(result, []string{strconv.Itoa(int(v.ID)), strconv.Itoa(int(v.Mid)), name, v.UserName, v.Number, v.Addr})
	}
	return result, nil
}

func (s *Service) SentOutGoods(ctx context.Context, ids []int64) error {
	_, err := s.dao.SentOutGoods(ctx, ids)
	return err
}

func (s *Service) BackupUsers(gid int64, mids []int64) {
	var (
		err error
		ctx = context.Background()
	)

	for i, mid := range mids {
		if err = s.dao.Redelivery(ctx, int64(i), mid, gid, 3); err != nil {
			log.Errorc(ctx, "s10 BackupUsers mid:%d,gid:%d error:%v", mid, gid, err)
		}
	}
}
