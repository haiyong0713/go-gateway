package rewards

import (
	"context"
	"fmt"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	"strconv"
)

const (
	sql4DelAwardConfigById = `
UPDATE rewards_award_config SET 
is_deleted = 1
WHERE
id = ?`

	sql4DelAwardConfigByActivityId = `
UPDATE rewards_award_config SET 
is_deleted = 1
WHERE
activity_id = ?`

	sql4DelActivity = `
UPDATE rewards_activity_config 
SET    is_deleted = 1 
WHERE  id = ? `

	sql4AddActivity = `
INSERT INTO rewards_activity_config 
            (name, 
             notify_sender_id, 
             notify_code, 
             notify_message, 
             notify_jump_url,
             notify_jump_url2) 
VALUES      (?, 
             ?, 
             ?, 
             ?, 
             ?, 
             ?) `

	sql4UpdateActivity = `
UPDATE rewards_activity_config 
SET    name = ?, 
       notify_sender_id = ?, 
       notify_code = ?, 
       notify_message = ?, 
       notify_jump_url = ?, 
       notify_jump_url2 = ? 
WHERE  id = ? and is_deleted = 0`

	sql4GetActivity = `
SELECT id, 
       name, 
       notify_sender_id, 
       notify_code, 
       notify_message, 
       notify_jump_url,
       notify_jump_url2 
FROM   rewards_activity_config 
WHERE  id =? and is_deleted = 0`

	sql4ListActivity = `
SELECT id, 
       name, 
       notify_sender_id, 
       notify_code, 
       notify_message, 
       notify_jump_url, 
       notify_jump_url2
FROM   rewards_activity_config
WHERE is_deleted = 0
`

	sql4CountActivity = `
SELECT COUNT(*)
FROM  rewards_activity_config where is_deleted=0`
)

// AddActivity: 添加活动
func (d *Dao) AddActivity(ctx context.Context, c *api.RewardsAddActivityReq) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards d.AddActivity error: %v", err)
		}
	}()
	_, err = d.db.Exec(ctx, sql4AddActivity, c.Name, c.NotifySenderId, c.NotifyCode, c.NotifyMessage, c.NotifyJumpUri1, c.NotifyJumpUri2)

	return err
}

// GetActivityDetail: 获取指定活动的信息(包括奖品)
func (d *Dao) GetActivityDetail(ctx context.Context, c *api.RewardsGetActivityDetailReq) (res *api.RewardsGetActivityDetailReply, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards d.GetActivityDetail error: %v", err)
		}
	}()
	row := d.db.QueryRow(ctx, sql4GetActivity, c.ActivityId)
	res = &api.RewardsGetActivityDetailReply{}
	if err = row.Scan(&res.Id, &res.Name, &res.NotifySenderId, &res.NotifyCode, &res.NotifyMessage, &res.NotifyJumpUri1, &res.NotifyJumpUri1); err != nil {
		if err == sql.ErrNoRows {
			err = ecode.ActivityIDNotExists
		}
		return
	}

	award, err := d.GetAwardSlice(ctx, c.ActivityId, "")
	if err != nil {
		return
	}
	res.List = award

	return
}

// ListActivity: 查看活动列表
func (d *Dao) ListActivity(ctx context.Context, req *api.RewardsListActivityReq, splitPage bool) (res *api.RewardsListActivityReply, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards d.ListActivity error: %v", err)
		}
	}()

	var (
		start = (req.PageNumber - 1) * req.PageSize
	)
	res = &api.RewardsListActivityReply{
		List: make([]*api.RewardsActivityListInfo, 0),
		Page: &api.RewardsListActivityPage{
			Num:   req.PageNumber,
			Ps:    req.PageSize,
			Total: 0,
		},
	}
	err = d.db.QueryRow(ctx, sql4CountActivity).Scan(&res.Page.Total)
	if err != nil {
		return
	}
	sqlStr := sql4ListActivity
	if req.Keyword != "" {
		cond := fmt.Sprintf(" AND ( name like '%%%v%%'", req.Keyword)
		if kInt, err := strconv.Atoi(req.Keyword); err == nil {
			cond = cond + fmt.Sprintf(" OR id like %v", kInt)
		}
		cond = cond + " ) "
		sqlStr = sqlStr + cond
	}
	var rows *sql.Rows
	sqlStr = sqlStr + " ORDER BY id DESC"
	if splitPage {
		sqlStr = fmt.Sprintf("%s LIMIT %d, %d", sqlStr, start, req.PageSize)
	}

	rows, err = d.db.Query(ctx, sqlStr)

	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		a := &api.RewardsActivityListInfo{}
		if err = rows.Scan(&a.Id, &a.Name, &a.NotifySenderId, &a.NotifyCode, &a.NotifyMessage, &a.NotifyJumpUri1, &a.NotifyJumpUri2); err != nil {
			return
		}
		res.List = append(res.List, a)
	}
	err = rows.Err()
	return
}

// DelActivity: 删除活动配置, 活动下关联的所有奖品也会被删除
func (d *Dao) DelActivity(ctx context.Context, activityId int64) (err error) {
	_, err = d.GetActivityDetail(ctx, &api.RewardsGetActivityDetailReq{ActivityId: activityId})
	if err != nil {
		return err
	}
	tx, err := d.db.Begin(ctx)
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards d.DelActivity error: %v", err)
			if tx != nil {
				log.Errorc(ctx, "rewards d.DelActivity rollback error: %v", tx.Rollback())
			}
		}
	}()
	if err != nil {
		return
	}
	_, err = tx.Exec(sql4DelActivity, activityId)
	if err != nil {
		return
	}
	//将活动下关联的奖品全部删除
	_, err = tx.Exec(sql4DelAwardConfigByActivityId, activityId)
	if err != nil {
		return
	}
	err = tx.Commit()
	return err
}

// UpdateActivity: 更新活动配置
func (d *Dao) UpdateActivity(ctx context.Context, c *api.RewardsUpdateActivityReq) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards d.UpdateActivity error: %v", err)
		}
	}()
	_, err = d.GetActivityDetail(ctx, &api.RewardsGetActivityDetailReq{ActivityId: c.Id})
	if err != nil {
		return err
	}
	_, err = d.db.Exec(ctx, sql4UpdateActivity, c.Name, c.NotifySenderId, c.NotifyCode, c.NotifyMessage, c.NotifyJumpUri1, c.NotifyJumpUri2, c.Id)

	return err
}
