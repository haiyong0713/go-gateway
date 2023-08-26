package native

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"

	v1 "go-gateway/app/web-svr/native-page/interface/api"
)

var (
	_pagesSQL       = "select `id`,`title`,`type`,`foreign_id`,`stime`,`creator`,`operator`,`share_title`,`share_image`,`share_url`,`ctime`,`mtime`,`state`,`skip_url`,`spmid`,`related_uid`,`act_type`,`hot`,`dynamic_id`,`attribute`,`etime`,`pc_url`,`another_title`,`share_caption`,`bg_color`,`from_type`,`ver`,`conf_set`,`act_origin`,`first_pid` from `native_page` where id in (%s)"
	_foreignSQL     = "select `id`,`foreign_id` from `native_page` where foreign_id in (%s) and `type` = ? and `state` = ?"
	_tsUIDsSQL      = "SELECT `id`,`type`,`from_type`,`mtime` FROM native_page WHERE `related_uid` = ? AND `type`=? AND`state` IN (%s) ORDER BY `mtime` DESC"
	_tsTitleSQL     = "SELECT `id`,`from_type` FROM `native_page` WHERE `foreign_id`=? AND `type`=? AND state IN (%s) Limit 1"
	_addPageSQL     = "INSERT INTO `native_page` (`title`,`type`,`foreign_id`,`state`,`related_uid`,`from_type`,`bg_color`,`attribute`,`act_origin`) VALUES (?,?,?,?,?,?,?,?,?)"
	_updatePageSQL  = "UPDATE `native_page` SET `bg_color`=?,`attribute`=?,`share_image`=? WHERE `id`=?"
	_updateAttrSQL  = "UPDATE `native_page` SET `attribute`=? WHERE `id`=? AND `attribute`=?"
	_updateWaitSQL  = "UPDATE `native_page` SET `bg_color`=?,`title`=?,`foreign_id`=?,`attribute`=? WHERE `id`=?"
	_bindSQL        = "UPDATE `native_page` SET `type`=?,`state`=? WHERE `id`=? AND `state`=?"
	_actTypeSQL     = "select `id` from `native_page` where `act_type`=? and `state`=? order by `id` desc limit 500"
	_updateStateSQL = "UPDATE `native_page` SET `state`=? WHERE `id`=? AND `state`=?"
)

// RawNativePages .
func (d *Dao) RawNativePages(c context.Context, ids []int64) (list map[int64]*v1.NativePage, err error) {
	if len(ids) == 0 {
		return
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_pagesSQL, xstr.JoinInts(ids)))
	if err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	list = make(map[int64]*v1.NativePage)
	for rows.Next() {
		t := &v1.NativePage{}
		if err = rows.Scan(&t.ID, &t.Title, &t.Type, &t.ForeignID, &t.Stime, &t.Creator, &t.Operator, &t.ShareTitle, &t.ShareImage, &t.ShareURL, &t.Ctime, &t.Mtime, &t.State, &t.SkipURL, &t.Spmid, &t.RelatedUid, &t.ActType, &t.Hot, &t.DynamicID, &t.Attribute, &t.Etime, &t.PcURL, &t.AnotherTitle, &t.ShareCaption, &t.BgColor, &t.FromType, &t.Ver, &t.ConfSet, &t.ActOrigin, &t.FirstPid); err != nil {
			err = errors.Wrap(err, "rows.Scan")
			return
		}
		list[t.ID] = t
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err")
	}
	return
}

// RawNativeForeigns .
func (d *Dao) RawNativeForeigns(c context.Context, fids []int64, pageType int64) (ids map[int64]int64, err error) {
	if len(fids) == 0 {
		return
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_foreignSQL, xstr.JoinInts(fids)), pageType, v1.OnlineState)
	if err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	ids = make(map[int64]int64)
	for rows.Next() {
		t := &v1.NativePage{}
		if err = rows.Scan(&t.ID, &t.ForeignID); err != nil {
			err = errors.Wrap(err, "rows.Scan")
			return
		}
		ids[t.ForeignID] = t.ID
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err")
	}
	return
}

// NtTsOnlineIDsSearch .
func (d *Dao) NtTsOnlineIDsSearch(c context.Context, uid int64) (list []*v1.NativePage, err error) {
	return d.baseSearch(c, uid, []int64{v1.OnlineState})
}

