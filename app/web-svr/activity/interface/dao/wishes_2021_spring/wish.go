package wishes_2021_spring

import (
	"context"
	xsql "database/sql"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/ecode"

	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/component"
	model "go-gateway/app/web-svr/activity/interface/model/wishes_2021_spring"
	"go-gateway/app/web-svr/activity/interface/tool"
)

const (
	cacheKey4UserCommitContentTimes   = "activity:common:user_commit_content_times:%v:%v"
	cacheKey4UserCommitManScriptTimes = "activity:common:user_commit_manuscript_times:%v:%v"
	cacheKey4UserCommitAggregation    = "activity:common:user_commit_aggregation:%v:%v"

	limitKey4UserCommitContentTimes2DBQuery    = "activity_user_commit_content_times_query"
	limitKey4UserCommitManuScriptTimes2DBQuery = "activity_user_commit_manuscript_times_query"
	limitKey4UserCommitAggregation2DBQuery     = "activity_user_commit_aggregation_query"

	sql4FetchUserCommitContentTimes = `
SELECT COUNT(1)
FROM user_commit_log_tmp_%v
WHERE mid = ?
    AND activity_id = ?
`
	sql4FetchUserCommitContentID = `
SELECT id
FROM user_commit_log_tmp_%v
WHERE mid = ?
    AND activity_id = ?
LIMIT 1
`
	sql4UpdateUserCommitContent = `
UPDATE user_commit_content_tmp_%v
SET content = ?
WHERE commit_id = ?
`
	sql4FetchUserManuScriptList = `
SELECT bvid, content
FROM user_commit_manuscript_tmp_%v
WHERE mid = ?
    AND activity_id = ?
`

	sql4CountUserManuScriptList = `
SELECT count(id)
FROM user_commit_manuscript_tmp_%v
`
	sql4FetchUserManuScriptCount = `
SELECT COUNT(1)
FROM user_commit_manuscript_tmp_%v
WHERE mid = ?
    AND activity_id = ?
`
	sql4FetchUserCommitContent = `
SELECT content
FROM user_commit_content_tmp_%v
WHERE commit_id = (
	SELECT id
	FROM user_commit_log_tmp_%v
	WHERE mid = ?
		AND activity_id = ?
    LIMIT 1
)
`
	sql4InsertUserCommitLog = `
INSERT INTO user_commit_log_tmp_%v (mid, activity_id)
VALUES (?, ?);
`
	sql4InsertUserCommitContent = `
INSERT INTO user_commit_content_tmp_%v (commit_id, content)
VALUES (?, ?);
`
	sql4InsertUserCommitManuScript = `
INSERT INTO user_commit_manuscript_tmp_%v (mid, activity_id, content, bvid)
VALUES (?, ?, ?, ?);
`
	sql4UserCommitContentListInLive = `
SELECT id, mid, bvid, content
	, UNIX_TIMESTAMP(ctime)
FROM user_commit_manuscript_tmp_%v
WHERE id > ?
	AND activity_id = ?
ORDER BY ID %v
LIMIT %v
`
	sql4UserCommitContentCountInLive = `
SELECT COUNT(*)
FROM user_commit_manuscript_tmp_%v
WHERE activity_id = ?
`

	sql4UserCommitAuditMaterial = `
INSERT INTO user_commit_manuscript_audit_material_%v (mid , avid , activity , material_user , material_text , material_images)
VALUES (?, ?, ?, ?, ? ,?);
`
	sql4UserCommitAuditMaterialCount = `
SELECT COUNT(DISTINCT(avid))
FROM user_commit_manuscript_audit_material_%v
WHERE mid = ?
`
)

func FetchUserCommitContentCountInLive(ctx context.Context, req *model.UserCommitListRequestInLive) (
	count int64, err error) {
	suffix := fmt.Sprintf("%02d", req.ActivityID%100)
	sqlStr := fmt.Sprintf(sql4UserCommitContentCountInLive, suffix)
	err = component.GlobalBnjDB.QueryRow(ctx, sqlStr, req.ActivityID).Scan(&count)

	return
}

