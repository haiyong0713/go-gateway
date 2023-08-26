package s10

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	xecode "go-common/library/ecode"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"go-gateway/app/web-svr/activity/admin/component"
	"go-gateway/app/web-svr/activity/admin/model/s10"
	"go-gateway/app/web-svr/activity/ecode"

	"go-common/library/log"
)

var dataToCsv [][]string
var s10cateHeader = []string{"mid", "昵称", "社区等级", "robin", "gid", "商品名称"}

func (s *Service) AllLotteryUser(ctx context.Context, robin int32) error {
	// 检查是否点击过
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	opened, err := redis.Int(conn.Do("GET", fmt.Sprintf("s10_lottery_user_robin_%d", robin)))
	if err != nil {
		log.Errorc(ctx, "redis get err[%v]", err)
	}
	if opened > 0 {
		return xecode.Error(xecode.RequestErr, "该轮次已开奖")
	}
	_, err = conn.Do("SET", fmt.Sprintf("s10_lottery_user_robin_%d", robin), time.Now().Unix())
	if err != nil {
		log.Errorc(ctx, "redis set err[%v]", err)
	}
	_, err = s.dao.DelSuperLuckyUser(ctx, robin)
	if err != nil {
		return err
	}
	go s.allLotteryUser(robin)
	return nil
}

func (s *Service) allLotteryUser(robin int32) {
	var (
		id          int64
		lotteryList []int64
		ctx         = context.Background()
		res         = make([]int64, 0, 1000000)
		action      = []string{"s10point"}
	)
	defer func() {
		dataToCsv = [][]string{}
	}()
	for {
		mids, next, err := s.dao.LotteryUsersByRobin(ctx, robin, id)
		if err != nil {
			component.Rebot("LotteryUsersByRobin fail")
			return
		}
		log.Info("s10 s.s10Dao.LotteryUsersByRobin (id:%d)", id)
		if id == next {
			break
		}
		id = next
		res = append(res, mids...)
		time.Sleep(50 * time.Millisecond)
	}
	allRes := res
	goodses, err := s.dao.GoodsByRobin(ctx, robin)
	if err != nil {
		component.Rebot("GoodsByRobin fail")
		return
	}
	sort.Slice(goodses, func(i, j int) bool {
		return goodses[i].Score > goodses[j].Score
	})
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, goods := range goodses {
		if err = s.cleanDataAboutExcept(ctx, goods.Gid, robin); err != nil {
			component.Rebot("cleanDataAboutExcept fail")
			log.Info("s10 cleanDataAboutExcept error:%v", err)
			return
		}
		v := goods.Stock
		if len(res) == 0 {
			break
		}
		writeList := make([]*s10.UserInfo, 0, 100)
		for {
			if v == 0 {
				break
			}
			length := len(res)
			if length == 0 {
				break
			}
			if v >= length {
				lotteryList = res
				res = res[:0]
			} else {
				for i := 0; i < v; i++ {
					tmp := length - i - 1
					j := r.Int() % tmp
					res[tmp], res[j] = res[j], res[tmp]
				}
				lotteryList = res[length-v:]
				res = res[:length-v]
			}
			for _, mid := range lotteryList {
				tmp := &s10.UserInfo{Mid: mid}
				tmp.Name, tmp.Level, err = s.accountCheck(ctx, mid, action)
				if err != nil {
					log.Warnc(ctx, "s10 UserInfo error:%v", err)
					continue
				}
				writeList = append(writeList, tmp)
			}
			v = goods.Stock - len(writeList)
		}
		if len(writeList) > 0 {
			log.Info("s10 lotteryGoods:%d", goods.Gid)
			component.Rebot(fmt.Sprintf("%s发放成功", goods.Gname))
			if goods.Type == 1 {
				s.luckyCertUser(ctx, goods.Gid, robin, writeList)
				continue
			}
			switch goods.State {
			case 0:
				s.luckyUser(ctx, goods.Gid, robin, writeList)
			case 2:
				s.genCsvData(goods.Gid, robin, goods.Gname, writeList)
			}
		}
	}
	fileName := fmt.Sprintf("s10第%d轮大奖名单.csv", robin)
	err = component.CreateCsvAndSend(s.s10FilePath, fileName, "s10大奖名单", s10cateHeader, dataToCsv, s.s10MaiInfo)
	if err != nil {
		log.Warnc(ctx, "s10 superLotteryList error:%v", err)
	}
	allResStr := make([]string, 0, len(allRes))
	for _, mid := range allRes {
		allResStr = append(allResStr, strconv.Itoa(int(mid)))
	}
	fileName = fmt.Sprintf("s10第%d轮参与名单.csv", robin)
	err = component.CreateSignleColCsvAndSend(s.s10FilePath, fileName, "s10参与名单", allResStr, s.s10MaiInfo)
	if err != nil {
		log.Warnc(ctx, "s10 fileName:%s error:%v", fileName, err)
	}
	fileName = fmt.Sprintf("s10第%d轮未中奖名单.csv", robin)
	allResStr = allResStr[:len(res)]
	err = component.CreateSignleColCsvAndSend(s.s10FilePath, fileName, "s10未中奖名单", allResStr, s.s10MaiInfo)
	if err != nil {
		log.Warnc(ctx, "s10 fileName:%s error:%v", fileName, err)
	}
	log.Info("s10 lotteryGoods finish")
	component.Rebot("s10 lotteryGoods finish")
}

