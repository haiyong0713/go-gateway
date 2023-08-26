package knowledge

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	model "go-gateway/app/web-svr/activity/interface/model/knowledge"
)

const sql4RawConfigs = `
SELECT id,  config_details
FROM act_knowledge_config
WHERE is_deleted = 0
`

func (d *Dao) RawFetchKnowledgeConfigs(ctx context.Context) (res map[int64]*model.KnowConfig, err error) {
	res = make(map[int64]*model.KnowConfig)
	rows, err := d.db.Query(ctx, sql4RawConfigs)
	if err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
		if err == nil {
			err = rows.Err()
		}
	}()
	for rows.Next() {
		var strDetail string
		kc := &model.KnowConfig{}
		err = rows.Scan(&kc.ID, &strDetail)
		if err != nil {
			log.Errorc(ctx, "RawFetchKnowledgeConfigs rows.scan() error(%+v)", err)
			return
		}
		if strDetail == "" {
			continue
		}
		tmpDetail := &model.KnowConfigDetail{}
		err = json.Unmarshal([]byte(strDetail), tmpDetail)
		if err != nil {
			log.Errorc(ctx, "RawFetchKnowledgeConfigs json.Unmarshal() error(%+v)", err)
			return
		}
		kc.ConfigDetails = tmpDetail
		res[kc.ID] = kc
	}
	return
}

const sql4RawUserKnowledgeTaskByMid = `
SELECT id,
       mid,
       had_arc,
       coin,
       favorite,
       share,
       year_2020_share,
       year_2021_share,
       see_2020_share,
       see_2021_share,
       super_2020_share,
       super_2021_share,
       gold_2020_share,
       gold_2021_share,
       dark_2020_share,
       dark_2021_share
FROM  %s
WHERE mid = ? limit 1
`

func (d *Dao) RawUserKnowledgeTask(ctx context.Context, mid int64, table string) (resMap map[string]int64, err error) {
	resMap = make(map[string]int64)
	userTask := &model.UserKnowTask{}
	row := d.db.QueryRow(ctx, fmt.Sprintf(sql4RawUserKnowledgeTaskByMid, table), mid)
	err = row.Scan(&userTask.ID, &userTask.Mid, &userTask.HadArc, &userTask.Coin, &userTask.Favorite, &userTask.Share,
		&userTask.Year2020Share, &userTask.Year2021Share, &userTask.See2020Share, &userTask.See2021Share,
		&userTask.Super2020Share, &userTask.Super2021Share, &userTask.Gold2020Share, &userTask.Gold2021Share,
		&userTask.Dark2020Share, &userTask.Dark2021Share)
	if err == sql.ErrNoRows {
		err = nil
		resMap["id"] = -1
		return
	}
	if err != nil {
		return
	}
	resMap = genConfigMap(userTask)
	return
}

func genConfigMap(userTask *model.UserKnowTask) (userTaskMap map[string]int64) {
	userTaskMap = make(map[string]int64)
	userTaskMap["id"] = userTask.ID
	userTaskMap["mid"] = userTask.Mid
	userTaskMap["had_arc"] = userTask.HadArc
	userTaskMap["coin"] = userTask.Coin
	userTaskMap["favorite"] = userTask.Favorite
	userTaskMap["share"] = userTask.Share
	userTaskMap["year_2020_share"] = userTask.Year2020Share
	userTaskMap["year_2021_share"] = userTask.Year2021Share
	userTaskMap["see_2020_share"] = userTask.See2020Share
	userTaskMap["see_2021_share"] = userTask.See2021Share
	userTaskMap["super_2020_share"] = userTask.Super2020Share
	userTaskMap["super_2021_share"] = userTask.Super2021Share
	userTaskMap["gold_2020_share"] = userTask.Gold2020Share
	userTaskMap["gold_2021_share"] = userTask.Gold2021Share
	userTaskMap["dark_2020_share"] = userTask.Dark2020Share
	userTaskMap["dark_2021_share"] = userTask.Dark2021Share
	return
}

const sql4UpdateUserKnowledgeTask = `
INSERT INTO act_knowledge_task (mid, %s)
VALUES (?,1)
ON DUPLICATE KEY UPDATE mid = ?, %s = 1
`

func (d *Dao) UpdateInsertUserKnowledgeTask(ctx context.Context, fieldName string, mid int64) (err error) {
	err = retry.WithAttempts(ctx, "updateActivityRankRefreshTime", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		_, err = d.db.Exec(ctx, fmt.Sprintf(sql4UpdateUserKnowledgeTask, fieldName, fieldName), mid, mid)
		return err
	})
	return
}

const sql4UpdateKnowledgeConfig = `
UPDATE act_knowledge_config
SET config_details=? WHERE id=?
`

func (d *Dao) UpdateKnowledgeConfig(ctx context.Context, jsonConfig string, id int64) (err error) {
	_, err = d.db.Exec(ctx, sql4UpdateKnowledgeConfig, jsonConfig, id)
	return
}
