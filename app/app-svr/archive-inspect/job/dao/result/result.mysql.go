package result

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	arcmdl "go-gateway/app/app-svr/archive-inspect/job/model"
	"go-gateway/app/app-svr/archive/service/api"

	"github.com/pkg/errors"
)

const (
	_additSQL         = "SELECT description,desc_v2,sub_type FROM archive_addit WHERE aid=?"
	_arcExpandSQL     = "SELECT aid,mid,arc_type,room_id,premiere_time FROM archive_expand WHERE aid=?"
	_arcStaffSQL      = "SELECT mid,title,attribute FROM archive_staff WHERE aid=? order by index_order asc"
	_lastMinuteAvsSQL = "SELECT aid FROM archive WHERE mtime>=? AND mtime <=?"
	_arcAvSQL         = "SELECT aid,mid,typeid,videos,copyright,title,cover,content,duration,attribute,state,access,pubtime,ctime,mission_id,order_id,redirect_url,forward,dynamic,cid,dimensions,season_id,attribute_v2,up_from,first_frame,IFNULL(INET6_NTOA(ipv6),\"\") FROM archive WHERE aid in(%s)"
	_videosSQL        = "SELECT cid,src_type,index_order,eptitle,duration,filename,weblink,dimensions,first_frame FROM archive_video WHERE aid=? ORDER BY index_order"
	_seasonEpisodeSQL = "SELECT season_id,section_id,episode_id,aid,attribute FROM season_episode WHERE season_id=? and aid=?"
)

func (d *Dao) RawSeasonEpisode(c context.Context, sid int64, aid int64) (*arcmdl.SeasonEpisode, error) {
	row := d.db.QueryRow(c, _seasonEpisodeSQL, sid, aid)
	tmp := &arcmdl.SeasonEpisode{}
	if err := row.Scan(&tmp.SeasonId, &tmp.SectionId, &tmp.EpisodeId, &tmp.Aid, &tmp.Attribute); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return tmp, nil
}

func (d *Dao) RawAdditV2(c context.Context, aid int64) (string, string, int64, error) {
	row := d.db.QueryRow(c, _additSQL, aid)
	desc := ""
	descV2 := ""
	subType := int64(0)
	if err := row.Scan(&desc, &descV2, &subType); err != nil {
		return "", "", 0, err
	}
	return desc, descV2, subType, nil
}

func (d *Dao) RawArchiveExpand(c context.Context, aid int64) (*arcmdl.ArcExpand, error) {
	row := d.db.QueryRow(c, _arcExpandSQL, aid)
	tmp := &arcmdl.ArcExpand{}
	if err := row.Scan(&tmp.Aid, &tmp.Mid, &tmp.ArcType, &tmp.RoomId, &tmp.PremiereTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return tmp, nil
}

// RawStaff get archives staff by avid.
func (d *Dao) RawStaff(c context.Context, aid int64) ([]*api.StaffInfo, error) {
	rows, err := d.db.Query(c, _arcStaffSQL, aid)
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("d.resultDB.Query(%d) error(%+v)", aid, err))
	}
	var res []*api.StaffInfo
	defer rows.Close()
	for rows.Next() {
		as := &api.StaffInfo{}
		if err = rows.Scan(&as.Mid, &as.Title, &as.Attribute); err != nil {
			return nil, errors.Wrapf(err, fmt.Sprintf("rows.Scan(%d) error(%+v)", aid, err))
		}
		res = append(res, as)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("rows.Err(%d) err(%+v)", aid, err))
	}
	return res, nil
}

func (d *Dao) ArcInfoAvs(c context.Context, aids []int64) ([]*arcmdl.ArchiveInfo, error) {
	if len(aids) == 0 {
		return make([]*arcmdl.ArchiveInfo, 0), nil
	}
	var (
		args []string
		sqls []interface{}
	)
	for _, tid := range aids {
		args = append(args, "?")
		sqls = append(sqls, tid)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_arcAvSQL, strings.Join(args, ",")), sqls...)
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("rows.Err err(%+v) aids(%+v)", err, aids))
	}
	var as []*arcmdl.ArchiveInfo
	defer rows.Close()
	for rows.Next() {
		a := &api.Arc{}
		var ipstr string
		dimension := ""
		if err := rows.Scan(&a.Aid, &(a.Author.Mid), &a.TypeID, &a.Videos, &a.Copyright, &a.Title, &a.Pic, &a.Desc, &a.Duration,
			&a.Attribute, &a.State, &a.Access, &a.PubDate, &a.Ctime, &a.MissionID, &a.OrderID, &a.RedirectURL, &a.Forward, &a.Dynamic, &a.FirstCid, &dimension, &a.SeasonID, &a.AttributeV2, &a.UpFromV2, &a.FirstFrame, &ipstr); err != nil {
			return nil, errors.Wrapf(err, fmt.Sprintf("rows.Err rows.Scan err(%+v) aids(%+v)", err, aids))
		}
		a.FillDimensionAndFF(dimension)
		as = append(as, &arcmdl.ArchiveInfo{Arc: a, Ip: ipstr})
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("checkEqual rows.Err err(%+v) aids(%+v)", err, aids))
	}
	return as, nil
}

// LastHourAvs get last 10 modified avs
func (d *Dao) LastMinuteAvs(c context.Context) ([]int64, error) {
	now := time.Now()
	lastMinute := time.Now().Add(-time.Duration(d.c.Custom.Internal))
	rows, err := d.db.Query(c, _lastMinuteAvsSQL, lastMinute, now)
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("rows.Err err(%+v) now(%+v)", err, now))
	}
	var as []int64
	defer rows.Close()
	for rows.Next() {
		a := &api.Arc{}
		if err = rows.Scan(&a.Aid); err != nil {
			return nil, errors.Wrapf(err, fmt.Sprintf("rows.Err rows.Scan err(%+v) now(%+v)", err, now))
		}
		as = append(as, a.Aid)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("checkEqual rows.Err err(%+v) now(%+v)", err, now))
	}
	return as, nil
}

// RawVideos get videos by aid.
func (d *Dao) RawVideos(c context.Context, aid int64) ([]*api.Page, error) {
	rows, err := d.db.Query(c, _videosSQL, aid)
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("d.db.Query(%d) error(%+v)", aid, err))
	}
	var res []*api.Page
	defer rows.Close()
	var page = int32(0)
	for rows.Next() {
		var (
			p          = &api.Page{}
			fn         string
			dimensions string
		)
		page++
		if err = rows.Scan(&p.Cid, &p.From, &p.Page, &p.Part, &p.Duration, &fn, &p.WebLink, &dimensions, &p.FirstFrame); err != nil {
			return nil, errors.Wrapf(err, fmt.Sprintf("rows.Scan error(%+v) aid(%d)", err, aid))
		}
		p.Page = page
		if p.From != "vupload" {
			p.Vid = fn
		}
		p.FillDimensionAndFF(dimensions)
		res = append(res, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("rows.Err(%+v) aid(%d)", err, aid))
	}
	return res, nil
}
