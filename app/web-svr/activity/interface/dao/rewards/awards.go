package rewards

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	"strconv"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
)

const (
	sql4GetAwardsConfigs = `
SELECT id, 
       activity_id, 
       display_name, 
       award_type, 
       config_content, 
       should_send_notify, 
       notify_sender_id, 
       notify_code, 
       notify_message, 
       notify_jump_url, 
       notify_jump_url2, 
       extra_info,
       icon_url
FROM   rewards_award_config 
WHERE  is_deleted = 0
`
	sql4GetAwardsConfigsByActivityId = `
SELECT id, 
       activity_id, 
       display_name, 
       award_type, 
       config_content, 
       should_send_notify, 
       notify_sender_id, 
       notify_code, 
       notify_message, 
       notify_jump_url,
       notify_jump_url2,
       extra_info,
       icon_url 
FROM   rewards_award_config 
WHERE  is_deleted = 0 
AND activity_id = %d
`
	sql4AddAwardConfig = `
INSERT INTO rewards_award_config 
            (activity_id, 
             display_name, 
             award_type, 
             config_content, 
             should_send_notify, 
             notify_sender_id, 
             notify_code, 
             notify_message, 
             notify_jump_url,
             notify_jump_url2,
             extra_info,
             icon_url) 
VALUES      ( ?, 
              ?, 
              ?, 
              ?, 
              ?, 
              ?,
              ?,
              ?,
              ?, 
              ?,
              ?,
              ? ) `

	sql4UpdateAwardConfig = `
UPDATE rewards_award_config 
SET    display_name = ?, 
       award_type = ?, 
       config_content = ?, 
       should_send_notify = ?,
       notify_sender_id = ?, 
       notify_code = ?, 
       notify_message = ?, 
       notify_jump_url = ?,
       notify_jump_url2 = ?,
       extra_info = ?, 
       icon_url = ?
WHERE  id = ? 
       AND activity_id = ?  
       AND is_deleted = 0  
 `
)

// GetAwardSlice: 获取所有奖励配置信息
// 特殊逻辑: 当前通知支持配置在奖品和活动级别.
// 1.如果奖品级别的通知配置存在, 那么使用奖品级别的配置
// 2.如果奖品级别的通知配置不存在,那么尝试使用活动级别的配置
func (d *Dao) GetAwardSlice(ctx context.Context, activityId int64, keyword string) (res []*api.RewardsAwardInfo, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards d.GetAwardMap error: %v", err)
		}
	}()
	res = make([]*api.RewardsAwardInfo, 0)
	activityList, err := d.ListActivity(ctx, &api.RewardsListActivityReq{}, false)
	if err != nil {
		return
	}
	activityMap := make(map[int64] /*activityId*/ *api.RewardsActivityListInfo, len(activityList.List))
	for _, a := range activityList.List {
		ta := a
		activityMap[a.Id] = ta
	}
	var rows *sql.Rows
	sqlStr := sql4GetAwardsConfigs
	if activityId != 0 {
		sqlStr = fmt.Sprintf(sql4GetAwardsConfigsByActivityId, activityId)
	}

	if keyword != "" {
		cond := fmt.Sprintf(" AND ( display_name like '%%%v%%'", keyword)
		if kInt, err := strconv.Atoi(keyword); err == nil {
			cond = cond + fmt.Sprintf(" OR id like %v", kInt)
		}
		cond = cond + " ) "
		sqlStr = sqlStr + cond
	}
	sqlStr = sqlStr + " ORDER  BY id DESC "
	rows, err = d.db.Query(ctx, sqlStr)
	if err == sql.ErrNoRows {
		err = nil
		return
	}
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		extraInfo := ""
		var shouldSend int64
		c := &api.RewardsAwardInfo{ExtraInfo: make(map[string]string)}
		if err = rows.Scan(&c.Id, &c.ActivityId, &c.Name, &c.Type,
			&c.JsonStr, &shouldSend, &c.NotifySenderId, &c.NotifyCode, &c.NotifyMessage, &c.NotifyJumpUri1, &c.NotifyJumpUri2, &extraInfo, &c.IconUrl); err != nil {
			return
		}
		if extraInfo != "" {
			err = json.Unmarshal([]byte(extraInfo), &c.ExtraInfo)
			if err != nil {
				return
			}
		}
		c.ShouldSendNotify = shouldSend == 1
		a := activityMap[c.ActivityId]
		if a != nil {
			c.ActivityName = a.Name
		}
		//奖品没有配置通知, 尝试获取活动级别的通知
		{
			if a != nil {
				if c.NotifySenderId == 0 {
					c.NotifySenderId = a.NotifySenderId
				}
				if c.NotifyCode == "" {
					c.NotifyCode = a.NotifyCode
				}
				if c.NotifyMessage == "" {
					c.NotifyMessage = a.NotifyMessage
				}
				if c.NotifyJumpUri1 == "" {
					c.NotifyJumpUri1 = a.NotifyJumpUri1
				}
				if c.NotifyJumpUri2 == "" {
					c.NotifyJumpUri2 = a.NotifyJumpUri2
				}
			}

		}
		res = append(res, c)
	}
	err = rows.Err()
	return
}

