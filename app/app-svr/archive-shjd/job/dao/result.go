package dao

import (
	"context"
	"fmt"
	"go-gateway/app/app-svr/archive-shjd/job/model"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/time"

	"go-gateway/app/app-svr/archive/service/api"
)

const (
	_statSharding = 100
	_arcSQL       = "SELECT aid,mid,typeid,videos,copyright,title,cover,content,duration,attribute,state,access,pubtime,ctime,mission_id,order_id,redirect_url,forward,dynamic,cid,dimensions,season_id,attribute_v2,up_from,first_frame,IFNULL(INET6_NTOA(ipv6),\"\") FROM archive WHERE aid=?"
	_arcStaffSQL  = "SELECT mid,title,attribute FROM archive_staff WHERE aid=? order by index_order asc"
	// nolint:gosec
	_upPassedSQL = "SELECT aid FROM archive WHERE mid=? AND state>=0 ORDER BY pubtime DESC"
	_videoSQL    = "SELECT cid,src_type,index_order,eptitle,duration,filename,weblink,description,dimensions,first_frame FROM archive_video WHERE aid=? AND cid=?"
	_videosSQL   = "SELECT cid,src_type,index_order,eptitle,duration,filename,weblink,dimensions,first_frame FROM archive_video WHERE aid=? ORDER BY index_order"
	_statSQL     = "SELECT aid,fav,share,reply,coin,dm,click,now_rank,his_rank,likes,follow FROM archive_stat_%02d WHERE aid=?"
	_upCntSQL    = "SELECT COUNT(*) FROM archive WHERE mid=? AND state>=0"
	// nolint:gosec
	_upPasSQL         = "SELECT aid,pubtime,copyright,attribute,attribute_v2 FROM archive WHERE mid=? AND state>=0 ORDER BY pubtime DESC"
	_idToAidSQL       = "SELECT aid FROM archive WHERE id=?"
	_additSQLResult   = "SELECT description,desc_v2,sub_type FROM archive_addit WHERE aid=?"
	_arcExpandSQL     = "SELECT aid,mid,arc_type,room_id,premiere_time FROM archive_expand WHERE aid=?"
	_seasonEpisodeSQL = "SELECT season_id,section_id,episode_id,aid,attribute FROM season_episode WHERE season_id=? and aid=?"
)

func statTbl(aid int64) int64 {
	return aid % _statSharding
}

func (d *Dao) IDToAid(c context.Context, id int64) (aid int64, err error) {
	row := d.db.QueryRow(c, _idToAidSQL, id)
	if err = row.Scan(&aid); err != nil {
		return 0, err
	}
	return aid, nil
}

// Stat archive stat.
func (d *Dao) Stat(c context.Context, aid int64) (st *api.Stat, err error) {
	row := d.statDB.QueryRow(c, fmt.Sprintf(_statSQL, statTbl(aid)), aid)
	st = &api.Stat{}
	if err = row.Scan(&st.Aid, &st.Fav, &st.Share, &st.Reply, &st.Coin, &st.Danmaku, &st.View, &st.NowRank, &st.HisRank, &st.Like, &st.Follow); err != nil {
		return nil, err
	}
	return st, nil
}

// Archive get a archive by aid.
func (d *Dao) Archive(c context.Context, aid int64) (a *api.Arc, ip string, err error) {
	row := d.db.QueryRow(c, _arcSQL, aid)
	a = &api.Arc{}
	var dimension string
	if err = row.Scan(&a.Aid, &(a.Author.Mid), &a.TypeID, &a.Videos, &a.Copyright, &a.Title, &a.Pic, &a.Desc, &a.Duration,
		&a.Attribute, &a.State, &a.Access, &a.PubDate, &a.Ctime, &a.MissionID, &a.OrderID, &a.RedirectURL, &a.Forward, &a.Dynamic, &a.FirstCid, &dimension, &a.SeasonID, &a.AttributeV2, &a.UpFromV2, &a.FirstFrame, &ip); err != nil {
		if err == sql.ErrNoRows {
			a = nil
			err = nil
		} else {
			log.Error("row.Scan error(%+v)", err)
		}
		return
	}
	a.FillDimensionAndFF(dimension)
	return
}

func (d *Dao) RawAdditV2(c context.Context, aid int64) (string, string, int64, error) {
	row := d.db.QueryRow(c, _additSQLResult, aid)
	desc := ""
	descV2 := ""
	subType := int64(0)
	if err := row.Scan(&desc, &descV2, &subType); err != nil {
		return "", "", 0, err
	}
	return desc, descV2, subType, nil
}

