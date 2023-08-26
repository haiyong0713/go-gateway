package lottery

import (
	"context"
	"encoding/json"
	xtime "go-common/library/time"
	"testing"

	"go-gateway/app/web-svr/activity/ecode"
	lottery "go-gateway/app/web-svr/activity/interface/dao/lottery_v2"
	l "go-gateway/app/web-svr/activity/interface/model/lottery_v2"

	// passportinfoapi "git.bilibili.co/bapis/bapis-go/passport/service/user"
	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	passportinfoapi "git.bilibili.co/bapis/bapis-go/passport/service/user"
	silverbulletapi "git.bilibili.co/bapis/bapis-go/silverbullet/service/silverbullet-proxy"
)

const (
	testRegStime           = 0
	testRegEtime           = 0
	testSid                = "11111"
	testMid                = 2090946683
	testLevel1             = 1
	testLevel0             = 0
	testVipNeedVipCheck    = 1
	testVipNeedMonthCheck  = 2
	testVipNeedAnnualCheck = 3
	testVipNoNeedCheck     = 0
	testAccountNoNeedCheck = 0
	testAccountMobileCheck = 1
	testAccountNameCheck   = 2
	testCoinConsume10      = 10
	testCoinConsume0       = 0
	testFsIPOpen           = 1
	testFsIPClose          = 0
	testHighRate           = 1
	testHighTypeClose      = 0
	testHighTypeBuyVip     = 1
	testHighTypeArchive    = 2
	testRateMustWin        = 1
	testRateWin10          = 10

	testTimesBaseCID    = 1
	testTimesWinCID     = 2
	testTimesShareCID   = 3
	testTimesFollowCID  = 4
	testTimesArchiveCID = 5
	testTimesBuyVipCID  = 6
	testTimesOtherCID   = 7
	testTimesVipCID     = 8
	testTimesOGVCID     = 9
	testTimesFeCID      = 10

	testTimesAddTypeDaily = 1
	testTimesAddTypeAll   = 0

	testGiftLeastMark   = 1
	testGiftNoLeastMark = 0

	testGiftVipParams       = "{\"token\":\"QbxkxWgMHzoarghV\",\"app_key\":\"qnVSWrciDTPVHXpT\"}"
	testGiftGrantParams     = "{\"pid\":444,\"expire\":7}"
	testGiftCoinParams      = "{\"coin\":222}"
	testGiftVipCouponParams = "{\"token\":\"873673034320200520181130\",\"app_key\":\"222222\"}"
	testGiftOGVParams       = "{\"token\":\"111\"}"
	testGiftVipBuyParams    = "{\"token\":\"222\"}"
	testGiftMoneyParams     = "{\"money\":11,\"customer_id\":\"10086\",\"activity_id\":\"537b72175df741e2cc2d4e7a8eabccfb\",\"trans_desc\":\"活动红包\",\"start_time\":111}"
)