// GetAwardMap: 跟GetAwardSlice逻辑相同, 只是返回结构不同
func (d *Dao) GetAwardMap(ctx context.Context, activityId int64) (res map[int64]*api.RewardsAwardInfo, err error) {
	cs, err := d.GetAwardSlice(ctx, activityId, "")
	if err != nil {
		return
	}
	res = make(map[int64]*api.RewardsAwardInfo, 0)
	for _, c := range cs {
		res[c.Id] = c
	}
	return
}

// AddAward: 向活动下添加奖品配置
func (d *Dao) AddAward(ctx context.Context, c *api.RewardsAddAwardReq) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards d.AddAward Exec error: %v", err)
		}
	}()
	//check activity id exists
	_, err = d.GetActivityDetail(ctx, &api.RewardsGetActivityDetailReq{ActivityId: c.ActivityId})
	if err != nil {
		return err
	}
	extraInfo := ""
	if c.ExtraInfo != nil {
		bs, _ := json.Marshal(c.ExtraInfo)
		extraInfo = string(bs)
	}
	_, err = d.db.Exec(ctx, sql4AddAwardConfig, c.ActivityId, c.Name, c.Type, c.JsonStr, c.ShouldSendNotify, c.NotifySenderId, &c.NotifyCode, c.NotifyMessage, c.NotifyJumpUri1, c.NotifyJumpUri2, extraInfo, c.IconUrl)

	return err
}

// DelAward: 从活动下删除奖品配置
func (d *Dao) DelAward(ctx context.Context, awardId int64) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards d.DelAward Exec error: %v", err)
		}
	}()

	res, err := d.db.Exec(ctx, sql4DelAwardConfigById, awardId)
	if err != nil {
		return err
	}
	af, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if af == 0 {
		err = ecode.ActivityIDNotExists
		return
	}

	return err
}

// UpdateAward: 更新活动下的奖品配置
func (d *Dao) UpdateAward(ctx context.Context, c *api.RewardsAwardInfo) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards d.UpdateAward  error: %v", err)
		}
	}()
	//check activity id exists
	activity, err := d.GetActivityDetail(ctx, &api.RewardsGetActivityDetailReq{ActivityId: c.ActivityId})
	if err != nil {
		return err
	}
	found := false
	for _, award := range activity.List {
		if award.Id == c.Id {
			found = true
			break
		}
	}
	if !found {
		err = ecode.ActivityIDNotExists
		return
	}
	extraInfo := ""
	if c.ExtraInfo != nil {
		bs, _ := json.Marshal(c.ExtraInfo)
		extraInfo = string(bs)
	}
	_, err = d.db.Exec(ctx, sql4UpdateAwardConfig, c.Name, c.Type, c.JsonStr, c.ShouldSendNotify, c.NotifySenderId, &c.NotifyCode, c.NotifyMessage, c.NotifyJumpUri1, c.NotifyJumpUri2, extraInfo, c.IconUrl, c.Id, c.ActivityId)

	return err
}

func userActCounterTimestampCacheKey(mid, timestamp int64) string {
	return fmt.Sprintf("act_counter_%v_%v", mid, timestamp)
}

// GetActCounterUnusedTimestampByMid: 获取此MID未被使用的timestamp(活动平台一个MID单秒内只会消费一条消息)
func (d *Dao) GetActCounterUnusedTimestampByMid(ctx context.Context, mid int64) (ts int64, err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() { _ = conn.Close() }()
	var ok bool
	ts = time.Now().Unix()
	for i := 0; i < 1000; i++ {
		key := userActCounterTimestampCacheKey(mid, ts)
		if ok, err = redis.Bool(conn.Do("SETNX", key, "1")); err != nil {
			if err == redis.ErrNil {
				err = nil
			} else {
				log.Error("conn.Do(SETNX(%s)) error(%v)", key, err)
				return
			}
		}
		if ok {
			//expire must > retry count
			_, _ = conn.Do("EXPIRE", key, 1001)
			return
		}
		log.Errorc(ctx, "GetActCounterUnusedTimestampByMid: %v already used will try incr 1", ts)
		//ts exists, incr by 1
		ts++
	}
	err = fmt.Errorf("can not find unused timestamp after 10 retry")
	return
}
