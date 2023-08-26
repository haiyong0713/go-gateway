package rank

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/job/model/rank"

	"github.com/pkg/errors"
)

const (
	rankName = "act_rank"
)

const (
	insertRankSQL           = "INSERT INTO %s (sid,mid,nickname,rank,score,state,batch,remark) VALUES %s"
	getRankByBatchSQL       = "SELECT sid,mid,nickname,rank,score,state,batch,remark FROM %s WHERE sid = ? and batch = ? and state = 0"
	getRankByBatchLimitSQL  = "SELECT sid,mid,nickname,rank,score,state,batch,remark FROM %s WHERE sid = ? and batch = ? and state = 0 limit ?,?"
	getMemberRankTimesSQL   = "SELECT mid,count(mid) FROM %s WHERE sid = ? and batch >= ? and batch <= ? and mid IN (%s) and state = 0 group by mid"
	getMemberHighestRankSQL = "SELECT mid,min(rank) FROM %s WHERE sid = ? and batch >= ? and batch <= ? and mid IN (%s) and state = 0 group by mid"
)

// BatchAddRank batch add rank
func (d *dao) BatchAddRank(c context.Context, rank []*rank.DB) (err error) {
	var (
		rows    []interface{}
		rowsTmp []string
		bs      []byte
	)
	for _, r := range rank {
		rowsTmp = append(rowsTmp, "(?,?,?,?,?,?,?,?)")

		if bs, err = json.Marshal(r.RemarkOrigin); err != nil {
			log.Errorc(c, "json.Marshal() error(%v)", err)
			return
		}
		r.Remark = string(bs)
		rows = append(rows, r.SID, r.Mid, r.NickName, r.Rank, r.Score, r.State, r.Batch, r.Remark)
	}
	sql := fmt.Sprintf(insertRankSQL, rankName, strings.Join(rowsTmp, ","))
	if _, err = d.db.Exec(c, sql, rows...); err != nil {
		err = errors.Wrap(err, "BatchAddRank: d.db.Exec")
	}
	return
}

// GetRankListByBatch 获取某个批次的排行结果
func (d *dao) GetRankListByBatch(c context.Context, sid, batch int64) (rs []*rank.DB, err error) {
	rs = []*rank.DB{}
	rows, err := d.db.Query(c, fmt.Sprintf(getRankByBatchSQL, rankName), sid, batch)
	if err != nil {
		err = errors.Wrap(err, "GetRankListByBatch:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &rank.DB{}
		err = rows.Scan(&r.SID, &r.Mid, &r.NickName, &r.Rank, &r.Score, &r.State, &r.Batch, &r.Remark)
		if err != nil {
			err = errors.Wrap(err, "GetRankListByBatch:rows.Scan error")
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "GetRankListByBatch:rows.Err")
	}
	return
}

// GetMemberRankTimes 用户上榜次数
func (d *dao) GetMemberRankTimes(c context.Context, sid, startBatch, endBatch int64, mids []int64) (rs []*rank.MemberRankTimes, err error) {
	rs = []*rank.MemberRankTimes{}
	rows, err := d.db.Query(c, fmt.Sprintf(getMemberRankTimesSQL, rankName, xstr.JoinInts(mids)), sid, startBatch, endBatch)
	if err != nil {
		err = errors.Wrap(err, "GetMemberRankTimes:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &rank.MemberRankTimes{}
		err = rows.Scan(&r.Mid, &r.Times)
		if err != nil {
			err = errors.Wrap(err, "GetMemberRankTimes:rows.Scan error")
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "GetMemberRankTimes:rows.Err")
	}
	return
}

// GetMemberHighest 获得用户最高记录
func (d *dao) GetMemberHighest(c context.Context, sid, startBatch, endBatch int64, mids []int64) (rs []*rank.MemberRankHighest, err error) {
	rs = []*rank.MemberRankHighest{}
	rows, err := d.db.Query(c, fmt.Sprintf(getMemberHighestRankSQL, rankName, xstr.JoinInts(mids)), sid, startBatch, endBatch)
	if err != nil {
		err = errors.Wrap(err, "GetMemberHighest:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &rank.MemberRankHighest{}
		err = rows.Scan(&r.Mid, &r.Rank)
		if err != nil {
			err = errors.Wrap(err, "GetMemberHighest:rows.Scan error")
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "GetMemberHighest:rows.Err")
	}
	return
}

// GetRankListByBatchPatch 获取某个批次的排行结果
func (d *dao) GetRankListByBatchPatch(c context.Context, sid, batch int64, offset, limit int) (rs []*rank.DB, err error) {
	rs = []*rank.DB{}
	rows, err := d.db.Query(c, fmt.Sprintf(getRankByBatchLimitSQL, rankName), sid, batch, offset, limit)
	if err != nil {
		err = errors.Wrap(err, "getRankByBatchLimitSQL:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &rank.DB{}
		err = rows.Scan(&r.SID, &r.Mid, &r.NickName, &r.Rank, &r.Score, &r.State, &r.Batch, &r.Remark)
		if err != nil {
			err = errors.Wrap(err, "getRankByBatchLimitSQL:rows.Scan error")
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "getRankByBatchLimitSQL:rows.Err")
	}
	return
}
