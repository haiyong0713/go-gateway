package newyear2021

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	model "go-gateway/app/web-svr/activity/interface/model/newyear2021"
	"go-gateway/app/web-svr/activity/interface/rewards"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v notify.go rewards.go rewards_test.go service.go service_test.go task.go
func TestSendLetter(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, nil, testService.SendLetter(ctx, []int64{216761}, "this is test message."))
}

func TestSendAward(t *testing.T) {
	ctx := context.Background()
	testMid := int64(216761)
	rewards.Init(conf.Conf)
	_, err := component.GlobalDB.Exec(ctx, "drop table bnj2021_award_record_61 ")
	assert.Equal(t, nil, err)
	_, err = component.GlobalDB.Exec(ctx, `
create table if not exists bnj2021_award_record_61 (mid int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户id',
task_id int(11) unsigned NOT NULL DEFAULT '0' COMMENT '任务id',
receive_date DATE NOT NULL COMMENT '领取日期',
award_id int(11)  NOT NULL DEFAULT '0' COMMENT '奖励类型',
ctime datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
mtime datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
PRIMARY KEY (mid,task_id,receive_date),
KEY ix_mtime (mtime))
ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='拜年祭用户任务领取记录表';
`)
	assert.Equal(t, nil, err)
	_, err = testService.ReceiveAward(ctx, testMid, "TEST", &model.Task{
		DisplayName: "测试任务",
		Id:          0,
		AwardId:     0,
	}, true)
	assert.Equal(t, nil, err)
}
