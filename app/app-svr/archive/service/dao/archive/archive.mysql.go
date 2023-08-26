package archive

import (
	"context"
	"fmt"

	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/time"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/archive/service/model/archive"
	"go-gateway/app/app-svr/archive/service/model/videoshot"
)

const (
	_statSharding   = 100
	_premiereArcCnt = 100
	_arcSQL         = "SELECT aid,mid,typeid,videos,copyright,title,cover,content,duration,attribute,state,access,pubtime,ctime,mission_id,order_id,redirect_url,forward,dynamic,cid,dimensions,season_id,attribute_v2,up_from,first_frame,IFNULL(INET6_NTOA(ipv6),\"\") FROM archive WHERE aid=?"
	_arcsSQL        = "SELECT aid,mid,typeid,videos,copyright,title,cover,content,duration,attribute,state,access,pubtime,ctime,mission_id,order_id,redirect_url,forward,dynamic,cid,dimensions,season_id,attribute_v2,up_from,first_frame,IFNULL(INET6_NTOA(ipv6),\"\") FROM archive WHERE aid IN (%s)"
	_arcStaffSQL    = "SELECT mid,title,attribute FROM archive_staff WHERE aid=? order by index_order asc"
	_arcsStaffSQL   = "SELECT aid,mid,title,attribute FROM archive_staff WHERE aid IN (%s) order by index_order asc"
	_statSQL        = "SELECT aid,fav,share,reply,coin,dm,click,now_rank,his_rank,likes,follow FROM archive_stat_%02d WHERE aid=?"
	_statsSQL       = "SELECT aid,fav,share,reply,coin,dm,click,now_rank,his_rank,likes,follow FROM archive_stat_%02d WHERE aid in (%s)"
	_additSQL       = "SELECT aid,description,desc_v2,sub_type FROM archive_addit WHERE aid=?"
	_additsSQL      = "SELECT aid,description,desc_v2,sub_type FROM archive_addit WHERE aid in (%s)"
	_redirectsSQL   = "SELECT aid,redirect_type,redirect_target,policy_type,policy_id FROM archive_redirect WHERE aid in (%s)"
	// nolint:gosec
	_upPasSQL = "SELECT aid,pubtime,copyright,attribute,attribute_v2 FROM archive WHERE mid=? AND state>=0 ORDER BY pubtime DESC"
	// nolint:gosec
	_upsPasSQL                  = "SELECT aid,mid,pubtime,copyright FROM archive WHERE mid IN (%s) AND state>=0 ORDER BY pubtime DESC"
	_vdosSQL                    = "SELECT cid,src_type,index_order,eptitle,duration,filename,weblink,dimensions,first_frame FROM archive_video WHERE aid=? ORDER BY index_order"
	_vdosByAidsSQL              = "SELECT aid,cid,src_type,index_order,eptitle,duration,filename,weblink,dimensions,first_frame FROM archive_video WHERE aid in (%s) ORDER BY index_order"
	_vdosByAidCidsSQL           = "SELECT aid,cid,src_type,index_order,eptitle,duration,filename,weblink,dimensions,first_frame FROM archive_video WHERE cid in (%s) ORDER BY index_order"
	_vdoSQL                     = "SELECT cid,src_type,index_order,eptitle,duration,filename,weblink,description,dimensions,first_frame FROM archive_video WHERE aid=? AND cid=?"
	_tpsSQL                     = "SELECT id,pid,name FROM archive_type"
	_videoShotSQL               = "SELECT cnt,hd_count,hd_image,sd_count,sd_image FROM archive_video_shot WHERE cid=?"
	_upCntSQL                   = "SELECT COUNT(*) FROM archive WHERE mid=? AND state>=0"
	_inRedirectSQL              = "INSERT INTO archive_redirect(aid,redirect_type,redirect_target,policy_type,policy_id) VALUE(?,?,?,?,?) ON DUPLICATE KEY UPDATE redirect_type=values(redirect_type),redirect_target=values(redirect_target),policy_type=values(policy_type),policy_id=values(policy_id)"
	_arcExpandByPremiereTimeSQL = "SELECT aid,mid,arc_type,room_id,premiere_time FROM archive_expand WHERE premiere_time > '%s' and premiere_time < '%s' order by premiere_time desc"
	_arcExpandSQL               = "SELECT aid,mid,arc_type,room_id,premiere_time FROM archive_expand WHERE aid in (%s)"
	_seasonEpisodeSQL           = "SELECT season_id,section_id,episode_id,aid,attribute FROM season_episode WHERE season_id=? and aid=?"
)