// NtTsUIDsSearch .
func (d *Dao) NtTsUIDsSearch(c context.Context, uid int64) (list []*v1.NativePage, err error) {
	return d.baseSearch(c, uid, []int64{v1.WaitForCheck, v1.CheckOffline, v1.WaitForOnline, v1.OnlineState, v1.OfflineState})
}

// NtTsUIDsSearch .
func (d *Dao) baseSearch(c context.Context, uid int64, state []int64) (list []*v1.NativePage, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, fmt.Sprintf(_tsUIDsSQL, xstr.JoinInts(state)), uid, v1.TopicActType); err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		t := &v1.NativePage{}
		if err = rows.Scan(&t.ID, &t.Type, &t.FromType, &t.Mtime); err != nil {
			err = errors.Wrap(err, "rows.Scan")
			return
		}
		if t.IsUpTopicAct() {
			list = append(list, t)
		}
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err")
	}
	return
}

// RawTagIDSearch .
func (d *Dao) RawNatTagIDExist(c context.Context, tagID int64) (int64, error) {
	row := d.db.QueryRow(c, fmt.Sprintf(_tsTitleSQL, xstr.JoinInts([]int64{v1.WaitForCheck, v1.WaitForOnline, v1.OnlineState})), tagID, v1.TopicActType)
	t := &v1.NativePage{}
	if err := row.Scan(&t.ID, &t.FromType); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		if err != nil {
			return 0, err
		}
	}
	if v1.IsFromTopicUpg(t.FromType) {
		return 0, nil
	}
	return t.ID, nil
}

// PageSave .
func (d *Dao) PageSave(c context.Context, p *v1.NativePage) (int64, error) {
	res, err := d.db.Exec(c, _addPageSQL, p.Title, p.Type, p.ForeignID, p.State, p.RelatedUid, p.FromType, p.BgColor, p.Attribute, p.ActOrigin)
	if err != nil {
		log.Error("PageSave arg:%v error(%v)", p, err)
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) PageUpdate(c context.Context, p *v1.NativePage) error {
	_, err := d.db.Exec(c, _updatePageSQL, p.BgColor, p.Attribute, p.ShareImage, p.ID)
	if err != nil {
		log.Error("PageColorUpdate arg:%v error(%v)", p, err)
	}
	return err
}

// PageAttrUpdate .
func (d *Dao) PageAttrUpdate(c context.Context, id, newAttr, oldAttr int64) error {
	_, err := d.db.Exec(c, _updateAttrSQL, newAttr, id, oldAttr)
	if err != nil {
		log.Error("PageAttrUpdate arg:%d,%d,%d error(%v)", id, newAttr, oldAttr, err)
	}
	return err
}

// PageWaitUpdate .
func (d *Dao) PageWaitUpdate(c context.Context, p *v1.NativePage) error {
	_, err := d.db.Exec(c, _updateWaitSQL, p.BgColor, p.Title, p.ForeignID, p.Attribute, p.ID)
	if err != nil {
		log.Error("PageColorUpdate arg:%v error(%v)", p, err)
	}
	return err
}

// _bindSQL
func (d *Dao) PageBind(c context.Context, p *v1.NativePage) error {
	_, err := d.db.Exec(c, _bindSQL, p.Type, p.State, p.ID, v1.WaitForCommit)
	if err != nil {
		log.Error("PageBind arg:%v error(%v)", p, err)
	}
	return err
}

func (d *Dao) RawNatIDsByActType(c context.Context, actType int64) ([]int64, error) {
	rows, err := d.db.Query(c, _actTypeSQL, actType, v1.OnlineState)
	if err != nil {
		if err == xsql.ErrNoRows {
			return []int64{}, nil
		}
		log.Error("Fail to query natIDs by actType, actType=%d err=%+v", actType, err)
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		t := &v1.NativePage{}
		if err = rows.Scan(&t.ID); err != nil {
			log.Error("Fail to scan natIDs, err=%+v", err)
			return nil, err
		}
		ids = append(ids, t.ID)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows has err, err=%+v", err)
		return nil, err
	}
	return ids, nil
}

func (d *Dao) UpdatePageState(c context.Context, id int64, newState, oldState int64) error {
	if _, err := d.db.Exec(c, _updateStateSQL, newState, id, oldState); err != nil {
		log.Error("Fail to update native_page state, id=%+v newState=%+v oldState=%+v error=%+v", id, newState, oldState, err)
		return err
	}
	return nil
}
