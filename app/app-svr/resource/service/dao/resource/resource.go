package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"strings"
	"time"

	"database/sql"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/resource/service/model"
)

var (
	_allResSQL    = `SELECT id,platform,name,parent,counter,position,rule,size,preview,description,mark,ctime,mtime,level,type,is_ad FROM resource ORDER BY counter desc,position ASC`
	_allAssignSQL = `SELECT id,name,contract_id,resource_id,pic,pic_main_color,litpic,url,rule,weight,agency,price,atype,username,
        inline_use_same,inline_type,inline_url,inline_barrage_switch FROM resource_assignment 
		WHERE resource_group_id=0 AND stime<? AND etime>? AND state=0 AND resource_id IN (%s) ORDER BY weight,stime desc`
	_allAssignNewSQL = `SELECT ra.id,rm.id,rm.name,rm.subtitle,ra.contract_id,ra.resource_id,rm.pic,rm.pic_main_color,rm.litpic,rm.url,ra.rule,ra.position,
		ra.agency,ra.price,ra.stime,ra.etime,ra.apply_group_id,rm.ctime,rm.mtime,rm.atype,ra.username,rm.player_category,ra.category,ra.weight,
        rm.inline_use_same,rm.inline_type,rm.inline_url,rm.inline_barrage_switch FROM resource_assignment AS ra,resource_material AS rm 
		WHERE ra.resource_group_id>0 AND ra.category=0 AND ra.stime<? AND ra.etime>? AND ra.state=0 AND ra.audit_state IN (2,3,4) AND 
		ra.id=rm.resource_assignment_id AND rm.audit_state=2 AND rm.category=0 ORDER BY ra.position ASC,ra.weight DESC,rm.mtime DESC`
	_categoryAssignSQL = `SELECT ra.id,rm.id,rm.name,rm.subtitle,ra.contract_id,ra.resource_id,rm.pic,rm.pic_main_color,rm.litpic,rm.url,ra.rule,ra.position,ra.agency,ra.price,
		ra.stime,ra.etime,ra.apply_group_id,rm.ctime,rm.mtime,rm.atype,ra.username,rm.player_category,ra.category,rm.inline_use_same,rm.inline_type,rm.inline_url,rm.inline_barrage_switch
        FROM resource_assignment AS ra,resource_material AS rm 
		WHERE ra.id=rm.resource_assignment_id AND rm.id IN (SELECT max(rm.id) FROM resource_assignment AS ra,resource_material AS rm WHERE ra.resource_group_id>0 
		AND ra.category=1 AND ra.position_id NOT IN (%s) AND ra.stime<? AND ra.etime>? AND ra.state=0 AND ra.audit_state IN (2,3,4) AND ra.id=rm.resource_assignment_id AND 
		rm.audit_state=2 AND rm.category=1 GROUP BY rm.resource_assignment_id) ORDER BY rand()`
	_bossAssignSQL = `SELECT ra.id,rm.id,rm.name,rm.subtitle,ra.contract_id,ra.resource_id,rm.pic,rm.pic_main_color,rm.litpic,rm.url,ra.rule,ra.position,ra.agency,ra.price,
		ra.stime,ra.etime,ra.apply_group_id,rm.ctime,rm.mtime,rm.atype,ra.username,rm.player_category,ra.category,rm.inline_use_same,rm.inline_type,rm.inline_url,rm.inline_barrage_switch
        FROM resource_assignment AS ra,resource_material AS rm 
		WHERE ra.id=rm.resource_assignment_id AND rm.id IN (SELECT max(rm.id) FROM resource_assignment AS ra,resource_material AS rm WHERE ra.resource_group_id>0 
		AND ra.category=2 AND ra.stime<? AND ra.etime>? AND ra.state=0 AND ra.audit_state IN (2,3,4) AND ra.id=rm.resource_assignment_id AND 
		rm.audit_state=2 AND rm.category=2 GROUP BY rm.resource_assignment_id)`
	_defBannerSQL = `SELECT id,name,contract_id,resource_id,pic,litpic,url,rule,weight,agency,price,atype,username FROM default_one WHERE state=0`
	// index-icon
	_indexIconSQL = `SELECT id,type,title,state,link,icon,weight,user_name,sttime,endtime,deltime,ctime,mtime FROM icon WHERE state=1 AND deltime=0 AND (type=1 OR (type=2 AND sttime>0))`
	_playIconSQL  = `SELECT icon1,hash1,icon2,hash2,stime,relate_type,relate_value,mtime FROM bar_icon WHERE stime<? AND etime>? AND is_deleted=0`
	// cmtbox
	_cmtboxSQL = `SELECT id,load_cid,server,port,size_factor,speed_factor,max_onscreen,style,style_param,top_margin,state,renqi_visible,renqi_fontsize,renqi_fmt,renqi_offset,renqi_color,ctime,mtime FROM cmtbox WHERE state=1`
	// update resource assignment etime
	_updateResourceAssignmentEtime = `UPDATE resource_assignment SET etime=? WHERE id=?`
	// update resource apply status
	_updateResourceApplyStatus = `UPDATE resource_apply SET audit_state=? WHERE apply_group_id IN (%s)`
	// insert resource logs
	_inResourceLogger = `INSERT INTO resource_logger (uname,uid,module,oid,content) VALUES (?,?,?,?,?)`
	// custom config
	_customConfigsSQL = `SELECT tp,oid,content,url,highlight_content,image,image_big,stime,etime FROM custom_config where oid=? and state = 1 and  stime <= ? and etime >= ?`
)