func TestDoLotteryBase(t *testing.T) {
	Convey("test do lottery from db", t, WithService(func(s *Service) {
		mockCtl := NewController(t)

		mockDao := lottery.NewMockDao(mockCtl)

		defer mockCtl.Finish()

		// mock

		mockDao.EXPECT().CacheQPSLimit(Any(), Any()).Return(int64(1), nil)

		var stime, etime xtime.Time
		stime = 1596255539
		etime = 1817093939
		lottery := testLotteryBase(1, stime, etime)
		mockDao.EXPECT().CacheLottery(Any(), Any()).Return(nil, ecode.ActivityLotteryRiskInfo)
		mockDao.EXPECT().RawLottery(Any(), Any()).Return(lottery, nil)
		mockDao.EXPECT().AddCacheLottery(Any(), Any(), Any()).Return(nil)

		info := testLotteryInfo(testSid, testLevel0, testRegStime, testRegEtime, testVipNoNeedCheck, testAccountNoNeedCheck,
			testCoinConsume0, testFsIPClose, testHighTypeBuyVip, testHighRate, testRateWin10)
		mockDao.EXPECT().CacheLotteryInfo(Any(), Any()).Return(nil, ecode.ActivityLotteryRiskInfo)
		mockDao.EXPECT().RawLotteryInfo(Any(), Any()).Return(info, nil)
		mockDao.EXPECT().AddCacheLotteryInfo(Any(), Any(), Any()).Return(nil)

		timesConf := testTimesBatch()
		mockDao.EXPECT().CacheLotteryTimesConfig(Any(), Any()).Return(nil, ecode.ActivityLotteryRiskInfo)
		mockDao.EXPECT().RawLotteryTimesConfig(Any(), Any()).Return(timesConf, nil)
		mockDao.EXPECT().AddCacheLotteryTimesConfig(Any(), Any(), Any()).Return(nil)

		gift := testAllGift()
		mockDao.EXPECT().CacheLotteryGift(Any(), Any()).Return(nil, ecode.ActivityLotteryRiskInfo)
		mockDao.EXPECT().RawLotteryGift(Any(), Any()).Return(gift, nil)
		mockDao.EXPECT().AddCacheLotteryGift(Any(), Any(), Any()).Return(nil)

		memberGroup := testMemberGroupAll()
		mockDao.EXPECT().CacheMemberGroup(Any(), Any()).Return(nil, ecode.ActivityLotteryRiskInfo)
		mockDao.EXPECT().RawMemberGroup(Any(), Any()).Return(memberGroup, nil)
		mockDao.EXPECT().AddCacheMemberGroup(Any(), Any(), Any()).Return(nil)

		addTimes := testAddTimesDB()
		mockDao.EXPECT().CacheLotteryTimes(Any(), Any(), Any(), Any()).Return(nil, ecode.ActivityLotteryRiskInfo)
		mockDao.EXPECT().RawLotteryAddTimes(Any(), Any(), Any()).Return(addTimes, nil)
		mockDao.EXPECT().AddCacheLotteryTimes(Any(), Any(), Any(), Any(), Any()).Return(nil)

		usedTimes := testUsedTimesDB()
		mockDao.EXPECT().CacheLotteryTimes(Any(), Any(), Any(), Any()).Return(nil, ecode.ActivityLotteryRiskInfo)
		mockDao.EXPECT().RawLotteryUsedTimes(Any(), Any(), Any()).Return(usedTimes, nil)
		mockDao.EXPECT().AddCacheLotteryTimes(Any(), Any(), Any(), Any(), Any()).Return(nil)

		mockDao.EXPECT().CacheLotteryMcNum(Any(), Any(), Any(), Any()).Return(int64(10), nil)
		mockDao.EXPECT().AddCacheLotteryMcNum(Any(), Any(), Any(), Any(), Any()).Return(nil)

		sendDayNum := testGetDaySendGiftNum()
		mockDao.EXPECT().CacheSendDayGiftNum(Any(), Any(), Any(), Any()).Return(sendDayNum, nil)
		sendNum := testGetSendGiftNum()
		mockDao.EXPECT().CacheSendGiftNum(Any(), Any(), Any()).Return(sendNum, nil)

		mockDao.EXPECT().IncrGiftSendDayNum(Any(), Any(), Any(), Any()).Return(testGetDaySendGiftNum1(), nil)
		// mockDao.EXPECT().IncrGiftSendDayNum(Any(), Any(), Any(), Any()).Return(testGetDaySendGiftNum1(), nil)

		mockDao.EXPECT().IncrGiftSendNum(Any(), Any(), Any()).Return(testGetSendGiftNum1(), nil)
		mockDao.EXPECT().UpdatelotteryGiftNumSQL(Any(), Any(), Any()).Return(int64(2), nil)

		mockDao.EXPECT().IncrTimes(Any(), Any(), Any(), Any(), Any()).Return(nil)

		mockDao.EXPECT().InsertLotteryRecard(Any(), Any(), Any(), Any(), Any()).Return(int64(2), nil)
		mockDao.EXPECT().DeleteLotteryActionLog(Any(), Any(), Any()).Return(nil)

		mockDao.EXPECT().InsertLotteryWin(Any(), Any(), Any(), Any(), Any()).Return(int64(2), nil)
		mockDao.EXPECT().InsertLotteryWin(Any(), Any(), Any(), Any(), Any()).Return(int64(2), nil)
		s.lottery = mockDao

		res, err := s.DoLottery(context.Background(), testSid, testMid, &l.FrontEndParams{
			IP:       "127.0.0.1",
			DeviceID: "fdsfdsfds",
			Ua:       "",
			API:      "",
			Referer:  "",
		}, 1, true)

		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))

}

func testGetSendGiftNum() map[int64]int64 {
	sendGiftNum := make(map[int64]int64)
	sendGiftNum[1] = 0
	sendGiftNum[2] = 0
	sendGiftNum[3] = 0
	sendGiftNum[4] = 0
	sendGiftNum[5] = 0
	sendGiftNum[6] = 0
	sendGiftNum[7] = 0
	sendGiftNum[8] = 0
	sendGiftNum[9] = 0
	sendGiftNum[10] = 0
	return sendGiftNum
}
func testGetDaySendGiftNum() map[string]int64 {
	sendDayGiftNum := make(map[string]int64)
	sendDayGiftNum["1_store"] = 9
	sendDayGiftNum["2_store"] = 9
	sendDayGiftNum["3_store"] = 9
	sendDayGiftNum["4_store"] = 9
	sendDayGiftNum["5_store"] = 0
	sendDayGiftNum["6_store"] = 9
	sendDayGiftNum["7_store"] = 9
	sendDayGiftNum["8_store"] = 9
	sendDayGiftNum["9_store"] = 9
	sendDayGiftNum["10_store"] = 9
	return sendDayGiftNum
}

func testGetDaySendGiftNum1() map[string]int64 {
	sendDayGiftNum := make(map[string]int64)
	sendDayGiftNum["2_store"] = 9
	return sendDayGiftNum
}

func testGetSendGiftNum1() map[int64]int64 {
	sendGiftNum := make(map[int64]int64)
	sendGiftNum[2] = 9
	return sendGiftNum
}

// testLotteryBase
func testLotteryBase(isInternal int, stime, etime xtime.Time) *l.Lottery {
	return &l.Lottery{
		ID:         1,
		LotteryID:  "222",
		Name:       "name",
		IsInternal: isInternal,
		Stime:      stime,
		Etime:      etime,
		Type:       1,
		State:      0,
	}

}

// testLotteryInfo
func testLotteryInfo(sid string, level int, regTimeStime, regTimeEtime int64, vipCheck, accountCheck, coin, fsip, highType int, highRate, giftRate int64) *l.Info {
	return &l.Info{
		ID:           1,
		Sid:          sid,
		Level:        level,
		RegTimeStime: regTimeStime,
		RegTimeEtime: regTimeEtime,
		VipCheck:     vipCheck,
		AccountCheck: accountCheck,
		Coin:         coin,
		FsIP:         fsip,
		GiftRate:     giftRate,
		HighType:     highType,
		HighRate:     highRate,
	}

}

func testTimesBatch() []*l.TimesConfig {
	timesBatch := make([]*l.TimesConfig, 0)
	timesBatch = append(timesBatch, testTimesConfigBase(testTimesAddTypeAll, 10, 2))
	timesBatch = append(timesBatch, testTimesConfigShare(testTimesAddTypeAll, 2, 3))
	timesBatch = append(timesBatch, testTimesConfigWin(testTimesAddTypeAll, 10000, 10000))
	// timesBatch = append(timesBatch, testTimesConfigBuyVip(testTimesAddTypeDaily, 10, 10))
	return timesBatch
}

func testTimesConfigBase(addType, times, most int) *l.TimesConfig {
	return testTimesConfig(testTimesBaseCID, testSid, testTimesBaseCID, addType, times, most)
}

func testTimesConfigWin(addType, times, most int) *l.TimesConfig {
	return testTimesConfig(testTimesWinCID, testSid, testTimesWinCID, addType, times, most)
}

func testTimesConfigShare(addType, times, most int) *l.TimesConfig {
	return testTimesConfig(testTimesShareCID, testSid, testTimesShareCID, addType, times, most)
}

func testTimesConfigArchive(addType, times, most int) *l.TimesConfig {
	return testTimesConfig(testTimesArchiveCID, testSid, testTimesArchiveCID, addType, times, most)
}

func testTimesConfigBuyVip(addType, times, most int) *l.TimesConfig {
	return testTimesConfig(testTimesBuyVipCID, testSid, testTimesBuyVipCID, addType, times, most)
}

func testTimesConfig(id int64, sid string, timesType, addType, times, most int) *l.TimesConfig {
	return &l.TimesConfig{
		ID:      id,
		Sid:     sid,
		Type:    timesType,
		AddType: addType,
		Times:   times,
		Most:    most,
	}
}

func testAddTimesShare() *l.AddTimes {
	var ctime xtime.Time
	ctime = 1596255539
	return testAddTimes(testMid, testTimesShareCID, 2, testTimesShareCID, ctime)
}

func testAddTimesBuyVip() *l.AddTimes {
	var ctime xtime.Time
	ctime = 1596345543
	return testAddTimes(testMid, testTimesBuyVipCID, 2, testTimesBuyVipCID, ctime)
}
func testAddTimes(mid int64, addType int, num int, cid int64, ctime xtime.Time) *l.AddTimes {
	return &l.AddTimes{
		Mid:   mid,
		Type:  addType,
		Num:   num,
		CID:   cid,
		Ctime: ctime,
	}

}

func testUsedTimesShare() *l.RecordDetail {
	var ctime xtime.Time
	ctime = 1596255539
	return testUsedTimes(testMid, testTimesShareCID, 1, 0, testTimesShareCID, 0, ctime)
}
func testUsedTimesBuyVip() *l.RecordDetail {
	var ctime xtime.Time
	ctime = 1596255539
	return testUsedTimes(testMid, testTimesBuyVipCID, 1, 1, testTimesBuyVipCID, 1, ctime)

}
func testUsedTimes(mid int64, cid int64, num int, giftID int64, timesType int, giftType int, ctime xtime.Time) *l.RecordDetail {
	return &l.RecordDetail{
		Mid:      mid,
		Num:      num,
		GiftID:   giftID,
		GiftType: giftType,
		Type:     timesType,
		Ctime:    ctime,
		CID:      cid,
	}
}

func testAllGift() []*l.GiftDB {
	var daySendNum, dayNum string
	daySendNum = "{\"store\":8}"
	dayNum = "{\"store\":9}"
	giftBatch := make([]*l.GiftDB, 0)
	giftBatch = append(giftBatch, testGiftMatrial(1, testSid, 10, "", "http", testGiftNoLeastMark, "", "", 9, daySendNum, dayNum, "", 1000, ""))
	giftBatch = append(giftBatch, testGiftVip(2, testSid, 10, "", "http", testGiftNoLeastMark, "", "", 9, daySendNum, dayNum, "", 1000, testGiftVipParams))
	giftBatch = append(giftBatch, testGiftGrant(3, testSid, 10, "", "http", testGiftNoLeastMark, "", "", 9, daySendNum, dayNum, "", 1000, testGiftGrantParams))
	giftBatch = append(giftBatch, testGiftCoupon(4, testSid, 10, "", "http", testGiftNoLeastMark, "", "", 9, daySendNum, dayNum, "", 1000, ""))
	giftBatch = append(giftBatch, testGiftCoin(5, testSid, 10, "", "http", testGiftLeastMark, "", "", 9, daySendNum, dayNum, "", 1000, testGiftCoinParams))
	giftBatch = append(giftBatch, testGiftVipCoupon(6, testSid, 10, "", "http", testGiftNoLeastMark, "", "", 9, daySendNum, dayNum, "", 1000, testGiftVipCouponParams))
	giftBatch = append(giftBatch, testGiftOther(7, testSid, 10, "", "http", testGiftNoLeastMark, "", "", 9, daySendNum, dayNum, "", 1000, ""))
	giftBatch = append(giftBatch, testGiftOGV(8, testSid, 10, "", "http", testGiftNoLeastMark, "", "", 9, daySendNum, dayNum, "", 1000, testGiftOGVParams))
	giftBatch = append(giftBatch, testGiftVipBuy(9, testSid, 10, "", "http", testGiftNoLeastMark, "", "", 9, daySendNum, dayNum, "", 1000, testGiftVipBuyParams))
	giftBatch = append(giftBatch, testGiftMoney(10, testSid, 10, "", "http", testGiftNoLeastMark, "", "", 9, daySendNum, dayNum, "", 1000, testGiftMoneyParams))
	return giftBatch
}

func testGiftMatrial(id int64, sid string, num int64, source string, imgURL string, leastMark int,
	messageTitle, messageContent string, sendNum int64, daySendNum, dayNum string, memberGroup string, probability int, params string) *l.GiftDB {
	return &l.GiftDB{
		ID:           id,
		Sid:          sid,
		Name:         "实物奖",
		Num:          num,
		Type:         1,
		Source:       source,
		ImgURL:       imgURL,
		LeastMark:    leastMark,
		MessageTitle: messageTitle,
		Efficient:    1,

		MessageContent: messageContent,
		SendNum:        sendNum,
		DaySendNum:     daySendNum,
		MemberGroup:    memberGroup,
		DayNum:         dayNum,
		Probability:    probability,
		Params:         params,
	}
}

func testGiftVip(id int64, sid string, num int64, source string, imgURL string, leastMark int,
	messageTitle, messageContent string, sendNum int64, daySendNum, dayNum string, memberGroup string, probability int, params string) *l.GiftDB {
	return &l.GiftDB{
		ID:             id,
		Sid:            sid,
		Name:           "大会员奖",
		Num:            num,
		Type:           2,
		Source:         source,
		ImgURL:         imgURL,
		LeastMark:      leastMark,
		MessageTitle:   messageTitle,
		MessageContent: messageContent,
		SendNum:        sendNum,
		Efficient:      1,
		DaySendNum:     daySendNum,
		MemberGroup:    memberGroup,
		DayNum:         dayNum,
		Probability:    probability,
		Params:         params,
	}
}

func testGiftGrant(id int64, sid string, num int64, source string, imgURL string, leastMark int,
	messageTitle, messageContent string, sendNum int64, daySendNum, dayNum string, memberGroup string, probability int, params string) *l.GiftDB {
	return &l.GiftDB{
		ID:             id,
		Sid:            sid,
		Name:           "头像挂件",
		Num:            num,
		Type:           3,
		Source:         source,
		ImgURL:         imgURL,
		LeastMark:      leastMark,
		MessageTitle:   messageTitle,
		MessageContent: messageContent,
		SendNum:        sendNum,
		DaySendNum:     daySendNum,
		MemberGroup:    memberGroup,
		Efficient:      1,
		DayNum:         dayNum,
		Probability:    probability,
		Params:         params,
	}
}

func testGiftCoupon(id int64, sid string, num int64, source string, imgURL string, leastMark int,
	messageTitle, messageContent string, sendNum int64, daySendNum, dayNum string, memberGroup string, probability int, params string) *l.GiftDB {
	return &l.GiftDB{
		ID:             id,
		Sid:            sid,
		Name:           "优惠券",
		Num:            num,
		Type:           4,
		Source:         source,
		ImgURL:         imgURL,
		LeastMark:      leastMark,
		MessageTitle:   messageTitle,
		MessageContent: messageContent,
		SendNum:        sendNum,
		Efficient:      1,
		DaySendNum:     daySendNum,
		MemberGroup:    memberGroup,
		DayNum:         dayNum,
		Probability:    probability,
		Params:         params,
	}
}

func testGiftCoin(id int64, sid string, num int64, source string, imgURL string, leastMark int,
	messageTitle, messageContent string, sendNum int64, daySendNum, dayNum string, memberGroup string, probability int, params string) *l.GiftDB {
	return &l.GiftDB{
		ID:             id,
		Sid:            sid,
		Name:           "硬币",
		Num:            num,
		Type:           5,
		Source:         source,
		ImgURL:         imgURL,
		LeastMark:      leastMark,
		Efficient:      1,
		MessageTitle:   messageTitle,
		MessageContent: messageContent,
		SendNum:        sendNum,
		DaySendNum:     daySendNum,
		MemberGroup:    memberGroup,
		DayNum:         dayNum,
		Probability:    probability,
		Params:         params,
	}
}

func testGiftVipCoupon(id int64, sid string, num int64, source string, imgURL string, leastMark int,
	messageTitle, messageContent string, sendNum int64, daySendNum, dayNum string, memberGroup string, probability int, params string) *l.GiftDB {
	return &l.GiftDB{
		ID:             id,
		Sid:            sid,
		Name:           "大会员抵用券",
		Num:            num,
		Type:           6,
		Source:         source,
		ImgURL:         imgURL,
		LeastMark:      leastMark,
		MessageTitle:   messageTitle,
		MessageContent: messageContent,
		SendNum:        sendNum,
		DaySendNum:     daySendNum,
		Efficient:      1,
		MemberGroup:    memberGroup,
		DayNum:         dayNum,
		Probability:    probability,
		Params:         params,
	}
}

func testGiftOther(id int64, sid string, num int64, source string, imgURL string, leastMark int,
	messageTitle, messageContent string, sendNum int64, daySendNum, dayNum string, memberGroup string, probability int, params string) *l.GiftDB {
	return &l.GiftDB{
		ID:           id,
		Sid:          sid,
		Name:         "其他",
		Num:          num,
		Type:         7,
		Source:       source,
		ImgURL:       imgURL,
		LeastMark:    leastMark,
		MessageTitle: messageTitle,
		Efficient:    1,

		MessageContent: messageContent,
		SendNum:        sendNum,
		DaySendNum:     daySendNum,
		MemberGroup:    memberGroup,
		DayNum:         dayNum,
		Probability:    probability,
		Params:         params,
	}
}

func testGiftOGV(id int64, sid string, num int64, source string, imgURL string, leastMark int,
	messageTitle, messageContent string, sendNum int64, daySendNum, dayNum string, memberGroup string, probability int, params string) *l.GiftDB {
	return &l.GiftDB{
		ID:             id,
		Sid:            sid,
		Name:           "OGV",
		Num:            num,
		Type:           8,
		Source:         source,
		Efficient:      1,
		ImgURL:         imgURL,
		LeastMark:      leastMark,
		MessageTitle:   messageTitle,
		MessageContent: messageContent,
		SendNum:        sendNum,
		DaySendNum:     daySendNum,
		MemberGroup:    memberGroup,
		DayNum:         dayNum,
		Probability:    probability,
		Params:         params,
	}
}

func testGiftVipBuy(id int64, sid string, num int64, source string, imgURL string, leastMark int,
	messageTitle, messageContent string, sendNum int64, daySendNum, dayNum string, memberGroup string, probability int, params string) *l.GiftDB {
	return &l.GiftDB{
		ID:             id,
		Sid:            sid,
		Name:           "会员购",
		Num:            num,
		Type:           9,
		Source:         source,
		ImgURL:         imgURL,
		LeastMark:      leastMark,
		MessageTitle:   messageTitle,
		Efficient:      1,
		MessageContent: messageContent,
		SendNum:        sendNum,
		DaySendNum:     daySendNum,
		MemberGroup:    memberGroup,
		DayNum:         dayNum,
		Probability:    probability,
		Params:         params,
	}
}

func testGiftMoney(id int64, sid string, num int64, source string, imgURL string, leastMark int,
	messageTitle, messageContent string, sendNum int64, daySendNum, dayNum string, memberGroup string, probability int, params string) *l.GiftDB {
	return &l.GiftDB{
		ID:             id,
		Sid:            sid,
		Name:           "现金",
		Num:            num,
		Type:           10,
		Source:         source,
		ImgURL:         imgURL,
		LeastMark:      leastMark,
		MessageTitle:   messageTitle,
		MessageContent: messageContent,
		SendNum:        sendNum,
		Efficient:      1,

		DaySendNum:  daySendNum,
		MemberGroup: memberGroup,
		DayNum:      dayNum,
		Probability: probability,
		Params:      params,
	}
}

func testMemberGroupAll() []*l.MemberGroupDB {
	memberGroup := make([]*l.MemberGroupDB, 0)
	group := make([]*l.Group, 0)
	oldGroup := make([]*l.Group, 0)
	group = append(group, testGroupNew(), testGroupNew7200(), testGroupVipMonth())
	oldGroup = append(oldGroup, testGroupOld(), testGroupVipMonth())
	memberGroup = append(memberGroup, testMemberGroup(1, "1小时内注册的新用户+月度大会员", group))
	memberGroup = append(memberGroup, testMemberGroup(2, "1小时外注册的老用户+月度大会员", oldGroup))
	return memberGroup
}

func testGroupNew7200() *l.Group {
	params := make(map[string]interface{})
	params["period"] = 7200
	params["is_new"] = 1
	return &l.Group{
		GroupType: 1,
		Params:    params,
	}
}
func testGroupNew() *l.Group {
	params := make(map[string]interface{})
	params["period"] = 3600
	params["is_new"] = 1
	return &l.Group{
		GroupType: 1,
		Params:    params,
	}
}

func testGroupOld() *l.Group {
	params := make(map[string]interface{})
	params["period"] = 3600
	params["is_new"] = 2
	return &l.Group{
		GroupType: 1,
		Params:    params,
	}
}

func testGroupVipMonth() *l.Group {
	params := make(map[string]interface{})
	params["vip_type"] = l.VipTypeMonth
	return &l.Group{
		GroupType: 2,
		Params:    params,
	}
}

func testMemberGroup(id int64, name string, group []*l.Group) *l.MemberGroupDB {
	groupB, _ := json.Marshal(group)
	return &l.MemberGroupDB{
		ID:    id,
		Name:  name,
		Group: string(groupB),
	}
}

func testMockRiskInfo() *silverbulletapi.RiskInfoReply {
	info := make(map[string]*silverbulletapi.RiskInfo)
	info["activity_strategy"] = &silverbulletapi.RiskInfo{Level: 0}
	return &silverbulletapi.RiskInfoReply{
		Infos: info,
	}
}

func testMockCheckResh() *passportinfoapi.CheckFreshUserReply {
	return &passportinfoapi.CheckFreshUserReply{
		IsNew: true,
	}
}

func testUserUsedTimesRedis() map[string]int {
	usedTimes := make(map[string]int)
	return usedTimes
}

func testUserAddTimesRedis() map[string]int {
	addTimes := make(map[string]int)
	return addTimes
}

func testAddTimesDB() []*l.AddTimes {
	addTimes := make([]*l.AddTimes, 0)
	addTimes = append(addTimes, testAddTimesShare())
	addTimes = append(addTimes, testAddTimesBuyVip())
	return addTimes
}

func testUsedTimesDB() []*l.RecordDetail {
	usedTimes := make([]*l.RecordDetail, 0)
	usedTimes = append(usedTimes, testUsedTimesShare())
	usedTimes = append(usedTimes, testUsedTimesBuyVip())
	return usedTimes
}
