package rewards

import (
	"context"
	"encoding/json"
	"flag"
	"testing"

	"go-common/library/cache/redis"

	"github.com/stretchr/testify/assert"

	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"
)

var testDao *Dao

func init() {
	flag.Set("conf", "../../cmd/activity-test.toml")
	if err := conf.Init(); err != nil {
		panic(err)
	}
	component.GlobalBnjCache = redis.NewPool(conf.Conf.Redis.Config)

	testDao = New(conf.Conf)

}

const (
	//拜年祭奖券
	AwardTypeLottery = "Bnj2021Lottery"
	//漫画折扣卷
	AwardTypeComicsCoupon = "Comics"
	//漫画卡
	AwardTypeComicsCard = "ComicsFreeCard"
	//直播弹幕
	AwardTypeLiveDanmaku = "Danmuku"
	//装扮套装
	AwardTypeGarbSuit = "GarbSuit"
	//装扮, 非套装
	AwardTypeGarbDressUp = "GarbDressUp"
	//会员购满减卷
	AwardTypeMallCoupon = "MallCoupon"
	//会员购商品
	AwardTypeMallPrize = "MallPrize"
	//大会员代金券
	AwardTypeVipCoupon = "VipCoupon"
	//魔晶
	AwardTypeMojing = "Mojing"

	//测试用, 直接返回成功
	AwardTypeDebug = "Debug"
	//活动平台Counter
	AwardTypeActCounter = "ActCounter"

	//课堂优惠券
	AwardTypeClassCoupon = "ClassCoupon"
)

