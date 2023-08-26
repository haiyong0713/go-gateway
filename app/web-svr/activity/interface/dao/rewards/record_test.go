package rewards

import (
	"context"
	"fmt"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/interface/api"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//create table if not exists rewards_award_record_01 (mid int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户id',
//unique_id varchar(50)  NOT NULL DEFAULT '0' COMMENT '幂等ID',
//state tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 开始领取 1 领取完成 2 领取失败',
//award_id int(11) unsigned NOT NULL DEFAULT '0' COMMENT '奖励id',
//award_type varchar(20)  NOT NULL DEFAULT '0' COMMENT '奖励类型',
//award_name varchar(50)  NOT NULL DEFAULT '0' COMMENT '奖励名称',
//award_config_content varchar(10000) NOT NULL DEFAULT "" COMMENT '奖励配置内容'
//ctime timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//mtime timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
//PRIMARY KEY (mid,unique_id),
//KEY ix_m_u_s (mid,unique_id,state))
//KEY ix_mtime (mtime))
//ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='奖品用户发放记录表';

//create table if not exists rewards_award_fail_record (mid int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户id',
//unique_id varchar(50)  NOT NULL DEFAULT '0' COMMENT '幂等ID',
//error_msg varchar(2000)  NOT NULL DEFAULT '0' COMMENT '失败原因',
//retry_state tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 尚未重试 1 重试完成 ',
//ctime timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//mtime timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
//PRIMARY KEY (mid,unique_id),
//KEY ix_mtime (mtime))
//ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='奖品发放失败记录表';

func TestRecord(t *testing.T) {
	ctx := context.Background()
	{
		_, err := testDao.db.Exec(ctx, "drop table if exists rewards_award_record_61")
		assert.Equal(t, nil, err)
		_, err = testDao.db.Exec(ctx, `
create table if not exists rewards_award_record_61 (mid int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户id',
activity_id int(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '关联活动ID',
unique_id varchar(50)  NOT NULL DEFAULT '0' COMMENT '幂等ID',
business varchar(50)  NOT NULL DEFAULT '0' COMMENT '业务标识',
state tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 开始领取 1 领取完成 2 领取失败',
award_id int(11) unsigned NOT NULL DEFAULT '0' COMMENT '奖励id',
award_type varchar(20)  NOT NULL DEFAULT '0' COMMENT '奖励类型',
award_name varchar(50)  NOT NULL DEFAULT '0' COMMENT '奖励名称',
award_config_content varchar(10000) NOT NULL DEFAULT "" COMMENT '奖励配置内容',
ctime datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
mtime datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
PRIMARY KEY (mid,activity_id,unique_id),
KEY ix_mtime (mtime))
ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='奖品用户发放记录表';
`)
		assert.Equal(t, nil, err)

		_, err = testDao.db.Exec(ctx, "drop table if exists rewards_award_fail_record")
		assert.Equal(t, nil, err)
		_, err = testDao.db.Exec(ctx, `
create table if not exists rewards_award_fail_record (mid int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户id',
activity_id int(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '关联活动ID',
award_id int(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '关联奖品ID',
unique_id varchar(50)  NOT NULL DEFAULT '0' COMMENT '幂等ID',
error_msg varchar(2000)  NOT NULL DEFAULT '0' COMMENT '失败原因',
retry_state tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 尚未重试 1 重试完成 ',
ctime datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
mtime datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
PRIMARY KEY (mid,activity_id,unique_id),
KEY ix_mtime (mtime))
ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='奖品发放失败记录表';
`)
		assert.Equal(t, nil, err)
	}
	mid := int64(216761)
	aid := int64(0)
	as, err := testDao.GetAwardSlice(ctx, 1)
	assert.Equal(t, nil, err)
	successUniqueIds := make([]string, 0)
	for _, a := range as {
		aid = a.ActivityId
		uniqueId := fmt.Sprintf("%v-%v-%v", mid, a.Id, rand.Int())
		err = testDao.SendAwardByFunc(ctx, mid, uniqueId, "TEST", a, func() error {
			t.Logf("call send for %v", uniqueId)
			return nil
		})
		assert.Equal(t, nil, err)
		successUniqueIds = append(successUniqueIds, uniqueId)
	}

	failUniqueIds := make([]string, 0)

	for _, a := range as {
		time.Sleep(time.Second)
		uniqueId := fmt.Sprintf("%v-%v-%v", mid, a.Id, rand.Int())
		err = testDao.SendAwardByFunc(ctx, mid, uniqueId, "TEST", a, func() error {
			t.Logf("call send for %v", uniqueId)
			return fmt.Errorf("send mid: %v, awardId:%v: fake_error_message", mid, a.Id)
		})
		assert.NotEqual(t, nil, err)
		failUniqueIds = append(failUniqueIds, uniqueId)
	}

	for _, id := range successUniqueIds {
		send, err := testDao.IsAwardAlreadySend(ctx, mid, aid, id)
		assert.Equal(t, nil, err)
		assert.Equal(t, true, send)
		assert.Equal(t, nil, err)
	}
	am, err := testDao.GetAwardMap(ctx, 0)
	b, err := testDao.IsActivityAddressExists(ctx, mid, aid)
	assert.Equal(t, nil, err)
	assert.Equal(t, int64(0), b)
	assert.Equal(t, nil, err)
	err = testDao.AddActivityAddress(ctx, mid, am[aid], 1)
	assert.Equal(t, nil, err)
	b, err = testDao.IsActivityAddressExists(ctx, mid, aid)
	assert.Equal(t, nil, err)
	assert.Equal(t, int64(1), b)

	for _, id := range failUniqueIds {
		send, err := testDao.IsAwardAlreadySend(ctx, mid, aid, id)
		assert.Equal(t, nil, err)
		assert.Equal(t, true, send)
	}

	res, err := testDao.GetAwardRecordByMidAndActivityIdsWithCache(ctx, mid, []int64{aid}, 5)
	assert.Equal(t, nil, err)
	lastTimeStamp := xtime.Time(0)
	for _, r := range res {
		t.Logf("%+v", r)
		assert.Equal(t, true, lastTimeStamp != 0 || r.SentTime <= lastTimeStamp)

		lastTimeStamp = r.SentTime
	}
	count, err := testDao.GetAwardCountByMidAndActivityIdWithCache(ctx, mid, aid)
	assert.Equal(t, nil, err)
	t.Logf("record count is %v\n", count)
	res, err = testDao.GetAwardRecordByMidAndActivityIdsWithCache(ctx, mid, []int64{aid}, 5)
	res, err = testDao.GetAwardRecordByMidAndActivityIdsWithCache(ctx, mid, []int64{aid}, 5)
	res, err = testDao.GetAwardRecordByMidAndActivityIdsWithCache(ctx, mid, []int64{aid}, 5)
	res, err = testDao.GetAwardRecordByMidAndActivityIdsWithCache(ctx, mid, []int64{aid}, 5)
	res, err = testDao.GetAwardRecordByMidAndActivityIdsWithCache(ctx, mid, []int64{aid}, 5)
	res, err = testDao.GetAwardRecordByMidAndActivityIdsWithCache(ctx, mid, []int64{88}, 5)
	res, err = testDao.GetAwardRecordByMidAndActivityIdsWithCache(ctx, mid, []int64{88}, 5)
	res, err = testDao.GetAwardRecordByMidAndActivityIdsWithCache(ctx, mid, []int64{88}, 5)
	res, err = testDao.GetAwardRecordByMidAndActivityIdsWithCache(ctx, mid, []int64{88}, 5)

	testDao.AddSingleRecordCache(mid, aid, &api.RewardsSendAwardReply{
		Mid:          mid,
		AwardId:      0,
		Name:         "AddSingleRecordCache",
		ActivityId:   aid,
		ActivityName: "",
		Type:         "",
		Icon:         "",
		ReceiveTime:  0,
		ExtraInfo:    nil,
	})

	newCount, err := testDao.GetAwardCountByMidAndActivityIdWithCache(ctx, mid, aid)
	assert.Equal(t, nil, err)
	t.Logf("record count is %v\n", newCount)

}