// staff get archives staff by avid.
func (d *Dao) Staff(c context.Context, aid int64) (res []*api.StaffInfo, err error) {
	rows, err := d.db.Query(c, _arcStaffSQL, aid)
	if err != nil {
		log.Error("d.resultDB.Query(%d) error(%+v)", aid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		as := &api.StaffInfo{}
		if err = rows.Scan(&as.Mid, &as.Title, &as.Attribute); err != nil {
			log.Error("rows.Scan error(%+v)", err)
			return
		}
		res = append(res, as)
	}
	err = rows.Err()
	return
}

// UpPassed is
func (d *Dao) UpPassed(c context.Context, mid int64) (aids []int64, err error) {
	rows, err := d.db.Query(c, _upPassedSQL, mid)
	if err != nil {
		log.Error("d.db.Query(%s, %d) error(%+v)", _upPassedSQL, mid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var aid int64
		if err = rows.Scan(&aid); err != nil {
			log.Error("rows.Scan(%d) error(%+v)", aid, err)
			return
		}
		aids = append(aids, aid)
	}
	err = rows.Err()
	return
}

// Video get video by aid & cid.
func (d *Dao) Video(c context.Context, aid, cid int64) (p *api.Page, err error) {
	var fn, dimension string
	row := d.db.QueryRow(c, _videoSQL, aid, cid)
	p = &api.Page{}
	if err = row.Scan(&p.Cid, &p.From, &p.Page, &p.Part, &p.Duration, &fn, &p.WebLink, &p.Desc, &dimension, &p.FirstFrame); err != nil {
		if err == sql.ErrNoRows {
			p = nil
			err = nil
		} else {
			log.Error("row.Scan error(%+v)", err)
		}
		return
	}
	if p.From != "vupload" {
		p.Vid = fn
	}
	p.FillDimensionAndFF(dimension)
	return
}

// Videos get videos by aid.
func (d *Dao) Videos(c context.Context, aid int64) (ps []*api.Page, err error) {
	rows, err := d.db.Query(c, _videosSQL, aid)
	if err != nil {
		return
	}
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
			return
		}
		p.Page = page
		if p.From != "vupload" {
			p.Vid = fn
		}
		p.FillDimensionAndFF(dimensions)
		ps = append(ps, p)
	}
	return ps, rows.Err()
}

// UpCount is
func (d *Dao) UpCount(c context.Context, mid int64) (cnt int64, err error) {
	row := d.db.QueryRow(c, _upCntSQL, mid)
	if err = row.Scan(&cnt); err != nil {
		return 0, err
	}
	return cnt, err
}

// RawUpperPassed get upper passed archives.
func (d *Dao) RawUpperPassed(c context.Context, mid int64) (aids []int64, ptimes []time.Time, copyrights []int64, err error) {
	rows, err := d.db.Query(c, _upPasSQL, mid)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			aid       int64
			ptime     time.Time
			copyright int64
			a         = &api.Arc{}
		)
		if err = rows.Scan(&aid, &ptime, &copyright, &a.Attribute, &a.AttributeV2); err != nil {
			return
		}
		// pugv付费和非公开小视频不出现在用户投稿里
		if a.AttrVal(api.AttrBitIsPUGVPay) == api.AttrYes || a.AttrValV2(api.AttrBitV2NoPublic) == api.AttrYes {
			continue
		}
		aids = append(aids, aid)
		ptimes = append(ptimes, ptime)
		copyrights = append(copyrights, copyright)
	}
	return aids, ptimes, copyrights, rows.Err()
}

func (d *Dao) RawArchiveExpand(c context.Context, aid int64) (*model.ArcExpand, error) {
	row := d.db.QueryRow(c, _arcExpandSQL, aid)
	tmp := &model.ArcExpand{}
	if err := row.Scan(&tmp.Aid, &tmp.Mid, &tmp.ArcType, &tmp.RoomId, &tmp.PremiereTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("RawArchiveExpand rows.Scan aid(%d) error(%v)", aid, err)
		return nil, err
	}
	return tmp, nil
}

func (d *Dao) RawSeasonEpisode(c context.Context, sid int64, aid int64) (*model.SeasonEpisode, error) {
	row := d.db.QueryRow(c, _seasonEpisodeSQL, sid, aid)
	tmp := &model.SeasonEpisode{}
	if err := row.Scan(&tmp.SeasonId, &tmp.SectionId, &tmp.EpisodeId, &tmp.Aid, &tmp.Attribute); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("RawSeasonEpisode rows.Scan sid(%d) aid(%d) error(%v)", sid, aid, err)
		return nil, err
	}
	return tmp, nil
}
