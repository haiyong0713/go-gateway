package dao

import (
	"context"
	"fmt"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/admin/component"
	"sync"
)

const (
	DSTypeOperVideo = "OPER_VIDEO"
	DSTypeOperUp    = "OPER_UP"
	DSTypeOperPic   = "OPER_PIC"
)

const (
	sql4GetSourceTypeByMainId = `
SELECT id,
       source_type
FROM act_vote_data_sources_group
WHERE main_id=?
`
	sql4GetNewVoteUserRecord = `
SELECT source_group_id,
       source_item_id,
       mid,
       had_risk,
       votes,
       ctime,
       is_undo
FROM act_vote_user_action_%02d
WHERE main_id= ?
`
)

type UserVoteRecord struct {
	SourceType    string
	SourceGroupId int64
	SourceItemId  int64
	Mid           int64
	Vote          int64
	HadRisk       int64
	Ctime         xtime.Time
	State         string
}

func (d *Dao) getSourceTypeMap(ctx context.Context, activityId int64) (sourceTypeMap map[int64]string, err error) {
	sourceTypeMap = make(map[int64]string, 0)
	rows, err := component.RewardsDB.Query(ctx, sql4GetSourceTypeByMainId, activityId)
	if err != nil {
		return
	}
	err = rows.Err()
	if err != nil {
		return
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		sourceGroupId := int64(0)
		sourceType := ""
		err = rows.Scan(&sourceGroupId, &sourceType)
		if err != nil {
			return
		}
		sourceTypeMap[sourceGroupId] = readableVoteSourceType(sourceType)
	}
	return
}

func (d *Dao) getUserVoteHistoryByIdx(ctx context.Context, activityId, tableIdx int64, sourceTypeMap map[int64]string) (userRecords []*UserVoteRecord, err error) {
	userRecords = make([]*UserVoteRecord, 0)
	rows, err := component.RewardsDB.Query(ctx, fmt.Sprintf(sql4GetNewVoteUserRecord, tableIdx), activityId)
	if err != nil {
		return
	}
	defer func() { _ = rows.Close() }()
	if err = rows.Err(); err != nil {
		return
	}
	for rows.Next() {
		tmp := &UserVoteRecord{}
		var state int64
		err = rows.Scan(&tmp.SourceGroupId, &tmp.SourceItemId, &tmp.Mid, &tmp.HadRisk, &tmp.Vote, &tmp.Ctime, &state)
		if err != nil {
			return
		}
		tmp.State = readableVoteState(state)
		tmp.SourceType = sourceTypeMap[tmp.SourceGroupId]
		userRecords = append(userRecords, tmp)
	}
	err = rows.Err()
	return
}

func (d *Dao) GetNewVoteUserHistory(ctx context.Context, activityId int64) (res []*UserVoteRecord, err error) {
	sourceTypeMap, err := d.getSourceTypeMap(ctx, activityId)
	if len(sourceTypeMap) == 0 {
		return
	}
	res = make([]*UserVoteRecord, 0)
	tmpRes := make([][]*UserVoteRecord, 0)
	tmpResMu := sync.Mutex{}
	var eg errgroup.Group
	for i := 0; i < 100; i++ {
		tmpI := int64(i)
		eg.Go(func(ctx context.Context) (err error) {
			innerRes, err := d.getUserVoteHistoryByIdx(ctx, activityId, tmpI, sourceTypeMap)
			if err != nil {
				return
			}
			tmpResMu.Lock()
			tmpRes = append(tmpRes, innerRes)
			tmpResMu.Unlock()
			return
		})
	}
	err = eg.Wait()
	if err != nil {
		return
	}
	for _, rs := range tmpRes {
		res = append(res, rs...)
	}
	return
}

func readableVoteSourceType(s string) string {
	switch s {
	case DSTypeOperVideo:
		return "运营数据源(视频)"
	case DSTypeOperUp:
		return "运营数据源(UP)"
	case DSTypeOperPic:
		return "运营数据源(图片)"
	default:
		return s
	}
}

func readableVoteState(state int64) string {
	switch state {
	case 0:
		return "正常"
	case 1:
		return "已回滚"
	default:
		return "未知"

	}
}