func TestDao_AwardConfig(t *testing.T) {
	ctx := context.Background()
	//prepare data
	{
		//prepare data for user info
		_, err := testDao.db.Exec(ctx, "drop table if exists rewards_award_config")
		assert.Equal(t, nil, err)
		_, err = testDao.db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS rewards_award_config (
	id int(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '配置ID',
	activity_id int(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '关联活动ID',
	is_deleted tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 未删除 1 已删除',
	award_type varchar(20) NOT NULL DEFAULT "" COMMENT '奖励类型',
	display_name varchar(20) NOT NULL DEFAULT "" COMMENT '展示名称',
	icon_url varchar(100) NOT NULL DEFAULT "" COMMENT '奖品图标',
	should_send_notify tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 不发送通知 1 发送通知',
	notify_sender_id int(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '通知发送方ID',
	notify_code varchar(100) NOT NULL DEFAULT "" COMMENT '通知码',
	notify_message varchar(100) NOT NULL DEFAULT "" COMMENT '通知模板',
	notify_jump_url varchar(1000) NOT NULL DEFAULT "" COMMENT '通知跳转链接',
	notify_jump_url2 varchar(1000) NOT NULL DEFAULT "" COMMENT '通知跳转链接',
	config_content varchar(10000) NOT NULL DEFAULT "" COMMENT '配置内容',
	extra_info varchar(2000) NOT NULL DEFAULT "" COMMENT '自定义tag',
	ctime datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
	mtime datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
	PRIMARY KEY (id),
	KEY ix_mtime (mtime)
) ENGINE = InnoDB CHARSET = utf8 COMMENT '奖励奖品信息表';
`)
		assert.Equal(t, nil, err)

		_, err = testDao.db.Exec(ctx, "drop table if exists rewards_activity_config")
		assert.Equal(t, nil, err)
		_, err = testDao.db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS rewards_activity_config (
	id int(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '活动ID',
	is_deleted tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 未删除 1 已删除',
	name varchar(50) NOT NULL DEFAULT "" COMMENT '活动名称',
	notify_sender_id int(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '通知发送方ID',
	notify_code varchar(100) NOT NULL DEFAULT "" COMMENT '通知码',
	notify_message varchar(100) NOT NULL DEFAULT "" COMMENT '通知模板',
	notify_jump_url varchar(1000) NOT NULL DEFAULT "" COMMENT '通知跳转链接',
	notify_jump_url2 varchar(1000) NOT NULL DEFAULT "" COMMENT '通知跳转链接',
	ctime datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
	mtime datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
	PRIMARY KEY (id),
	KEY ix_mtime (mtime)
) ENGINE = InnoDB CHARSET = utf8 COMMENT '奖励活动信息表';
`)
		assert.Equal(t, nil, err)

		_, err = testDao.db.Exec(ctx, "drop table if exists rewards_entity_award_addresses")
		assert.Equal(t, nil, err)
		_, err = testDao.db.Exec(ctx, `
CREATE TABLE rewards_entity_award_addresses (
id int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
mid int(11) unsigned NOT NULL DEFAULT 0 COMMENT '用户id',
activity_id int(11) unsigned NOT NULL DEFAULT 0 COMMENT '关联活动ID',
activity_name varchar(50) NOT NULL DEFAULT '0' COMMENT '关联活动名称',
address_id int(11) unsigned NOT NULL DEFAULT 0 COMMENT '收货地址ID',
ctime datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
mtime datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp() COMMENT '修改时间',
PRIMARY KEY (id),
KEY ix_mtime (mtime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='实体奖励收货地址表';
`)
		assert.Equal(t, nil, err)
	}

	//configs should be empty
	configs, err := testDao.GetAwardMap(ctx, 0)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(configs))

	err = testDao.AddActivity(ctx, &model.AddActivityParam{
		Name:           "测试活动",
		NotifySenderId: 37090048,
		NotifyMessage:  "恭喜您在 {{ACTIVITY_NAME}} 活动中获得奖品: {{AWARD_NAME}}",
		NotifyJumpUrl:  "https://www.bilibili.com/",
	})
	assert.Equal(t, nil, err)
	as, err := testDao.ListActivity(ctx)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(as))
	aid := as[0].Id
	err = testDao.UpdateActivity(ctx, &model.UpdateActivityParam{
		Id:             aid,
		Name:           "我已经被修改了",
		NotifySenderId: 37090048,
		NotifyMessage:  "恭喜您在 {{ACTIVITY_NAME}} 活动中获得奖品: {{AWARD_NAME}}",
		NotifyJumpUrl:  "https://www.bilibili.com/",
	})
	assert.Equal(t, nil, err)
	as, err = testDao.ListActivity(ctx)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(as))
	assert.Equal(t, "我已经被修改了", as[0].Name)
	assert.Equal(t, nil, err)
	//添加奖品
	c1 := &model.ComicsCouponConfig{
		Type: 1}
	bs, _ := json.Marshal(c1)
	err = testDao.AddAward(ctx, &model.AddAwardParam{
		ActivityId:  aid,
		Type:        AwardTypeComicsCoupon,
		DisplayName: "漫画5折优惠券",
		//NotifySenderId: 0,
		//NotifyMessage:  "",
		//NotifyJumpUri1:  "",
		JsonStr:   string(bs),
		ExtraInfo: map[string]string{"haha": "haha"},
	})
	assert.Equal(t, nil, err)
	rs, err := testDao.GetAwardSlice(ctx, as[0].Id)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(rs))
	assert.Equal(t, false, rs[0].ShouldSendNotify)
	t.Logf("%+v", rs[0].ExtraInfo)
	err = testDao.UpdateAward(ctx, &model.UpdateAwardParam{
		Id:               as[0].Id,
		ActivityId:       aid,
		Type:             AwardTypeComicsCoupon,
		ShouldSendNotify: 1,
		DisplayName:      "漫画5折优惠券 -- 我也被修改过了",
		JsonStr:          string(bs),
		ExtraInfo:        map[string]string{"haha-123": "haha"},
	})
	assert.Equal(t, nil, err)
	rs, err = testDao.GetAwardSlice(ctx, as[0].Id)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(rs))
	assert.Equal(t, "漫画5折优惠券 -- 我也被修改过了", rs[0].DisplayName)
	//奖品没有单独设置Notify, 应当使用活动级别的配置
	assert.Equal(t, int64(37090048), rs[0].NotifySenderId)
	assert.Equal(t, true, rs[0].ShouldSendNotify)

	c2 := &model.DanmukuConfig{
		Color:      1,
		ExpireDays: 3,
		RoomIds:    []int64{0},
	}
	bs, _ = json.Marshal(c2)
	err = testDao.AddAward(ctx, &model.AddAwardParam{
		ActivityId:  aid,
		Type:        AwardTypeLiveDanmaku,
		DisplayName: "弹幕",
		JsonStr:     string(bs),
	})
	assert.Equal(t, nil, err)

	c3 := &model.SuitConfig{
		Id:         123,
		ExpireDays: 3,
	}
	bs, _ = json.Marshal(c3)
	err = testDao.AddAward(ctx, &model.AddAwardParam{
		ActivityId:  aid,
		Type:        AwardTypeGarbSuit,
		DisplayName: "装扮套装",
		JsonStr:     string(bs),
	})
	assert.Equal(t, nil, err)

	c4 := &model.DressUpConfig{
		Id:         234,
		ExpireDays: 30,
	}
	bs, _ = json.Marshal(c4)
	err = testDao.AddAward(ctx, &model.AddAwardParam{
		ActivityId:  aid,
		Type:        AwardTypeGarbDressUp,
		DisplayName: "装扮(非套装)",
		JsonStr:     string(bs),
	})
	assert.Equal(t, nil, err)

	c5 := &model.MallCouponConfig{
		SourceId: 10,
		CouponId: "81a5eb6e65770e2d",
	}
	bs, _ = json.Marshal(c5)
	err = testDao.AddAward(ctx, &model.AddAwardParam{
		ActivityId:  aid,
		Type:        AwardTypeMallCoupon,
		DisplayName: "会员购优惠券",
		JsonStr:     string(bs),
	})
	assert.Equal(t, nil, err)

	c11 := &model.MallCouponConfig{
		SourceId:         10,
		CouponId:         "e31bbf2740d6fbc0",
		SourceActivityID: "masteractivity",
	}
	bs, _ = json.Marshal(c11)
	err = testDao.AddAward(ctx, &model.AddAwardParam{
		ActivityId:  aid,
		Type:        AwardTypeMallCoupon,
		DisplayName: "魔晶",
		JsonStr:     string(bs),
	})
	assert.Equal(t, nil, err)

	c6 := &model.MallPrizeConfig{
		SourceId:    123,
		PrizeNo:     1,
		PrizePoolId: 2,
	}
	bs, _ = json.Marshal(c6)
	err = testDao.AddAward(ctx, &model.AddAwardParam{
		ActivityId:  aid,
		Type:        AwardTypeMallPrize,
		DisplayName: "会员购商品",
		JsonStr:     string(bs),
	})
	assert.Equal(t, nil, err)

	c7 := &model.VipCouponConfig{
		BatchToken: "test",
	}
	bs, _ = json.Marshal(c7)
	err = testDao.AddAward(ctx, &model.AddAwardParam{
		ActivityId:  aid,
		Type:        AwardTypeVipCoupon,
		DisplayName: "大会员优惠券",
		JsonStr:     string(bs),
	})
	assert.Equal(t, nil, err)

	c8 := &model.ActCounterConfig{
		Points:   1000,
		Source:   408933983,
		Activity: "happy2021",
		Extra:    "",
	}
	bs, _ = json.Marshal(c8)
	err = testDao.AddAward(ctx, &model.AddAwardParam{
		ActivityId:  aid,
		Type:        AwardTypeActCounter,
		DisplayName: "拜年祭积分*1000",
		JsonStr:     string(bs),
	})
	assert.Equal(t, nil, err)

	c9 := &model.ClassCouponConfig{
		BatchToken: "B20200909180432336918941",
		SendVc:     1,
		Sync:       true,
	}
	bs, _ = json.Marshal(c9)
	err = testDao.AddAward(ctx, &model.AddAwardParam{
		ActivityId:  aid,
		Type:        AwardTypeClassCoupon,
		DisplayName: "课堂优惠券",
		JsonStr:     string(bs),
	})
	assert.Equal(t, nil, err)

	configs, err = testDao.GetAwardMap(ctx, 0)
	assert.Equal(t, nil, err)
	assert.Equal(t, 10, len(configs))
	for _, c := range configs {
		t.Logf("%+v\n", c)
	}
	////delete all
	//err = testDao.DelActivity(ctx, aid)
	//assert.Equal(t, nil, err)
	//
	////configs should be empty
	//configs, err = testDao.GetAwardMap(ctx, 0)
	//assert.Equal(t, nil, err)
	//assert.Equal(t, 0, len(configs))
}