func FetchUserCommitContentListInLive(ctx context.Context, req *model.UserCommitListRequestInLive) (
	list []map[string]interface{}, err error) {
	list = make([]map[string]interface{}, 0)
	suffix := fmt.Sprintf("%02d", req.ActivityID%100)
	sqlStr := fmt.Sprintf(sql4UserCommitContentListInLive, suffix, req.Order, req.Ps)

	var rows *sql.Rows
	rows, err = component.GlobalBnjDB.Query(ctx, sqlStr, req.LastID, req.ActivityID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id, mid, ctime int64
			bvid, content  string
		)

		if tmpErr := rows.Scan(&id, &mid, &bvid, &content, &ctime); tmpErr == nil {
			tmpM := make(map[string]interface{}, 0)
			{
				tmpM["id"] = id
				tmpM["mid"] = mid
				tmpM["bvid"] = bvid
				tmpM["postinfo"] = content
				tmpM["ctime"] = ctime
				tmpM["user_info"] = new(model.UserInfo)
			}

			list = append(list, tmpM)
		}
	}

	err = rows.Err()

	return
}

func InsertUserCommitManuScript(ctx context.Context, req *api.CommonActivityUserCommitReq, removeCache bool) (
	lastInsertID int64, err error) {
	suffix := fmt.Sprintf("%02d", req.ActivityID%100)
	sqlStr := fmt.Sprintf(sql4InsertUserCommitManuScript, suffix)
	var result xsql.Result
	result, err = component.GlobalBnjDB.Exec(ctx, sqlStr, req.MID, req.ActivityID, req.Content, req.BvID)
	if err == nil {
		lastInsertID, err = result.LastInsertId()
		if removeCache {
			_ = RemoveUserCommitAggregationCache(ctx, req.MID, req.ActivityID)
		}

		for i := int64(0); i < 10; i++ {
			if tmpErr := DeleteUserCommitManuScriptTimesCache(ctx, req.MID, req.ActivityID); tmpErr == nil {
				break
			}

			time.Sleep(time.Duration(i * 10 * int64(time.Millisecond)))
		}
	}

	return
}

func UpdateUserCommitContent(ctx context.Context, commitID int64, req *api.CommonActivityUserCommitReq, removeCache bool) (err error) {
	suffix := fmt.Sprintf("%02d", req.ActivityID%100)
	sqlStr := fmt.Sprintf(sql4UpdateUserCommitContent, suffix)
	var result xsql.Result
	result, err = component.GlobalBnjDB.Exec(ctx, sqlStr, req.Content, commitID)
	if err == nil && removeCache {
		var affected int64
		affected, err = result.RowsAffected()
		if err == nil && affected > 0 {
			_ = RemoveUserCommitAggregationCache(ctx, req.MID, req.ActivityID)
		}

	}

	return
}

func InsertUserCommitLogAndContent(ctx context.Context, req *api.CommonActivityUserCommitReq, removeCache bool) (err error) {
	var tx *sql.Tx
	tx, err = component.GlobalBnjDB.Begin(ctx)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var (
		result       xsql.Result
		lastInsertID int64
	)
	suffix := fmt.Sprintf("%02d", req.ActivityID%100)
	sql4Log := fmt.Sprintf(sql4InsertUserCommitLog, suffix)
	result, err = tx.Exec(sql4Log, req.MID, req.ActivityID)
	if err != nil {
		return
	}

	lastInsertID, err = result.LastInsertId()
	if err != nil {
		return
	}

	sql4Content := fmt.Sprintf(sql4InsertUserCommitContent, suffix)
	_, err = tx.Exec(sql4Content, lastInsertID, req.Content)

	err = tx.Commit()
	if err == nil && removeCache {
		_ = RemoveUserCommitAggregationCache(ctx, req.MID, req.ActivityID)
	}

	return
}

func genCacheKey4UserCommitAggregation(mid, activityID int64) (key string) {
	return fmt.Sprintf(cacheKey4UserCommitAggregation, mid, activityID)
}