func statTbl(aid int64) int64 {
	return aid % _statSharding
}

// RawArc get a archive by aid.
func (d *Dao) RawArc(c context.Context, aid int64) (a *api.Arc, ip string, err error) {
	d.infoProm.Incr("RawArc")
	row := d.resultDB.QueryRow(c, _arcSQL, aid)
	a = &api.Arc{}
	var dimension string
	if err = row.Scan(&a.Aid, &(a.Author.Mid), &a.TypeID, &a.Videos, &a.Copyright, &a.Title, &a.Pic, &a.Desc, &a.Duration,
		&a.Attribute, &a.State, &a.Access, &a.PubDate, &a.Ctime, &a.MissionID, &a.OrderID, &a.RedirectURL, &a.Forward, &a.Dynamic, &a.FirstCid, &dimension, &a.SeasonID, &a.AttributeV2, &a.UpFromV2, &a.FirstFrame, &ip); err != nil {
		if err == sql.ErrNoRows {
			a = nil
			err = nil
		} else {
			log.Error("row.Scan error(%v)", err)
			d.errProm.Incr("RawArc")
		}
		return
	}
	a.FillDimensionAndFF(dimension)
	return
}

// RawArcs multi get archives by aids.
func (d *Dao) RawArcs(c context.Context, aids []int64) (res map[int64]*api.Arc, ips map[int64]string, err error) {
	d.infoProm.Incr("RawArcs")
	query := fmt.Sprintf(_arcsSQL, xstr.JoinInts(aids))
	rows, err := d.resultDB.Query(c, query)
	if err != nil {
		log.Error("db.Query(%s) error(%v)", query, err)
		d.errProm.Incr("RawArcs")
		return
	}
	defer rows.Close()
	res = make(map[int64]*api.Arc, len(aids))
	ips = make(map[int64]string, len(aids))

	for rows.Next() {
		a := &api.Arc{}
		var dimension string
		var ip string
		if err = rows.Scan(&a.Aid, &(a.Author.Mid), &a.TypeID, &a.Videos, &a.Copyright, &a.Title, &a.Pic, &a.Desc, &a.Duration,
			&a.Attribute, &a.State, &a.Access, &a.PubDate, &a.Ctime, &a.MissionID, &a.OrderID, &a.RedirectURL, &a.Forward, &a.Dynamic, &a.FirstCid, &dimension, &a.SeasonID, &a.AttributeV2, &a.UpFromV2, &a.FirstFrame, &ip); err != nil {
			log.Error("rows.Scan error(%v)", err)
			d.errProm.Incr("RawArcs")
			return
		}
		a.FillDimensionAndFF(dimension)
		res[a.Aid] = a
		if len(ip) > 0 {
			ips[a.Aid] = ip
		}
	}
	err = rows.Err()
	return
}

// RawStaff get archives staff by avid.
func (d *Dao) RawStaff(c context.Context, aid int64) (res []*api.StaffInfo, err error) {
	d.infoProm.Incr("RawStaff")
	rows, err := d.resultDB.Query(c, _arcStaffSQL, aid)
	if err != nil {
		log.Error("d.resultDB.Query(%d) error(%v)", aid, err)
		d.errProm.Incr("RawStaff")
		return
	}
	defer rows.Close()
	for rows.Next() {
		as := &api.StaffInfo{}
		if err = rows.Scan(&as.Mid, &as.Title, &as.Attribute); err != nil {
			log.Error("rows.Scan error(%v)", err)
			d.errProm.Incr("RawStaff")
			return
		}
		res = append(res, as)
	}
	err = rows.Err()
	return
}

