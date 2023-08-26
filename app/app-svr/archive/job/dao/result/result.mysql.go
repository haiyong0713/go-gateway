package result

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/database/sql"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/archive/job/model/archive"
	"go-gateway/app/app-svr/archive/service/api"
)

// nolint:gosec
const (
	_statSharding = 100
	_addStaffSQL  = "INSERT INTO archive_staff (aid,mid,title,ctime,index_order,attribute) VALUES "
	_delStaffSQL  = "DELETE FROM archive_staff WHERE aid=?"
	_arcStaffSQL  = "SELECT mid,title,attribute FROM archive_staff WHERE aid=? order by index_order asc"
	_arcSQL       = "SELECT aid,mid,typeid,videos,copyright,title,cover,content,duration,attribute,state,access,pubtime,ctime,mission_id,order_id,redirect_url,forward,dynamic,cid,dimensions,season_id,attribute_v2,up_from,first_frame,IFNULL(INET6_NTOA(ipv6),\"\") FROM archive WHERE aid=?"
	_videosSQL    = "SELECT cid,src_type,index_order,eptitle,duration,filename,weblink,dimensions,first_frame FROM archive_video WHERE aid=? ORDER BY index_order"
	_inArchiveSQL = `INSERT IGNORE INTO archive (aid,mid,typeid,videos,title,cover,content,duration,attribute,copyright,access,pubtime,state,mission_id,order_id,redirect_url,forward,dynamic,cid,dimensions,attribute_v2,up_from,first_frame,ipv6)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_upArchiveSQL     = "UPDATE archive SET mid=?,typeid=?,videos=?,title=?,cover=?,content=?,duration=?,attribute=?,copyright=?,access=?,pubtime=?,state=?,mission_id=?,order_id=?,redirect_url=?,mtime=?,forward=?,dynamic=?,cid=?,dimensions=?,attribute_v2=?,up_from=?,first_frame=?,ipv6=? WHERE aid=?"
	_upPassedSQL      = "SELECT aid FROM archive WHERE mid=? AND state>=0 ORDER BY pubtime DESC"
	_upArcSidSQL      = "UPDATE archive SET season_id=? WHERE aid=?"
	_delArcSidSQL     = "UPDATE archive SET season_id=? WHERE aid=? AND season_id=?"
	_upArcFirstCidSQL = "UPDATE archive SET cid=? WHERE aid=?"
	_delTypesSQL      = "DELETE FROM archive_type WHERE id NOT IN (%s)"
	_tpsSQL           = "SELECT id,name FROM archive_type"
	_inTypeSQL        = "INSERT INTO archive_type(id,pid,name) value(?,?,?) ON DUPLICATE KEY UPDATE pid=values(pid),name=values(name)"
	_inVideoSQL       = `INSERT INTO archive_video (aid,eptitle,description,filename,src_type,cid,duration,index_order,attribute,weblink,dimensions,first_frame) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
				ON DUPLICATE KEY UPDATE eptitle=?,description=?,filename=?,src_type=?,duration=?,index_order=?,attribute=?,weblink=?,dimensions=?,first_frame=?`
	_delVideoByCidSQL  = "DELETE FROM archive_video WHERE aid=? and cid=?"
	_sortVideosSQL     = "UPDATE archive_video SET index_order=2 WHERE aid=? AND cid<>?"
	_stickVideoSQL     = "UPDATE archive_video SET index_order=1 WHERE aid=? and cid=?"
	_inVideoShotSQL    = "INSERT INTO archive_video_shot(cid,cnt,hd_count,hd_image,sd_count,sd_image) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE cnt=VALUES(cnt),hd_count=VALUES(hd_count),hd_image=VALUES(hd_image),sd_count=VALUES(sd_count),sd_image=VALUES(sd_image)"
	_delVideoShotSQL   = "DELETE FROM archive_video_shot WHERE cid=?"
	_inAdditSQL        = "INSERT INTO archive_addit(aid,description,desc_v2,sub_type) VALUE(?,?,?,?) ON DUPLICATE KEY UPDATE description=values(description),desc_v2=values(desc_v2),sub_type=values(sub_type)"
	_additSQL          = "SELECT description,desc_v2,sub_type FROM archive_addit WHERE aid=?"
	_statSQL           = "SELECT aid,fav,share,reply,coin,dm,click,now_rank,his_rank,likes,follow FROM archive_stat_%02d WHERE aid=?"
	_upCntSQL          = "SELECT COUNT(*) FROM archive WHERE mid=? AND state>=0"
	_upPasSQL          = "SELECT aid,pubtime,copyright,attribute,attribute_v2 FROM archive WHERE mid=? AND state>=0 ORDER BY pubtime DESC"
	_maxVideoShotIDSQL = "SELECT MAX(id) FROM archive_video_shot"
	_idToAidSQL        = "SELECT aid FROM archive WHERE id=?"
	_upVideoFFSQL      = "UPDATE archive_video SET first_frame=? WHERE cid=?"
	_videoByCidSQL     = "SELECT aid,cid,src_type,index_order,eptitle,duration,filename,weblink,description,dimensions,first_frame FROM archive_video WHERE cid=?"
	_upArchiveFFSQL    = "UPDATE archive SET first_frame=? WHERE aid=?"
	_inArcExpandSQL    = "INSERT INTO archive_expand(aid,mid,arc_type,room_id,premiere_time) VALUE(?,?,?,?,?) ON DUPLICATE KEY UPDATE aid=values(aid),mid=values(mid),arc_type=values(arc_type),room_id=values(room_id),premiere_time=values(premiere_time)"
	_arcExpandSQL      = "SELECT aid,mid,arc_type,room_id,premiere_time FROM archive_expand WHERE aid=?"
	_seasonEpisodeSQL  = "SELECT season_id,section_id,episode_id,aid,attribute FROM season_episode WHERE season_id=? and aid=?"
)

func statTbl(aid int64) int64 {
	return aid % _statSharding
}

// Types is
func (d *Dao) RawTypes(c context.Context) (types map[int32]string, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(c, _tpsSQL); err != nil {
		log.Error("d.arcDB.Query error(%+v)", err)
		return
	}
	defer rows.Close()
	types = make(map[int32]string)
	for rows.Next() {
		var (
			id   int32
			name string
		)
		if err = rows.Scan(&id, &name); err != nil {
			log.Error("rows.Scan error(%+v)", err)
			return nil, err
		}
		types[id] = name
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return types, nil
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

// RawUpCount is
func (d *Dao) RawUpCount(c context.Context, mid int64) (cnt int64, err error) {
	row := d.db.QueryRow(c, _upCntSQL, mid)
	if err = row.Scan(&cnt); err != nil {
		return 0, err
	}
	return cnt, err
}

// RawUpperPassed get upper passed archives.
func (d *Dao) RawUpperPassed(c context.Context, mid int64) (aids []int64, ptimes []xtime.Time, copyrights []int64, err error) {
	rows, err := d.db.Query(c, _upPasSQL, mid)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			aid       int64
			ptime     xtime.Time
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

// RawStat archive stat.
func (d *Dao) RawStat(c context.Context, aid int64) (st *api.Stat, err error) {
	row := d.statDB.QueryRow(c, fmt.Sprintf(_statSQL, statTbl(aid)), aid)
	st = &api.Stat{}
	if err = row.Scan(&st.Aid, &st.Fav, &st.Share, &st.Reply, &st.Coin, &st.Danmaku, &st.View, &st.NowRank, &st.HisRank, &st.Like, &st.Follow); err != nil {
		return nil, err
	}
	return st, nil
}

// RawVideos get videos by aid.
func (d *Dao) RawVideos(c context.Context, aid int64) (ps []*api.Page, err error) {
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

// RawStaff get archives staff by avid.
func (d *Dao) RawStaff(c context.Context, aid int64) (res []*api.StaffInfo, err error) {
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

// TxAddAddit is
func (d *Dao) TxAddAddit(tx *sql.Tx, aid int64, a *archive.Addit, biz *archive.Biz, pay *archive.Biz) error {
	var (
		desc    string
		descV2  string
		subType int
		err     error
	)
	if a != nil {
		desc = a.Desc
	}
	if biz != nil {
		descV2 = biz.Data
	}
	if pay != nil {
		subType = pay.SubType
	}
	_, err = tx.Exec(_inAdditSQL, aid, desc, descV2, subType)
	return err
}

// TxAddVideo add videos result
func (d *Dao) TxAddVideo(tx *sql.Tx, v *archive.Video, ff string) (rows int64, err error) {
	res, err := tx.Exec(_inVideoSQL, v.Aid, v.Title, v.Desc, v.Filename, v.SrcType, v.Cid, v.Duration, v.Index, v.Attribute, v.WebLink, v.Dimensions, ff,
		v.Title, v.Desc, v.Filename, v.SrcType, v.Duration, v.Index, v.Attribute, v.WebLink, v.Dimensions, ff)
	if err != nil {
		log.Error("tx.Exec error(%+v)", err)
		return
	}
	return res.RowsAffected()
}

// TxDelVideoByCid del videos by aid and cid
func (d *Dao) TxDelVideoByCid(tx *sql.Tx, aid, cid int64) (rows int64, err error) {
	res, err := tx.Exec(_delVideoByCidSQL, aid, cid)
	if err != nil {
		log.Error("tx.Exec error(%+v)", err)
		return
	}
	return res.RowsAffected()
}

// TxSortVideos put all the videos' order to 2 except the root one
func (d *Dao) TxSortVideos(tx *sql.Tx, aid, cid int64) (rows int64, err error) {
	res, err := tx.Exec(_sortVideosSQL, aid, cid)
	if err != nil {
		log.Error("tx.Exec(%s, %d, %d ) error(%+v)", _sortVideosSQL, aid, cid, err)
		return
	}
	return res.RowsAffected()
}

// TxStickVideo sticks the root video of the graph inside the archive
func (d *Dao) TxStickVideo(tx *sql.Tx, aid, cid int64) (rows int64, err error) {
	res, err := tx.Exec(_stickVideoSQL, aid, cid)
	if err != nil {
		log.Error("tx.Exec(%s, %d, %d ) error(%+v)", _stickVideoSQL, aid, cid, err)
		return
	}
	return res.RowsAffected()
}

// AddVideoShot is
func (d *Dao) AddVideoShot(c context.Context, cid, cnt, hdCnt, sdCnt int64, hdImg, sdImg string) (err error) {
	_, err = d.db.Exec(c, _inVideoShotSQL, cid, cnt, hdCnt, hdImg, sdCnt, sdImg)
	return err
}

func (d *Dao) DelVideoShot(c context.Context, cid int64) (err error) {
	_, err = d.db.Exec(c, _delVideoShotSQL, cid)
	return err
}

func (d *Dao) MaxVideoShotID(c context.Context) (id int64, err error) {
	row := d.db.QueryRow(c, _maxVideoShotIDSQL)
	if err = row.Scan(&id); err != nil {
		log.Error("%+v", err)
		return 0, err
	}
	return id, nil
}

func (d *Dao) DelTypes(c context.Context, tids []int64) (err error) {
	_, err = d.db.Exec(c, fmt.Sprintf(_delTypesSQL, xstr.JoinInts(tids)))
	return err
}

func (d *Dao) AddType(c context.Context, t *archive.ArcType) (err error) {
	_, err = d.db.Exec(c, _inTypeSQL, t.ID, t.PID, t.Name)
	return err
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

// RawArc get a archive by aid.
func (d *Dao) RawArc(c context.Context, aid int64) (a *api.Arc, ip string, err error) {
	row := d.db.QueryRow(c, _arcSQL, aid)
	a = &api.Arc{}
	dimension := ""

	if err = row.Scan(&a.Aid, &(a.Author.Mid), &a.TypeID, &a.Videos, &a.Copyright, &a.Title, &a.Pic, &a.Desc, &a.Duration,
		&a.Attribute, &a.State, &a.Access, &a.PubDate, &a.Ctime, &a.MissionID, &a.OrderID, &a.RedirectURL, &a.Forward, &a.Dynamic, &a.FirstCid, &dimension, &a.SeasonID, &a.AttributeV2, &a.UpFromV2, &a.FirstFrame, &ip); err != nil {
		if err == sql.ErrNoRows {
			a = nil
			err = nil
		} else {
			log.Error("row.Scan aid(%d) error(%+v)", aid, err)
		}
		return
	}
	a.FillDimensionAndFF(dimension)
	return
}

// TxAddArchive add archive result
func (d *Dao) TxAddArchive(tx *sql.Tx, a *archive.Archive, ad *archive.Addit, videoCnt int, firstCid int64, dimensions, firstFrame string) (rows int64, err error) {
	res, err := tx.Exec(_inArchiveSQL, a.ID, a.Mid, a.TypeID, videoCnt, a.Title, a.Cover, a.Content, a.Duration, a.Attribute, a.Copyright, a.Access, a.PubTime, a.State, ad.MissionID, ad.OrderID, ad.RedirectURL, a.Forward, ad.Dynamic, firstCid, dimensions, ad.InnerAttr, ad.UpFrom, firstFrame, ad.Ipv6)
	if err != nil {
		log.Error("tx.Exec(%s) error(%+v)", _inArchiveSQL, err)
		return
	}
	return res.RowsAffected()
}

// TxUpArchive update archive result
func (d *Dao) TxUpArchive(tx *sql.Tx, a *archive.Archive, ad *archive.Addit, videoCnt int, firstCid int64, dimensions, firstFrame string) (rows int64, err error) {
	res, err := tx.Exec(_upArchiveSQL, a.Mid, a.TypeID, videoCnt, a.Title, a.Cover, a.Content, a.Duration, a.Attribute, a.Copyright, a.Access, a.PubTime, a.State, ad.MissionID, ad.OrderID, ad.RedirectURL, time.Now(), a.Forward, ad.Dynamic, firstCid, dimensions, ad.InnerAttr, ad.UpFrom, firstFrame, ad.Ipv6, a.ID)
	if err != nil {
		log.Error("tx.Exec(%s) error(%+v)", _upArchiveSQL, err)
		return
	}
	return res.RowsAffected()
}

// UpArcSID update archive.season_id
func (d *Dao) UpArcSID(c context.Context, sid int64, aid int64) (err error) {
	_, err = d.db.Exec(c, _upArcSidSQL, sid, aid)
	if err != nil {
		log.Error("tx.Exec error(%+v)", err)
		return
	}
	return
}

// DelArcSID update archive.season_id=0
func (d *Dao) DelArcSID(c context.Context, sid, aid int64) (err error) {
	_, err = d.db.Exec(c, _delArcSidSQL, 0, aid, sid)
	if err != nil {
		log.Error("tx.Exec error(%+v)", err)
		return
	}
	return
}

// TxUpArcFirstCID update archive.cid=cid
func (d *Dao) TxUpArcFirstCID(tx *sql.Tx, aid, cid int64) (rows int64, err error) {
	res, err := tx.Exec(_upArcFirstCidSQL, cid, aid)
	if err != nil {
		log.Error("tx.Exec Aid %d Cid %d error(%+v)", aid, cid, err)
		return
	}
	return res.RowsAffected()
}

// TxDelStaff del archive staff
func (d *Dao) TxDelStaff(tx *sql.Tx, aid int64) (err error) {
	_, err = tx.Exec(_delStaffSQL, aid)
	if err != nil {
		log.Error("tx.Exec error(%+v)", err)
		return
	}
	return
}

// TxAddStaff add archive staff
func (d *Dao) TxAddStaff(tx *sql.Tx, staff []*archive.Staff) (err error) {
	var valSQL []string
	for _, s := range staff {
		valSQL = append(valSQL, fmt.Sprintf("(%d,%d,'%s','%s',%d,%d)", s.Aid, s.Mid, s.Title, s.Ctime, s.IndexOrder, s.Attribute))
	}
	valSQLStr := strings.Join(valSQL, ",")
	_, err = tx.Exec(_addStaffSQL + valSQLStr)
	if err != nil {
		log.Error("tx.Exec error(%+v)", err)
		return
	}
	return
}

func (d *Dao) IDToAid(c context.Context, id int64) (aid int64, err error) {
	row := d.db.QueryRow(c, _idToAidSQL, id)
	if err = row.Scan(&aid); err != nil {
		return 0, err
	}
	return aid, nil
}

// RawVideoFistFrame get video first frame
func (d *Dao) RawVideoFistFrame(c context.Context, cid int64) (int64, *api.Page, error) {
	row := d.db.QueryRow(c, _videoByCidSQL, cid)
	var (
		fn, dimension string
		p             = &api.Page{}
		aid           int64
		err           error
	)
	if err = row.Scan(&aid, &p.Cid, &p.From, &p.Page, &p.Part, &p.Duration, &fn, &p.WebLink, &p.Desc, &dimension, &p.FirstFrame); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil, nil
		}
		log.Error("row.Scan error(%+v)", err)
		return 0, nil, err
	}
	if p.From != "vupload" {
		p.Vid = fn
	}
	p.FillDimensionAndFF(dimension)
	return aid, p, err
}

// UpVideoFF is
func (d *Dao) UpVideoFF(c context.Context, cid int64, ff string) error {
	_, err := d.db.Exec(c, _upVideoFFSQL, ff, cid)
	return err
}

// UpArcFF is
func (d *Dao) UpArcFF(c context.Context, aid int64, ff string) error {
	_, err := d.db.Exec(c, _upArchiveFFSQL, ff, aid)
	return err
}

func (d *Dao) TxArchiveExpand(tx *sql.Tx, arcExpand *archive.ArcExpand) error {
	_, err := tx.Exec(_inArcExpandSQL, arcExpand.Aid, arcExpand.Mid, arcExpand.ArcType, arcExpand.RoomId, arcExpand.PremiereTime)
	return err
}

func (d *Dao) RawArchiveExpand(c context.Context, aid int64) (*archive.ArcExpand, error) {
	row := d.db.QueryRow(c, _arcExpandSQL, aid)
	tmp := &archive.ArcExpand{}
	if err := row.Scan(&tmp.Aid, &tmp.Mid, &tmp.ArcType, &tmp.RoomId, &tmp.PremiereTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("RawArchiveExpand rows.Scan aid(%d) error(%v)", aid, err)
		return nil, err //错误抛出
	}
	return tmp, nil
}

func (d *Dao) RawSeasonEpisode(c context.Context, sid int64, aid int64) (*archive.SeasonEpisode, error) {
	row := d.db.QueryRow(c, _seasonEpisodeSQL, sid, aid)
	tmp := &archive.SeasonEpisode{}
	if err := row.Scan(&tmp.SeasonId, &tmp.SectionId, &tmp.EpisodeId, &tmp.Aid, &tmp.Attribute); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("RawSeasonEpisode rows.Scan sid(%d) aid(%d) error(%v)", sid, aid, err)
		return nil, err //错误抛出
	}
	return tmp, nil
}