func FetchUserCommitContent(ctx context.Context, mid, activityID int64) (resp *model.UserCommit4Aggregation, err error) {
	resp = model.NewUserCommit4Aggregation()
	cacheKey := genCacheKey4UserCommitAggregation(mid, activityID)
	err = component.S10GlobalMC.Get(ctx, cacheKey).Scan(&resp)
	if err == nil {
		return
	}

	if err == memcache.ErrNotFound {
		if tool.IsLimiterAllowedByUniqBizKey(limitKey4UserCommitAggregation2DBQuery, limitKey4UserCommitAggregation2DBQuery) {
			var content string
			content, err = FetchUserCommitContentFromDB(ctx, mid, activityID)
			if err == nil {
				manuScriptLit := make([]map[string]interface{}, 0)
				manuScriptLit, err = FetchUserCommitManuScriptListFromDB(ctx, mid, activityID)
				if err == nil {
					resp.Content = content
					resp.ExtraList = manuScriptLit

					item := new(memcache.Item)
					{
						item.Key = cacheKey
						item.Object = resp
						item.Expiration = int32(tool.CalculateExpiredSeconds(1))
						item.Flags = memcache.FlagJSON
					}
					err = component.S10GlobalMC.Set(ctx, item)
				}
			}
		} else {
			err = ecode.LimitExceed
		}
	}

	return
}

func FetchUserCommitManuScriptListFromDB(ctx context.Context, mid, activityID int64) (list []map[string]interface{}, err error) {
	list = make([]map[string]interface{}, 0)
	sqlStr := fmt.Sprintf(sql4FetchUserManuScriptList, fmt.Sprintf("%02d", activityID%100))
	var rows *sql.Rows
	rows, err = component.GlobalBnjDB.Query(ctx, sqlStr, mid, activityID)
	if err == sql.ErrNoRows {
		err = nil

		return
	}

	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var bvid, content string
		if tmpErr := rows.Scan(&bvid, &content); tmpErr == nil {
			tmp := make(map[string]interface{}, 0)
			{
				tmp["bvid"] = bvid
				tmp["content"] = content
			}
			list = append(list, tmp)
		}
	}

	err = rows.Err()

	return
}

func RemoveUserCommitAggregationCache(ctx context.Context, mid, activityID int64) (err error) {
	for i := 0; i < 10; i++ {
		err = component.S10GlobalMC.Delete(ctx, genCacheKey4UserCommitAggregation(mid, activityID))
		if err == nil {
			break
		}
	}

	return
}

func ResetUserCommitContentInCache(ctx context.Context, mid, activityID int64, resp *model.UserCommit4Aggregation) (err error) {
	item := new(memcache.Item)
	{
		item.Key = genCacheKey4UserCommitAggregation(mid, activityID)
		item.Object = resp
		item.Expiration = int32(tool.CalculateExpiredSeconds(1))
		item.Flags = memcache.FlagJSON
	}
	err = component.S10GlobalMC.Set(ctx, item)

	return
}

func FetchUserCommitContentFromDB(ctx context.Context, mid, activityID int64) (content string, err error) {
	suffix := fmt.Sprintf("%02d", activityID%100)
	sqlStr := fmt.Sprintf(sql4FetchUserCommitContent, suffix, suffix)
	err = component.GlobalBnjDB.QueryRow(ctx, sqlStr, mid, activityID).Scan(&content)
	if err == sql.ErrNoRows {
		err = nil
	}

	return
}

func genCacheKey4UserCommitContentTimes(mid, activityID int64) (key string) {
	return fmt.Sprintf(cacheKey4UserCommitContentTimes, mid, activityID)
}

func FetchUserCommitContentTimes(ctx context.Context, mid, activityID int64) (times int64, err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	cacheKey := genCacheKey4UserCommitContentTimes(mid, activityID)
	times, err = redis.Int64(conn.Do("GET", cacheKey))
	if err == nil {
		return
	}

	if err == redis.ErrNil {
		if tool.IsLimiterAllowedByUniqBizKey(limitKey4UserCommitContentTimes2DBQuery, limitKey4UserCommitContentTimes2DBQuery) {
			times, err = FetchUserCommitContentTimesFromDB(ctx, mid, activityID)
			if err == nil {
				_, _ = conn.Do("SETEX", cacheKey, tool.CalculateExpiredSeconds(14), times)
			}
		} else {
			err = ecode.LimitExceed
		}
	}

	return
}

func FetchUserCommitContentIDFromDB(ctx context.Context, mid, activityID int64) (commitID int64, err error) {
	sqlStr := fmt.Sprintf(sql4FetchUserCommitContentID, fmt.Sprintf("%02d", activityID%100))
	err = component.GlobalBnjDB.QueryRow(ctx, sqlStr, mid, activityID).Scan(&commitID)
	if err == sql.ErrNoRows {
		err = nil
	}

	return
}