// RawStaffs get archives staff by aids.
func (d *Dao) RawStaffs(c context.Context, aids []int64) (res map[int64][]*api.StaffInfo, err error) {
	d.infoProm.Incr("RawStaffs")
	query := fmt.Sprintf(_arcsStaffSQL, xstr.JoinInts(aids))
	rows, err := d.resultDB.Query(c, query)
	if err != nil {
		log.Error("d.resultDB.Query(%s) error(%v)", query, err)
		d.errProm.Incr("RawStaffs")
		return
	}
	defer rows.Close()
	res = make(map[int64][]*api.StaffInfo)
	for rows.Next() {
		as := &api.StaffInfo{}
		var aid int64
		if err = rows.Scan(&aid, &as.Mid, &as.Title, &as.Attribute); err != nil {
			log.Error("rows.Scan error(%v)", err)
			d.errProm.Incr("RawStaffs")
			return
		}
		res[aid] = append(res[aid], as)
	}
	err = rows.Err()
	return
}

// RawStat archive stat.
func (d *Dao) RawStat(c context.Context, aid int64) (st *api.Stat, err error) {
	d.infoProm.Incr("RawStat")
	row := d.statDB.QueryRow(c, fmt.Sprintf(_statSQL, statTbl(aid)), aid)
	st = &api.Stat{}
	if err = row.Scan(&st.Aid, &st.Fav, &st.Share, &st.Reply, &st.Coin, &st.Danmaku, &st.View, &st.NowRank, &st.HisRank, &st.Like, &st.Follow); err != nil {
		if err == sql.ErrNoRows {
			st = nil
			err = nil
		} else {
			d.errProm.Incr("RawStat")
			log.Error("row.Scan error(%v)", err)
		}
		return
	}
	return
}

// RawStats archive stats.
func (d *Dao) RawStats(c context.Context, aids []int64) (sts map[int64]*api.Stat, err error) {
	d.infoProm.Incr("RawStats")
	tbls := make(map[int64][]int64)
	for _, aid := range aids {
		tbls[statTbl(aid)] = append(tbls[statTbl(aid)], aid)
	}
	sts = make(map[int64]*api.Stat, len(aids))
	for tbl, ids := range tbls {
		var rows *sql.Rows
		if rows, err = d.statDB.Query(c, fmt.Sprintf(_statsSQL, tbl, xstr.JoinInts(ids))); err != nil {
			log.Error("d.statDB.Query(%s) error(%v)", fmt.Sprintf(_statsSQL, tbl, xstr.JoinInts(ids)), err)
			d.errProm.Incr("RawStats")
			return
		}
		if err = func() error {
			defer rows.Close()
			for rows.Next() {
				st := &api.Stat{}
				if err = rows.Scan(&st.Aid, &st.Fav, &st.Share, &st.Reply, &st.Coin, &st.Danmaku, &st.View, &st.NowRank, &st.HisRank, &st.Like, &st.Follow); err != nil {
					log.Error("rows.Scan error(%v)", err)
					d.errProm.Incr("RawStats")
					return err
				}
				sts[st.Aid] = st
			}
			if err = rows.Err(); err != nil {
				log.Error("rows.Err() error=%+v", err)
				return err
			}
			return nil
		}(); err != nil {
			return nil, err
		}
	}
	return
}

