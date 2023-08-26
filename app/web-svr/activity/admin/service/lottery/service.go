package lottery

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/conf"
	"go-gateway/app/web-svr/activity/admin/dao"
	"go-gateway/app/web-svr/activity/admin/dao/lottery"
	componentmdl "go-gateway/app/web-svr/activity/admin/model/component"
	lotmdl "go-gateway/app/web-svr/activity/admin/model/lottery"
	actapi "go-gateway/app/web-svr/activity/interface/api"

	api "git.bilibili.co/bapis/bapis-go/account/service"
	vipresource "git.bilibili.co/bapis/bapis-go/vip/resource/service"
)

// Service struct
type Service struct {
	c              *conf.Config
	lotDao         *lottery.Dao
	dao            *dao.Dao
	accClient      api.AccountClient
	UploadInfo     map[string]*lotmdl.UploadInfo
	GiftTasks      map[string]int64
	GiftTaskLock   *sync.RWMutex
	resourceClient vipresource.ResourceClient
	maiInfo        *componentmdl.EmailInfo
	actClient      actapi.ActivityClient
}

// Close service
func (s *Service) Close() {
	if s.lotDao != nil {
		s.lotDao.Close()
	}
	if s.dao != nil {
		s.dao.Close()
	}
}

// New Service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:            c,
		lotDao:       lottery.New(c),
		dao:          dao.New(c),
		UploadInfo:   make(map[string]*lotmdl.UploadInfo),
		GiftTasks:    make(map[string]int64),
		GiftTaskLock: new(sync.RWMutex),
		maiInfo:      c.Lottery.MailInfo,
	}
	var err error
	if s.accClient, err = api.NewClient(s.c.AccClient); err != nil {
		panic(err)
	}
	if s.resourceClient, err = vipresource.NewClient(c.VipClient); err != nil {
		panic(err)
	}
	if s.actClient, err = actapi.NewClient(c.ActClient); err != nil {
		panic(err)
	}
	go cleanUploadInfo(s)
	go runGiftTasks(s)
	return
}

func cleanUploadInfo(s *Service) {
	var (
		err error
		c   = context.Background()
		id  int64
	)
	for {
		time.Sleep(time.Duration(1) * time.Second)
		for k, v := range s.UploadInfo {
			if v.Status != lotmdl.UploadStart {
				if _, id, err = lotmdl.SplitUploadKey(k); err != nil {
					log.Errorc(c, "lottery task cleanUploadInfo error. lotmdl.SplitUploadKey() failed. error(%v)", err)
					continue
				}
				if err = s.UpdUploadStatus(c, v.Status, id); err != nil {
					log.Errorc(c, "lottery task cleanUploadInfo error. upload status update failed. error(%v)", err)
					continue
				}
				if err = s.UpdUploadStatusDraft(c, v.Status, id); err != nil {
					log.Errorc(c, "lottery task UpdUploadStatusDraft error. upload status update failed. error(%v)", err)
					continue
				}
				delete(s.UploadInfo, k)
			}
		}
	}
}

func runGiftTasks(s *Service) {
	c := context.Background()
	var err error
	if err = s.FixLotteryGiftTask(c); err != nil {
		log.Error("activity-admin lottery runGiftTasks init failed. error(%v)", err)
	}
	log.Info("activity-admin lottery initTask. data: %+v", s.GiftTasks)
	for {
		time.Sleep(time.Duration(1) * time.Second)
		s.GiftTaskLock.RLock()

		for k, v := range s.GiftTasks {
			if time.Now().Unix() > v {
				if err = s.updateGift(c, k); err != nil {
					continue
				}

				delete(s.GiftTasks, k)
			}
		}
		s.GiftTaskLock.RUnlock()
	}
}

func (s *Service) updateGift(c context.Context, k string) (err error) {
	var (
		tx *sql.Tx
	)
	if tx, err = s.lotDao.BeginTran(c); err != nil {
		log.Errorc(c, "s.lotDao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "%v", r)
			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(c, "tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(c, "tx.Commit() error(%v)", err)
		}
	}()
	sp := strings.Split(k, "|")
	var (
		id      int64
		t, num  int
		gift    *lotmdl.GiftInfo
		lotInfo *lotmdl.LotInfo
	)
	if len(sp) < 3 {
		return
	}
	if t, err = strconv.Atoi(sp[2]); err != nil {
		log.Errorc(c, "activity-admin lottery runGiftTasks error. strconv.ParseInt(%v) failed. error(%v)", sp[2], err)
		return
	}
	if id, err = strconv.ParseInt(sp[1], 10, 64); err != nil {
		log.Errorc(c, "activity-admin lottery runGiftTasks error. strconv.ParseInt(%v) failed. error(%v)", sp[2], err)
		return
	}
	if t == lotmdl.GiftTypeSend {
		if gift, err = s.lotDao.GiftDetailByIDTx(c, tx, id); err != nil {
			log.Errorc(c, "activity-admin lottery runGiftTasks error. s.lotDao.GiftDetailByID(%v) failed. error(%v)", id, err)
			return
		}
		if lotInfo, err = s.lotDao.LotDetailBySIDTx(c, tx, gift.Sid); err != nil {
			log.Errorc(c, "activity-admin lottery runGiftTasks error. s.lotDao.LotDetailBySID(%v) failed. error(%v)", gift.Sid, err)
			return
		}
		if num, err = s.lotDao.CountUploadTx(c, tx, lotInfo.ID, gift.ID); err != nil {
			log.Errorc(c, "activity-admin lottery runGiftTasks error. s.lotDao.CountUpload(lotID:%v, giftID:%v) failed. error(%v)", lotInfo.ID, gift.ID, err)
			return
		}
		if num < gift.Num {
			log.Errorc(c, "activity-admin lottery runGiftTasks error. 优惠券所上传的兑换码数量小于设置数量，无法进入奖池。giftID: %v. "+
				"当前上传数量: %d，设置数量: %d", gift.ID, num, gift.Num)
			delete(s.GiftTasks, k)
			return
		}
	}
	log.Infoc(c, "activity-admin lottery update, id: %v", id)
	if err = s.lotDao.UpdateGiftEffectTx(c, tx, id, 1); err != nil {
		log.Errorc(c, "activity-admin lottery runGiftTasks error. s.lotDao.UpdateGiftEffect() failed. error(%v)", err)
		return
	}
	if err = s.lotDao.UpdateGiftEffectDraftTx(c, tx, id, 1); err != nil {
		log.Errorc(c, "activity-admin lottery runGiftTasks error. s.lotDao.UpdateGiftEffect() failed. error(%v)", err)
		return
	}
	return nil
}