// Resources get resource infos from db
func (d *Dao) Resources(c context.Context) (rscs []*model.Resource, err error) {
	var size sql.NullString
	rows, err := d.db.Query(c, _allResSQL)
	if err != nil {
		log.Error("d.Resources query error (%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		rsc := &model.Resource{}
		if err = rows.Scan(&rsc.ID, &rsc.Platform, &rsc.Name, &rsc.Parent, &rsc.Counter, &rsc.Position, &rsc.Rule, &size, &rsc.Previce,
			&rsc.Desc, &rsc.Mark, &rsc.CTime, &rsc.MTime, &rsc.Level, &rsc.Type, &rsc.IsAd); err != nil {
			log.Error("Resources rows.Scan err (%v)", err)
			return
		}
		rsc.Size = size.String
		rscs = append(rscs, rsc)
	}
	err = rows.Err()
	return
}

// Assignment get assigment from db
func (d *Dao) Assignment(c context.Context) (asgs []*model.Assignment, err error) {
	var (
		sqls []string
		args []interface{}
	)
	now := time.Now()
	args = append(args, now)
	args = append(args, now)
	if len(d.c.BannerID) == 0 {
		log.Error("日志告警 bannerID未配置")
		return
	}
	for _, val := range d.c.BannerID {
		sqls = append(sqls, "?")
		args = append(args, val)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_allAssignSQL, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("d.Assignment query error (%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		asg := &model.Assignment{}
		if err = rows.Scan(&asg.ID, &asg.Name, &asg.ContractID, &asg.ResID, &asg.Pic, &asg.PicMainColor, &asg.LitPic,
			&asg.URL, &asg.Rule, &asg.Weight, &asg.Agency, &asg.Price, &asg.Atype, &asg.Username, &asg.Inline.InlineUseSame,
			&asg.Inline.InlineType, &asg.Inline.InlineUrl, &asg.Inline.InlineBarrageSwitch); err != nil {
			log.Error("Assignment rows.Scan err (%v)", err)
			return
		}
		asg.AsgID = asg.ID
		asgs = append(asgs, asg)
	}
	err = rows.Err()
	return
}

// AssignmentNew get resource_assigment from new db
func (d *Dao) AssignmentNew(c context.Context) (asgs map[string][]*model.Assignment, err error) {
	rows, err := d.db.Query(c, _allAssignNewSQL, time.Now(), time.Now())
	if err != nil {
		log.Error("d.AssignmentNew query error (%v)", err)
		return
	}
	defer rows.Close()
	asgs = make(map[string][]*model.Assignment)
	asgIDm := make(map[int]struct{})
	for rows.Next() {
		asg := &model.Assignment{}
		if err = rows.Scan(&asg.AsgID, &asg.ID, &asg.Name, &asg.SubTitle, &asg.ContractID, &asg.ResID, &asg.Pic, &asg.PicMainColor, &asg.LitPic,
			&asg.URL, &asg.Rule, &asg.Weight, &asg.Agency, &asg.Price, &asg.STime, &asg.ETime, &asg.ApplyGroupID, &asg.CTime, &asg.MTime, &asg.Atype,
			&asg.Username, &asg.PlayerCategory, &asg.Category, &asg.PositionWeight, &asg.Inline.InlineUseSame, &asg.Inline.InlineType,
			&asg.Inline.InlineUrl, &asg.Inline.InlineBarrageSwitch); err != nil {
			log.Error("AssignmentNew rows.Scan err (%v)", err)
			return
		}
		if _, ok := asgIDm[asg.AsgID]; ok {
			continue
		}
		if d.InPosition(asg.ResID) {
			asg.ContractID = "rec_video"
		}
		pindex := fmt.Sprintf("%d_%d", asg.ResID, asg.Weight)
		asgs[pindex] = append(asgs[pindex], asg)
		asgIDm[asg.AsgID] = struct{}{}
	}
	err = rows.Err()
	return
}

// CategoryAssignment get recommend resource_assigment from db
func (d *Dao) CategoryAssignment(c context.Context) (asgs []*model.Assignment, err error) {
	var (
		sqls []string
		args []interface{}
	)
	if len(d.c.BannerID) == 0 {
		log.Error("日志告警 bannerID未配置")
		return
	}
	for _, val := range d.c.BannerID {
		sqls = append(sqls, "?")
		args = append(args, val)
	}
	now := time.Now()
	args = append(args, now)
	args = append(args, now)
	rows, err := d.db.Query(c, fmt.Sprintf(_categoryAssignSQL, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("d.CategoryAssignment query error (%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		asg := &model.Assignment{}
		if err = rows.Scan(&asg.AsgID, &asg.ID, &asg.Name, &asg.SubTitle, &asg.ContractID, &asg.ResID, &asg.Pic, &asg.LitPic, &asg.PicMainColor,
			&asg.URL, &asg.Rule, &asg.Weight, &asg.Agency, &asg.Price, &asg.STime, &asg.ETime, &asg.ApplyGroupID, &asg.CTime, &asg.MTime, &asg.Atype,
			&asg.Username, &asg.PlayerCategory, &asg.Category, &asg.Inline.InlineUseSame, &asg.Inline.InlineType, &asg.Inline.InlineUrl,
			&asg.Inline.InlineBarrageSwitch); err != nil {
			log.Error("CategoryAssignment rows.Scan err (%v)", err)
			return
		}
		if d.InResourceID(asg.ResID) {
			asg.ContractID = "rec_video"
		}
		asgs = append(asgs, asg)
	}
	err = rows.Err()
	return
}

// BossAssignment get boss resource_assigment from db
func (d *Dao) BossAssignment(c context.Context) (asgs []*model.Assignment, err error) {
	rows, err := d.db.Query(c, _bossAssignSQL, time.Now(), time.Now())
	if err != nil {
		log.Error("d.BossAssignment query error (%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		asg := &model.Assignment{}
		if err = rows.Scan(&asg.AsgID, &asg.ID, &asg.Name, &asg.SubTitle, &asg.ContractID, &asg.ResID, &asg.Pic, &asg.PicMainColor, &asg.LitPic,
			&asg.URL, &asg.Rule, &asg.Weight, &asg.Agency, &asg.Price, &asg.STime, &asg.ETime, &asg.ApplyGroupID, &asg.CTime,
			&asg.MTime, &asg.Atype, &asg.Username, &asg.PlayerCategory, &asg.Category, &asg.Inline.InlineUseSame, &asg.Inline.InlineType,
			&asg.Inline.InlineUrl, &asg.Inline.InlineBarrageSwitch); err != nil {
			log.Error("BossAssignment rows.Scan err (%v)", err)
			return
		}
		asgs = append(asgs, asg)
	}
	err = rows.Err()
	return
}

// DefaultBanner get default banner info
func (d *Dao) DefaultBanner(c context.Context) (asg *model.Assignment, err error) {
	row := d.db.QueryRow(c, _defBannerSQL)
	asg = &model.Assignment{}
	if err = row.Scan(&asg.ID, &asg.Name, &asg.ContractID, &asg.ResID, &asg.Pic, &asg.LitPic,
		&asg.URL, &asg.Rule, &asg.Weight, &asg.Agency, &asg.Price, &asg.Atype, &asg.Username); err != nil {
		if err == sql.ErrNoRows {
			asg = nil
			err = nil
		} else {
			log.Error("d.DefaultBanner.Scan error(%v)", err)
		}
	}
	return
}

// IndexIcon get index icon.
func (d *Dao) IndexIcon(c context.Context) (icons map[int][]*model.IndexIcon, err error) {
	rows, err := d.db.Query(c, _indexIconSQL)
	if err != nil {
		log.Error("d.IndexIcon query error (%v)", err)
		return
	}
	defer rows.Close()
	icons = make(map[int][]*model.IndexIcon)
	for rows.Next() {
		var link string
		icon := &model.IndexIcon{}
		if err = rows.Scan(&icon.ID, &icon.Type, &icon.Title, &icon.State, &link, &icon.Icon,
			&icon.Weight, &icon.UserName, &icon.StTime, &icon.EndTime, &icon.DelTime, &icon.CTime, &icon.MTime); err != nil {
			log.Error("IndexIcon rows.Scan err (%v)", err)
			return
		}
		icon.Links = strings.Split(link, ",")
		icons[icon.Type] = append(icons[icon.Type], icon)
	}
	err = rows.Err()
	return
}

// PlayerIcon get play icon
func (d *Dao) PlayerIcon(c context.Context) (res []*model.PlayerIcon, err error) {
	rows, err := d.db.Query(c, _playIconSQL, time.Now(), time.Now())
	if err != nil {
		log.Error("d.PlayerIcon query error (%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &model.PlayerIcon{}
		if err = rows.Scan(&re.URL1, &re.Hash1, &re.URL2, &re.Hash2, &re.CTime, &re.Type, &re.TypeValue, &re.MTime); err != nil {
			log.Error("IndexIcon rows.Scan err (%v)", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// Cmtbox sql live danmaku box
func (d *Dao) Cmtbox(c context.Context) (res map[int64]*model.Cmtbox, err error) {
	rows, err := d.db.Query(c, _cmtboxSQL)
	if err != nil {
		log.Error("d.db.Query error (%v)", err)
		return
	}
	defer rows.Close()
	res = make(map[int64]*model.Cmtbox)
	for rows.Next() {
		re := &model.Cmtbox{}
		if err = rows.Scan(&re.ID, &re.LoadCID, &re.Server, &re.Port, &re.SizeFactor, &re.SpeedFactor, &re.MaxOnscreen,
			&re.Style, &re.StyleParam, &re.TopMargin, &re.State, &re.RenqiVisible, &re.RenqiFontsize, &re.RenqiFmt, &re.RenqiOffset, &re.RenqiColor, &re.CTime, &re.MTime); err != nil {
			log.Error("Cmtbox rows.Scan err (%v)", err)
			return
		}
		res[re.ID] = re
	}
	err = rows.Err()
	return
}

// TxOffLine off line resource
func (d *Dao) TxOffLine(tx *xsql.Tx, id int) (row int64, err error) {
	res, err := tx.Exec(_updateResourceAssignmentEtime, time.Now(), id)
	if err != nil {
		log.Error("TxOffLine tx.Exec() error(%v)", err)
		return
	}
	row, err = res.RowsAffected()
	return
}

// TxFreeApply free apply
func (d *Dao) TxFreeApply(tx *xsql.Tx, ids []string) (row int64, err error) {
	var (
		sqls []string
		args []interface{}
	)
	args = append(args, model.ApplyNoAssignment)
	for _, id := range ids {
		sqls = append(sqls, "?")
		args = append(args, id)
	}
	res, err := tx.Exec(fmt.Sprintf(_updateResourceApplyStatus, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("TxFreeApply tx.Exec() error(%v)", err)
		return
	}
	row, err = res.RowsAffected()
	return
}

// TxInResourceLogger add resource log
func (d *Dao) TxInResourceLogger(tx *xsql.Tx, module, content string, oid int) (row int64, err error) {
	res, err := tx.Exec(_inResourceLogger, "rejob", 1203, module, oid, content)
	if err != nil {
		log.Error("TxInResourceLogger tx.Exec() error(%v)", err)
		return
	}
	row, err = res.RowsAffected()
	return
}

// CustomConfigs is
func (d *Dao) GetCustomConfigByIdFromDB(ctx context.Context, id int64) (cc *model.CustomConfig, err error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	row := d.db.QueryRow(ctx, _customConfigsSQL, id, now, now)

	if row != nil {
		cc = new(model.CustomConfig)
		if err = row.Scan(&cc.TP, &cc.Oid, &cc.Content, &cc.URL, &cc.HighlightContent, &cc.Image, &cc.ImageBig, &cc.STime, &cc.ETime); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			log.Error("GetCustomConfigByIdFromDB fail 0: %s", err.Error())
			return nil, err
		}
		go func() {
			conn := d.showRedis.Conn(ctx)
			defer func() {
				conn.Close()
				if r := recover(); r != nil {
					log.Error("GetCustomConfigByIdFromDB error (%v)", r)
				}
			}()
			ccJson, err := json.Marshal(cc)
			if err != nil {
				log.Error("GetCustomConfigByIdFromDB fail 1 : %+v, data: %+v", err, ccJson)
			}
			key := fmt.Sprintf("cc_%d", cc.Oid)

			if _, err = redis.String(conn.Do("SET", key, ccJson)); err != nil {
				log.Error("GetCustomConfigByIdFromDB fail 2: %+v, data: %+v", err, ccJson)
			}
			if _, err = redis.String(conn.Do("EXPIREAT", key, cc.ETime.Unix())); err != nil {
				log.Error("GetCustomConfigByIdFromDB fail 3: %+v, data: %+v", err, ccJson)
			}
		}()
	}
	return cc, nil
}

// CustomConfigs is
func (d *Dao) GetCustomConfigByIdFromRedis(ctx context.Context, id int64) (cc *model.CustomConfig, err error) {
	conn := d.showRedis.Conn(ctx)
	defer conn.Close()

	key := fmt.Sprintf("cc_%d", id)

	var ccJson string

	if ccJson, err = redis.String(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Error("GetCustomConfigByIdFromRedis fail 0: %+v, data: %+v", err, ccJson)
		return nil, err
	}

	cc = new(model.CustomConfig)
	if err = json.Unmarshal([]byte(ccJson), cc); err != nil {
		log.Error("GetCustomConfigByIdFromRedis fail 1: %+v, data: %+v", err, ccJson)
		return nil, err
	}

	return cc, nil
}

func (d *Dao) GetCustomConfigBySF(ctx context.Context, tp int32, id int64) (cc *model.CustomConfig, err error) {
	key := fmt.Sprintf("%d-%d", tp, id)

	val, err, _ := d.singleGetCC.Do(key, func() (ret interface{}, err error) {
		if cc, err := d.GetCustomConfigByIdFromRedis(ctx, id); err != nil {
			return nil, err
		} else {
			if cc != nil {
				return cc, nil
			}
			if cc, err = d.GetCustomConfigByIdFromDB(ctx, id); err != nil {
				return nil, err
			}
			return cc, nil
		}
	})
	if err != nil {
		return nil, err
	}
	return val.(*model.CustomConfig), nil
}

func (d *Dao) InResourceID(id int) (ok bool) {
	if d.c.ResourceLabel == nil {
		return
	}
	for _, resid := range d.c.ResourceLabel.ResourceIDs {
		if id == resid {
			ok = true
			break
		}
	}
	return
}

func (d *Dao) InPosition(id int) (ok bool) {
	if d.c.ResourceLabel == nil {
		return
	}
	for _, resid := range d.c.ResourceLabel.PositionIDs {
		if id == resid {
			ok = true
			break
		}
	}
	return
}