// RawAddit get archive addit
func (d *Dao) RawAddits(c context.Context, aids []int64) (map[int64]*archive.Addit, error) {
	addit := make(map[int64]*archive.Addit)
	rows, err := d.resultDB.Query(c, fmt.Sprintf(_additsSQL, xstr.JoinInts(aids)))
	if err != nil {
		log.Error("d.RawAddits.Query(%s) error(%v)", fmt.Sprintf(_additsSQL, xstr.JoinInts(aids)), err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		tmpAddit := &archive.Addit{}
		if err := rows.Scan(&tmpAddit.Aid, &tmpAddit.Description, &tmpAddit.DescV2, &tmpAddit.Subtype); err != nil {
			log.Error("rows.Scan error(%v)", err)
			continue
		}
		addit[tmpAddit.Aid] = tmpAddit
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return addit, nil
}

// RawAddit get archive addit
func (d *Dao) RawAddit(c context.Context, aid int64) (addit *archive.Addit, err error) {
	row := d.resultDB.QueryRow(c, _additSQL, aid)
	d.infoProm.Incr("RawAddit")
	addit = &archive.Addit{}
	if err = row.Scan(&addit.Aid, &addit.Description, &addit.DescV2, &addit.Subtype); err != nil {
		if err == sql.ErrNoRows {
			addit = nil
			err = nil
		} else {
			log.Error("row.Scan error(%v)", err)
			d.errProm.Incr("RawAddit")
		}
	}
	return
}

func (d *Dao) RawRedirects(c context.Context, aids []int64) (map[int64]*archive.ArcRedirect, error) {
	redirect := make(map[int64]*archive.ArcRedirect)
	rows, err := d.resultDB.Query(c, fmt.Sprintf(_redirectsSQL, xstr.JoinInts(aids)))
	if err != nil {
		log.Error("d.RawRedirects.Query(%s) error(%v)", fmt.Sprintf(_redirectsSQL, xstr.JoinInts(aids)), err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &archive.ArcRedirect{}
		if err := rows.Scan(&tmp.Aid, &tmp.RedirectType, &tmp.RedirectTarget, &tmp.PolicyType, &tmp.PolicyId); err != nil {
			log.Error("rows.Scan error(%v)", err)
			continue
		}
		redirect[tmp.Aid] = tmp
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return redirect, nil
}

func (d *Dao) InsertRedirect(c context.Context, redirect *archive.ArcRedirect) error {
	_, err := d.resultDB.Exec(c, _inRedirectSQL, redirect.Aid, redirect.RedirectType, redirect.RedirectTarget, redirect.PolicyType, redirect.PolicyId)
	return err
}

// RawUpperPassed get upper passed archives.
func (d *Dao) RawUpperPassed(c context.Context, mid int64) (aids []int64, ptimes []time.Time, copyrights []int8, err error) {
	d.infoProm.Incr("RawUpperPassed")
	rows, err := d.resultDB.Query(c, _upPasSQL, mid)
	if err != nil {
		log.Error("d.resultDB.Query(%d) error(%v)", mid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			aid       int64
			ptime     time.Time
			copyright int8
			a         = &api.Arc{}
		)
		if err = rows.Scan(&aid, &ptime, &copyright, &a.Attribute, &a.AttributeV2); err != nil {
			log.Error("rows.Scan error(%v)", err)
			d.errProm.Incr("RawUpperPassed")
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
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, nil, nil, err
	}
	return
}

// RawUppersPassed get uppers passed archives.
func (d *Dao) RawUppersPassed(c context.Context, mids []int64) (aidm map[int64][]int64, ptimes map[int64][]time.Time, copyrights map[int64][]int8, err error) {
	d.infoProm.Incr("RawUppersPassed")
	rows, err := d.resultDB.Query(c, fmt.Sprintf(_upsPasSQL, xstr.JoinInts(mids)))
	if err != nil {
		d.errProm.Incr("RawUppersPassed")
		log.Error("UpsPassed error(%v)", err)
		return
	}
	defer rows.Close()
	aidm = make(map[int64][]int64, len(mids))
	ptimes = make(map[int64][]time.Time, len(mids))
	copyrights = make(map[int64][]int8, len(mids))
	for rows.Next() {
		var (
			aid, mid  int64
			ptime     time.Time
			copyright int8
		)
		if err = rows.Scan(&aid, &mid, &ptime, &copyright); err != nil {
			log.Error("rows.Scan error(%v)", err)
			d.errProm.Incr("RawUppersPassed")
			return
		}
		aidm[mid] = append(aidm[mid], aid)
		ptimes[mid] = append(ptimes[mid], ptime)
		copyrights[mid] = append(copyrights[mid], copyright)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, nil, nil, err
	}
	return
}

// RawPages get pages by aid.
func (d *Dao) RawPages(c context.Context, aid int64) (ps []*api.Page, err error) {
	d.infoProm.Incr("RawPages")
	rows, err := d.resultDB.Query(c, _vdosSQL, aid)
	if err != nil {
		d.errProm.Incr("RawPages")
		log.Error("d.resultDB.Query(%d) error(%v)", aid, err)
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
			d.errProm.Incr("RawPages")
			log.Error("rows.Scan error(%v)", err)
			return
		}
		p.Page = page
		if p.From != "vupload" {
			p.Vid = fn
		}
		p.FillDimensionAndFF(dimensions)
		ps = append(ps, p)
	}
	err = rows.Err()
	return
}

// RawVideosByAids get videos by aids
func (d *Dao) RawVideosByCids(c context.Context, cids []int64) (map[int64][]*api.Page, error) {
	if len(cids) == 0 {
		return nil, ecode.RequestErr
	}
	rows, err := d.resultDB.Query(c, fmt.Sprintf(_vdosByAidCidsSQL, xstr.JoinInts(cids)))
	if err != nil {
		log.Error("d.resultDB.Query(%s) error(%v)", fmt.Sprintf(_vdosByAidCidsSQL, xstr.JoinInts(cids)), err)
		return nil, err
	}
	vs := make(map[int64][]*api.Page)
	defer rows.Close()
	for rows.Next() {
		var (
			p          = &api.Page{}
			aid        int64
			fn         string
			dimensions string
		)
		if err = rows.Scan(&aid, &p.Cid, &p.From, &p.Page, &p.Part, &p.Duration, &fn, &p.WebLink, &dimensions, &p.FirstFrame); err != nil {
			log.Error("rows.Scan error(%+v)", err)
			return nil, err
		}
		if p.From != "vupload" {
			p.Vid = fn
		}
		p.FillDimensionAndFF(dimensions)
		vs[aid] = append(vs[aid], p)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%+v)", err)
		return nil, err
	}
	return vs, nil
}

// RawVideosByAids get videos by aids
func (d *Dao) RawVideosByAids(c context.Context, aids []int64) (vs map[int64][]*api.Page, err error) {
	d.infoProm.Incr("RawVideosByAids")
	rows, err := d.resultDB.Query(c, fmt.Sprintf(_vdosByAidsSQL, xstr.JoinInts(aids)))
	if err != nil {
		log.Error("d.resultDB.Query(%s) error(%v)", fmt.Sprintf(_vdosByAidsSQL, xstr.JoinInts(aids)), err)
		return
	}
	vs = make(map[int64][]*api.Page, len(aids))
	var pages = make(map[int64]int32, len(aids))
	defer rows.Close()
	for rows.Next() {
		var (
			p          = &api.Page{}
			aid        int64
			fn         string
			dimensions string
		)
		if err = rows.Scan(&aid, &p.Cid, &p.From, &p.Page, &p.Part, &p.Duration, &fn, &p.WebLink, &dimensions, &p.FirstFrame); err != nil {
			d.errProm.Incr("RawVideosByAids")
			log.Error("rows.Scan error(%v)", err)
			return
		}
		pages[aid]++
		p.Page = pages[aid]
		if p.From != "vupload" {
			p.Vid = fn
		}
		p.FillDimensionAndFF(dimensions)
		vs[aid] = append(vs[aid], p)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// RawPage get video by aid & cid.
func (d *Dao) RawPage(c context.Context, aid, cid int64) (p *api.Page, err error) {
	d.infoProm.Incr("RawPage")
	var fn, dimension string
	row := d.resultDB.QueryRow(c, _vdoSQL, aid, cid)
	p = &api.Page{}
	if err = row.Scan(&p.Cid, &p.From, &p.Page, &p.Part, &p.Duration, &fn, &p.WebLink, &p.Desc, &dimension, &p.FirstFrame); err != nil {
		if err == sql.ErrNoRows {
			p = nil
			err = nil
		} else {
			d.errProm.Incr("RawPage")
			log.Error("row.Scan error(%v)", err)
		}
		return
	}
	if p.From != "vupload" {
		p.Vid = fn
	}
	p.FillDimensionAndFF(dimension)
	return
}

// RawTypes is
func (d *Dao) RawTypes(c context.Context) (types map[int16]*archive.ArcType, err error) {
	d.infoProm.Incr("RawTypes")
	var rows *sql.Rows
	if rows, err = d.resultDB.Query(c, _tpsSQL); err != nil {
		log.Error("d.arcReadDB.Query error(%v)", err)
		return
	}
	defer rows.Close()
	types = make(map[int16]*archive.ArcType)
	for rows.Next() {
		var (
			rid, pid int16
			name     string
		)
		if err = rows.Scan(&rid, &pid, &name); err != nil {
			log.Error("rows.Scan error(%v)", err)
			d.errProm.Incr("RawTypes")
			return
		}
		types[rid] = &archive.ArcType{ID: rid, Pid: pid, Name: name}
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// RawVideoShot is
func (d *Dao) RawVideoShot(c context.Context, cid int64) (*videoshot.Videoshot, error) {
	d.infoProm.Incr("RawVideoShot")
	row := d.resultDB.QueryRow(c, _videoShotSQL, cid)
	vs := &videoshot.Videoshot{}
	if err := row.Scan(&vs.Count, &vs.HDCount, &vs.HDImg, &vs.SdCount, &vs.SdImg); err != nil {
		return nil, err
	}
	return vs, nil
}

// RawUpperCount get the count of archives by mid of Up.
func (d *Dao) RawUpperCount(c context.Context, mid int64) (int, error) {
	d.infoProm.Incr("RawUpperCount")
	count := 0
	row := d.resultDB.QueryRow(c, _upCntSQL, mid)
	err := row.Scan(&count)
	if err != nil {
		d.errProm.Incr("RawUpperCount")
		log.Error("row.Scan error(%v)", err)
		return 0, err
	}
	return count, nil
}

func (d *Dao) RawArchiveExpandByPremiereTime(c context.Context, startTime string, endTime string) ([]*archive.ArcExpand, error) {
	res := make([]*archive.ArcExpand, 0, _premiereArcCnt)
	d.infoProm.Incr("RawArchiveExpandByPremiereTime")
	rows, err := d.resultDB.Query(c, fmt.Sprintf(_arcExpandByPremiereTimeSQL, startTime, endTime))
	if err != nil {
		log.Error("d.RawArchiveExpandByMids.Query(%s) error(%v)", fmt.Sprintf(_arcExpandByPremiereTimeSQL, startTime, endTime), err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &archive.ArcExpand{}
		if err = rows.Scan(&tmp.Aid, &tmp.Mid, &tmp.ArcType, &tmp.RoomId, &tmp.PremiereTime); err != nil {
			log.Error("rows.Scan error(%v)", err)
			continue
		}
		res = append(res, tmp)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) RawArchiveExpand(c context.Context, aids []int64) (map[int64]*archive.ArcExpand, error) {
	res := make(map[int64]*archive.ArcExpand, len(aids))
	if len(aids) == 0 {
		return res, nil
	}
	d.infoProm.Incr("RawArchiveExpand")
	rows, err := d.resultDB.Query(c, fmt.Sprintf(_arcExpandSQL, xstr.JoinInts(aids)))
	if err != nil {
		log.Error("d.RawArchiveExpandByMids.Query(%s) error(%v)", fmt.Sprintf(_arcExpandSQL, xstr.JoinInts(aids)), err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &archive.ArcExpand{}
		if err = rows.Scan(&tmp.Aid, &tmp.Mid, &tmp.ArcType, &tmp.RoomId, &tmp.PremiereTime); err != nil {
			log.Error("rows.Scan error(%v)", err)
			continue
		}
		res[tmp.Aid] = tmp
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) RawSeasonEpisode(c context.Context, sid int64, aid int64) (*archive.SeasonEpisode, error) {
	row := d.resultDB.QueryRow(c, _seasonEpisodeSQL, sid, aid)
	tmp := &archive.SeasonEpisode{}
	if err := row.Scan(&tmp.SeasonId, &tmp.SectionId, &tmp.EpisodeId, &tmp.Aid, &tmp.Attribute); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("RawSeasonEpisode rows.Scan sid(%d) aid(%d) error(%v)", sid, aid, err)
	}
	return tmp, nil
}
