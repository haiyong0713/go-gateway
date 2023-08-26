package newyear2021

import (
	"context"
	"encoding/json"
	"testing"

	"go-gateway/app/web-svr/activity/interface/component"
	model "go-gateway/app/web-svr/activity/interface/model/newyear2021"

	"github.com/stretchr/testify/assert"
)

func TestDao_Config(t *testing.T) {
	ctx := context.Background()
	//prepare data
	{
		//prepare data for user info
		_, err := component.GlobalBnjDB.Exec(ctx, "drop table bnj2021_config")
		assert.Equal(t, nil, err)
		_, err = component.GlobalBnjDB.Exec(ctx, `
CREATE TABLE IF NOT EXISTS bnj2021_config (
	id int(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '配置版本',
	is_deleted tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 未删除 1 已删除',
	config_content varchar(10000) NOT NULL DEFAULT "" COMMENT '配置内容',
	ctime datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
	mtime datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
	PRIMARY KEY (id),
	KEY ix_mtime (mtime)
) ENGINE = InnoDB CHARSET = utf8 COMMENT '拜年祭配置信息表';
`)
		assert.Equal(t, nil, err)
	}

	//init config
	config := &model.Config{}
	assert.Equal(t, nil, json.Unmarshal([]byte(configContent), &config))
	assert.Equal(t, nil, testDao.UpdateConf(ctx, config))
	config = &model.Config{}
	version, config, err := testDao.GetLatestConf(ctx)
	assert.Equal(t, nil, err)
	assert.Equal(t, int64(1), version)
	assert.Equal(t, int64(15000), config.TaskConfig.GlobalTasks[1].Stages[2].RequireCount)

	for _, s := range config.TaskConfig.GlobalTasks {
		for _, st := range s.Stages {
			t.Logf("%v stage %v awardId: %v", s.DisplayName, st.DisplayName, st.AwardId)
		}
	}

	for _, s := range config.TaskConfig.DailyTasks.Tasks {
		t.Logf("dailyTask %v awardId: %v", s.DisplayName, s.AwardId)
	}

	for _, s := range config.TaskConfig.LevelTask.Stages {
		t.Logf("levelTasks stage %v awardId: %v", s.DisplayName, s.AwardId)
	}

	//update config
	config.TaskConfig.GlobalTasks[1].Stages[2].RequireCount = 16000
	assert.Equal(t, nil, testDao.UpdateConf(ctx, config))
	version, config, err = testDao.GetLatestConf(ctx)
	assert.Equal(t, nil, err)
	assert.Equal(t, int64(2), version)
	assert.Equal(t, int64(16000), config.TaskConfig.GlobalTasks[1].Stages[2].RequireCount)

	//delete config
	err = testDao.DeleteConf(ctx, 2)
	assert.Equal(t, nil, err)
	version, config, err = testDao.GetLatestConf(ctx)
	assert.Equal(t, nil, err)
	assert.Equal(t, int64(1), version)
	assert.Equal(t, int64(15000), config.TaskConfig.GlobalTasks[1].Stages[2].RequireCount)
}

