package like

import (
	"context"
	"fmt"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/model/like"
	lmdl "go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/pkg/errors"
)

const (
	_actProtocolSQL  = "SELECT id,sid,protocol,mtime,ctime,types,tags,pubtime,deltime,editime,hot,bgm_id,paster_id,oids,screen_set,priority_region,region_weight,global_weight,weight_stime,weight_etime,instep_id,tag_show_platform,award,award_url FROM act_subject_protocol WHERE sid = ? LIMIT 1"
	_actProtocolsSQL = "SELECT id,sid,protocol,mtime,ctime,types,tags,pubtime,deltime,editime,hot,bgm_id,paster_id,oids,screen_set,priority_region,region_weight,global_weight,weight_stime,weight_etime,instep_id,tag_show_platform,award,award_url FROM act_subject_protocol WHERE sid IN (%s)"
)

// RawActSubjectProtocol .
func (dao *Dao) RawActSubjectProtocol(c context.Context, sid int64) (res *lmdl.ActSubjectProtocol, err error) {
	row := dao.db.QueryRow(c, _actProtocolSQL, sid)
	res = new(lmdl.ActSubjectProtocol)
	if err = row.Scan(&res.ID, &res.Sid, &res.Protocol, &res.Mtime, &res.Ctime, &res.Types, &res.Tags, &res.Pubtime,
		&res.Deltime, &res.Editime, &res.Hot, &res.BgmID, &res.PasterID, &res.Oids, &res.ScreenSet, &res.PriorityRegion,
		&res.RegionWeight, &res.GlobalWeight, &res.WeightStime, &res.WeightEtime, &res.InstepID, &res.TagShowPlatform, &res.Award, &res.AwardURL); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("RawActSubjectProtocol:row.Scan error(%v)", err)
		}
	}
	return
}

// RawActSubjects batch get subject.
func (dao *Dao) RawActSubjectProtocols(c context.Context, ids []int64) (res map[int64]*lmdl.ActSubjectProtocol, err error) {
	var rows *sql.Rows
	if rows, err = dao.db.Query(c, fmt.Sprintf(_actProtocolsSQL, xstr.JoinInts(ids))); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawActSubjectProtocols:dao.db.Query()")
		}
		return
	}
	defer rows.Close()
	res = make(map[int64]*like.ActSubjectProtocol)
	for rows.Next() {
		r := new(like.ActSubjectProtocol)
		if err = rows.Scan(&r.ID, &r.Sid, &r.Protocol, &r.Mtime, &r.Ctime, &r.Types, &r.Tags, &r.Pubtime, &r.Deltime,
			&r.Editime, &r.Hot, &r.BgmID, &r.PasterID, &r.Oids, &r.ScreenSet, &r.PriorityRegion, &r.RegionWeight,
			&r.GlobalWeight, &r.WeightStime, &r.WeightEtime, &r.InstepID, &r.TagShowPlatform, &r.Award, &r.AwardURL); err != nil {
			err = errors.Wrap(err, "RawActSubjectProtocols:QueryRow")
			return
		}
		res[r.Sid] = r
	}
	err = rows.Err()
	return
}