func (s *Service) cleanDataAboutExcept(ctx context.Context, gid, robin int32) error {
	mids, err := s.dao.LotteryUserByGid(ctx, gid)
	if err != nil {
		return err
	}
	for _, mid := range mids {
		if err = s.dao.DelLotteryCache(ctx, mid, robin); err != nil {
			return err
		}
	}
	_, err = s.dao.UpdateLotteryUserStateByGid(ctx, gid)
	return err
}

func (s *Service) genCsvData(gid, robin int32, gname string, list []*s10.UserInfo) {
	for _, v := range list {
		dataToCsv = append(dataToCsv, []string{fmt.Sprintf("%d", v.Mid), v.Name, fmt.Sprintf("%d", v.Level), fmt.Sprintf("%d", robin), fmt.Sprintf("%d", gid), gname})
	}
}

func (s *Service) luckyCertUser(ctx context.Context, gid, robin int32, list []*s10.UserInfo) {
	batchSize := 50
	certs, err := s.dao.GiftCertsByGid(ctx, gid)
	if err != nil {
		fmt.Println(err)
		return
	}
	mids := make([]int64, 0, len(list))
	for _, v := range list {
		mids = append(mids, v.Mid)
	}
	allMids := mids
	allCerts := certs
	count := len(list)
	if count > len(certs) {
		count = len(certs)
	}
	allCount := count
	for count > 0 {
		if batchSize > count {
			batchSize = count
		}
		_, err := s.retry(func() (int64, error) {
			return s.dao.BatchAddLuckyCertUser(ctx, gid, robin, mids[:batchSize], certs[:batchSize])
		})
		if err != nil {
			log.Errorc(ctx, "s10 luckyCertUser mids:%v,certs:%+v,error:%v", mids[:batchSize], certs[:batchSize], err)
		}
		mids = mids[batchSize:]
		certs = certs[batchSize:]
		count -= batchSize
	}
	gift := &s10.MatchUser{IsLottery: true}
	gift.Lucky = &s10.Lucky{Gid: gid}
	for i := 0; i < allCount; i++ {
		gift.Lucky.Extra = allCerts[i]
		_, err := s.retry(func() (res int64, err error) {
			err = s.dao.AddLotteryFieldCache(ctx, allMids[i], robin, gift)
			return
		})
		if err != nil {
			log.Errorc(ctx, "s10 luckyCertUser error:%v", err)
		}
	}

}