var configContent = `
{
    "TimePeriod":{
        "Start":1605785589,
        "End":1609430400
    },
    "TaskConfig":{
        "DailyTasks":{
            "DisplayName":"每日游戏额外奖励",
            "Id":101,
            "AwardId":111,
            "Tasks":[
                {
                    "DisplayName":"每日游戏任务",
                    "Id":101,
                    "ActPlatId":"happy2021",
                    "ActPlatCounterId":"game",
                    "RequireCount":3,
                    "AwardId":2
                },
                {
                    "DisplayName":"每日视频任务",
                    "Id":102,
                    "ActPlatId":"happy2021",
                    "ActPlatCounterId":"view",
                    "RequireCount":3,
                    "AwardId":3
                },
                {
                    "DisplayName":"每日分享任务",
                    "Id":103,
                    "ActPlatId":"happy2021",
                    "ActPlatCounterId":"share",
                    "RequireCount":3,
                    "AwardId":4
                },
                {
                    "DisplayName":"每日投稿任务",
                    "Id":104,
                    "ActPlatId":"happy2021",
                    "ActPlatCounterId":"videoup",
                    "RequireCount":3,
                    "AwardId":5
                }
            ]
        },
        "GlobalTasks":[
            {
                "DisplayName":"预约人数任务",
                "Id":211,
                "ActPlatId":"happy2021",
                "ActPlatCounterId":"view",
                "Stages":[
                    {
                        "DisplayName":"预约人数到达5000",
                        "Id":201,
                        "RequireCount":5000,
                        "AwardId":6
                    },
                    {
                        "DisplayName":"预约人数到达10000",
                        "Id":202,
                        "RequireCount":10000,
                        "AwardId":7
                    },
                    {
                        "DisplayName":"预约人数到达15000",
                        "Id":203,
                        "RequireCount":15000,
                        "AwardId":8
                    }
                ]
            },
            {
                "DisplayName":"参与打年兽人数任务",
                "Id":231,
                "ActPlatId":"happy2021",
                "ActPlatCounterId":"view",
                "Stages":[
                    {
                        "DisplayName":"参与人数到达5000",
                        "Id":221,
                        "RequireCount":5000,
                        "AwardId":10
                    },
                    {
                        "DisplayName":"参与人数到达10000",
                        "Id":222,
                        "RequireCount":10000,
                        "AwardId":11
                    },
                    {
                        "DisplayName":"参与人数到达15000",
                        "Id":223,
                        "RequireCount":15000,
                        "AwardId":12
                    }
                ]
            }
        ],
        "LevelTask":{
            "DisplayName":"战令系统",
            "ActPlatId":"happy2021",
            "ActPlatCounterId":"view",
            "Stages":[
                {
                    "DisplayName":"Lv.1",
                    "Id":301,
                    "RequireCount":500,
                    "AwardId":1
                },
                {
                    "DisplayName":"Lv.2",
                    "Id":302,
                    "RequireCount":1500,
                    "AwardId":2
                },
                {
                    "DisplayName":"Lv.3",
                    "Id":303,
                    "RequireCount":3000,
                    "AwardId":3
                },
                {
                    "DisplayName":"Lv.4",
                    "Id":304,
                    "RequireCount":5000,
                    "AwardId":4
                },
                {
                    "DisplayName":"Lv.5",
                    "Id":305,
                    "RequireCount":7500,
                    "AwardId":5
                },
                {
                    "DisplayName":"Lv.5",
                    "Id":30501,
                    "RequireCount":7500,
                    "AwardId":5,
                    "VipHidden":true,
                    "VipSuitID":33385
                },
                {
                    "DisplayName":"Lv.6",
                    "Id":306,
                    "RequireCount":7500,
                    "AwardId":5
                },
                {
                    "DisplayName":"Lv.7",
                    "Id":307,
                    "RequireCount":7500,
                    "AwardId":5
                },
                {
                    "DisplayName":"Lv.8",
                    "Id":308,
                    "RequireCount":7500,
                    "AwardId":20
                },
                {
                    "DisplayName":"Lv.9",
                    "Id":309,
                    "RequireCount":7500,
                    "AwardId":20
                },
                {
                    "DisplayName":"Lv.10",
                    "Id":310,
                    "RequireCount":7500,
                    "AwardId":21
                },
                {
                    "DisplayName":"Lv.11",
                    "Id":311,
                    "RequireCount":7500,
                    "AwardId":22
                },
                {
                    "DisplayName":"Lv.12",
                    "Id":312,
                    "RequireCount":7500,
                    "AwardId":23
                },
                {
                    "DisplayName":"Lv.13",
                    "Id":313,
                    "RequireCount":7500,
                    "AwardId":24
                },
                {
                    "DisplayName":"Lv.14",
                    "Id":314,
                    "RequireCount":3000,
                    "AwardId":25
                },
                {
                    "DisplayName":"Lv.15",
                    "Id":315,
                    "RequireCount":7500,
                    "AwardId":26
                },
                {
                    "DisplayName":"Lv.16",
                    "Id":316,
                    "RequireCount":5000,
                    "AwardId":27
                },
                {
                    "DisplayName":"Lv.17",
                    "Id":317,
                    "RequireCount":7500,
                    "AwardId":28
                },
                {
                    "DisplayName":"Lv.18",
                    "Id":318,
                    "RequireCount":7500,
                    "AwardId":29
                },
                {
                    "DisplayName":"Lv.19",
                    "Id":319,
                    "RequireCount":3000,
                    "AwardId":30
                },
                {
                    "DisplayName":"Lv.20",
                    "Id":320,
                    "RequireCount":3000,
                    "AwardId":31
                }
            ]
        }
    }
}`