func FetchUserCommitContentTimesFromDB(ctx context.Context, mid, activityID int64) (times int64, err error) {
	sqlStr := fmt.Sprintf(sql4FetchUserCommitContentTimes, fmt.Sprintf("%02d", activityID%100))
	row := component.GlobalBnjDB.QueryRow(ctx, sqlStr, mid, activityID)
	err = row.Scan(&times)
	if err == sql.ErrNoRows {
		err = nil
	}

	return
}

func FetchUserCommitManuScriptTimes(ctx context.Context, mid, activityID int64) (times int64, err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	cacheKey := genCacheKey4UserCommitManuScriptTimes(mid, activityID)
	times, err = redis.Int64(conn.Do("GET", cacheKey))
	if err == nil {
		return
	}

	if err == redis.ErrNil {
		if tool.IsLimiterAllowedByUniqBizKey(limitKey4UserCommitManuScriptTimes2DBQuery, limitKey4UserCommitManuScriptTimes2DBQuery) {
			times, err = FetchUserCommitManuScriptTimesFromDB(ctx, mid, activityID)
			if err == nil {
				_, _ = conn.Do("SETEX", cacheKey, tool.CalculateExpiredSeconds(14), times)
			}
		} else {
			err = ecode.LimitExceed
		}
	}

	return
}

func DeleteUserCommitManuScriptTimesCache(ctx context.Context, mid, activityID int64) (err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.Do("DEL", genCacheKey4UserCommitManuScriptTimes(mid, activityID))

	return
}

func genCacheKey4UserCommitManuScriptTimes(mid, activityID int64) (key string) {
	return fmt.Sprintf(cacheKey4UserCommitManScriptTimes, mid, activityID)
}

func FetchUserCommitManuScriptTimesFromDB(ctx context.Context, mid, activityID int64) (times int64, err error) {
	sqlStr := fmt.Sprintf(sql4FetchUserManuScriptCount, fmt.Sprintf("%02d", activityID%100))
	row := component.GlobalBnjDB.QueryRow(ctx, sqlStr, mid, activityID)
	err = row.Scan(&times)
	if err == sql.ErrNoRows {
		err = nil
	}

	return
}

func CountAuditMaterialRecord(ctx context.Context, mid, actId int64) (count int64, err error) {
	sqlStr := fmt.Sprintf(sql4UserCommitAuditMaterialCount, fmt.Sprintf("%02d", actId%100))

	err = component.GlobalBnjDB.QueryRow(ctx, sqlStr, mid).Scan(&count)
	return
}

func InsertAuditMaterialRecord(ctx context.Context, mid, actId int64, record *model.AuditInfo) (lastInsertID int64, err error) {

	sqlStr := fmt.Sprintf(sql4UserCommitAuditMaterial, fmt.Sprintf("%02d", actId%100))

	var result xsql.Result
	materialUser, _ := json.Marshal(record.Materials.User)
	materialText, _ := json.Marshal(record.Materials.Text)
	materialImg, _ := json.Marshal(record.Materials.Images)

	result, err = component.GlobalBnjDB.Exec(ctx, sqlStr, mid, record.Avid, record.Activity,
		string(materialUser), string(materialText), string(materialImg))

	if err == nil {
		lastInsertID, err = result.LastInsertId()
	}
	return
}

//nolint:bilirailguncheck
func ManuScriptAuditPub(ctx context.Context, record *model.AuditInfo) (err error) {
	if record == nil {
		return
	}
	dbKey := fmt.Sprintf("audit_material_%v", record.Avid)
	err = component.ActAuditMaterialProducer.Send(ctx, dbKey, record)
	recordStr, _ := json.Marshal(record)
	log.Infoc(ctx, "ManuScriptAuditPub databus key:%v , message:%s ,err:%+v", dbKey, recordStr, err)
	return
}

func CountUserManuScriptList(ctx context.Context, actId int64) (count int64, err error) {
	sqlStr := fmt.Sprintf(sql4CountUserManuScriptList, fmt.Sprintf("%02d", actId%100))
	err = component.GlobalBnjDB.QueryRow(ctx, sqlStr).Scan(&count)
	return
}