func (s *Service) retry(f func() (int64, error)) (res int64, err error) {
	for i := 0; i < 3; i++ {
		if res, err = f(); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func (s *Service) luckyUser(ctx context.Context, gid, robin int32, list []*s10.UserInfo) {
	mids := make([]int64, 0, len(list))
	for _, v := range list {
		mids = append(mids, v.Mid)
	}
	allMids := mids
	batchSize := 50
	for len(mids) > 0 {
		if batchSize > len(mids) {
			batchSize = len(mids)
		}
		_, err := s.retry(func() (int64, error) {
			return s.dao.BatchAddLuckyUser(ctx, gid, robin, mids[:batchSize])

		})
		if err != nil {
			log.Errorc(ctx, "s10 luckyUser mids: %v error:%v", mids[:batchSize], err)
		}
		mids = mids[batchSize:]
	}
	gift := &s10.MatchUser{IsLottery: true}
	gift.Lucky = &s10.Lucky{Gid: gid}
	for _, mid := range allMids {
		_, err := s.retry(func() (i int64, err error) {
			err = s.dao.AddLotteryFieldCache(ctx, mid, robin, gift)
			return
		})
		if err != nil {
			log.Errorc(ctx, "s10 luckyUser error:%v", err)
		}
	}
}

func (s *Service) accountCheck(ctx context.Context, mid int64, action []string) (string, int32, error) {
	accInfo, err := s.dao.Profile(ctx, mid)
	if err != nil {
		return "", 0, err
	}
	if accInfo.GetTelStatus() != 1 {
		return "", 0, ecode.ActivityVogueTelValid
	}
	ok, err := s.dao.InBackList(ctx, mid, action)
	if err != nil {
		return "", 0, err
	}
	if ok {
		return "", 0, ecode.ActivityUerInBlackList
	}
	return accInfo.Name, accInfo.Level, nil
}

func (s *Service) GenBackupUsers() {
	var (
		err             error
		id, next, check int64
		mids            []int64
		ctx             = context.Background()
		userMap         = make(map[int64]int32, 5000000)
		giftUserMap     = make(map[int64]struct{}, 100000)
		action          = []string{"s10point"}
	)
	for {
		if mids, next, err = s.dao.LotteryUsers(ctx, id); err != nil {
			time.Sleep(100 * time.Millisecond)
			log.Error("s10 LotteryUsers(id:%d) error:%v", id, err)
			continue
		}
		log.Warn("s10LotteryUsers(id:%d)", id)
		if id == next {
			break
		}
		id = next
		for _, mid := range mids {
			userMap[mid] += 1
		}
		time.Sleep(50 * time.Millisecond)
	}
	for {
		if mids, next, err = s.dao.GiftUsers(ctx, id); err != nil {
			time.Sleep(100 * time.Millisecond)
			log.Error("s10 GiftUsers(id:%d) error:%v", id, err)
			continue
		}
		log.Warn("s10 GiftUsers(id:%d)", id)
		if id == next {
			break
		}
		id = next
		for _, mid := range mids {
			giftUserMap[mid] = struct{}{}
		}
		time.Sleep(50 * time.Millisecond)
	}

	res := make([]int64, 0, len(userMap))
	for k, v := range userMap {
		_, ok := giftUserMap[k]
		if ok {
			continue
		}
		if v < 2 {
			continue
		}
		check, err = s.retry(func() (int64, error) {
			accInfo, err := s.dao.Profile(ctx, k)
			if err != nil {
				return 0, err
			}
			if accInfo.GetTelStatus() != 1 {
				return 0, nil
			}
			return 1, nil
		})
		if err != nil || check == 0 {
			continue
		}
		check, err = s.retry(func() (int64, error) {
			ok, err := s.dao.InBackList(ctx, k, action)
			if err != nil {
				return 0, err
			}
			if ok {
				return 0, nil
			}
			return 1, nil
		})
		if err != nil || check == 0 {
			continue
		}
		res = append(res, k)
	}
	resStr := make([]string, 0, len(res))
	for _, v := range res {
		resStr = append(resStr, strconv.Itoa(int(v)))
	}
	fileName := "非酋头像框名单.csv"
	err = component.CreateSignleColCsvAndSend(s.s10FilePath, fileName, "非酋头像框名单", resStr, s.s10MaiInfo)
	if err != nil {
		log.Infoc(ctx, "s10 fileName:%s error:%v", fileName, err)
	}
	component.Rebot(fmt.Sprintf("%s error:%v", fileName, err))
}
