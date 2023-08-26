package dao

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"time"

	xsql "go-common/library/database/sql"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"go-gateway/app/web-svr/space/interface/model"

	"github.com/pkg/errors"
)

const (
	_topPhotoWeb          = "SELECT mid,sid,platfrom,expire FROM member_topphoto%d WHERE mid = ? AND is_activated = 1 AND expire > ? AND (platfrom = 0 OR platfrom = 1)"
	_topPhotoUploadByID   = "SELECT img_path FROM member_upload_topphoto WHERE id = ?"
	_topPhotoUpload       = "SELECT id,mid,platfrom,img_path,status FROM member_upload_topphoto WHERE mid = ? AND deleted = 0 AND status in (0,1) ORDER BY upload_date DESC"
	_deletePhotoUploadSQL = `UPDATE member_upload_topphoto SET deleted = 1 WHERE mid = ? AND (platfrom = ? OR platfrom = ?)`
	_insertPhotoUploadSQL = "INSERT INTO member_upload_topphoto (img_path,mid,platfrom) VALUES (?,?,?)"
	_topPhotoActiveDel    = "UPDATE member_topphoto%d SET is_activated = 0 WHERE mid = ?"
	_topPhotoCount        = "SELECT count(*) as count from member_topphoto%d WHERE mid = ? AND sid = ?"
	_topPhotoActiveYes    = "UPDATE member_topphoto%d SET is_activated = 1 WHERE mid = ? AND sid = ?"
	_topPhotoInsert       = "INSERT INTO member_topphoto%d (mid,sid,is_activated,expire) VALUES (?,?,1,?)"
	_updateTopPhotoExpire = "UPDATE member_topphoto%d SET expire = ? WHERE id = ?"
	_allMemberTopPhoto    = "SELECT id,mid,sid,expire,platfrom FROM member_topphoto%d WHERE mid = ?"
)

const (
	_webTopPhotoURI = "/api/member/getTopPhoto"
	_topPhotoURI    = "/api/member/getUploadTopPhoto"
	_setTopPhotoURI = "/api/member/setTopPhoto"
	//
	_memTopPhotoHit = 10
)

// WebTopPhoto getTopPhoto from space
func (d *Dao) WebTopPhoto(c context.Context, mid, loginMid int64, mobiapp, device string) (space *model.TopPhoto, err error) {
	var (
		params   = url.Values{}
		remoteIP = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("mid", strconv.FormatInt(mid, 10))
	if mobiapp != "" {
		params.Set("mobiapp", mobiapp)
	}
	if device != "" {
		params.Set("device", device)
	}
	if loginMid > 0 {
		params.Set("login_mid", strconv.FormatInt(loginMid, 10))
	}
	var res struct {
		Code int `json:"code"`
		model.TopPhoto
	}
	if err = d.httpR.Get(c, d.webTopPhotoURL, remoteIP, params, &res); err != nil {
		log.Error("d.httpR.Get(%s,%d) error(%v)", d.webTopPhotoURL, mid, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("d.httpR.Get(%s,%d) code error(%d)", d.webTopPhotoURL, mid, res.Code)
		err = ecode.Int(res.Code)
		return
	}
	space = &res.TopPhoto
	return
}

// TopPhoto getTopPhoto from space php.
func (d *Dao) TopPhoto(c context.Context, mid, vmid int64, platform, device string) (imgURL string, err error) {
	var (
		params   = url.Values{}
		remoteIP = metadata.String(c, metadata.RemoteIP)
	)
	if mid > 0 {
		params.Set("mid", strconv.FormatInt(mid, 10))
	}
	params.Set("vmid", strconv.FormatInt(vmid, 10))
	params.Set("platform", platform)
	if device != "" {
		params.Set("device", device)
	}
	var res struct {
		Code int `json:"code"`
		Data struct {
			ImgURL string `json:"imgUrl"`
		}
	}
	if err = d.httpR.Get(c, d.topPhotoURL, remoteIP, params, &res); err != nil {
		log.Error("d.httpR.Get(%s,%d) error(%v)", d.topPhotoURL, mid, err)
		return
	}

	if res.Code != ecode.OK.Code() {
		log.Error("d.httpR.Get(%s,%d) code error(%d)", d.topPhotoURL, mid, res.Code)
		err = ecode.Int(res.Code)
		return
	}
	imgURL = res.Data.ImgURL
	return
}

func (d *Dao) SetTopPhoto(c context.Context, mid, id int64, mobiapp string) (err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("id", strconv.FormatInt(id, 10))
	params.Set("mobiapp", mobiapp)
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err = d.httpW.Post(c, d.setTopPhotoURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		err = errors.Wrapf(err, "SetTopPhoto d.httpW.Post(%s)", d.setTopPhotoURL+"?"+params.Encode())
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(res.Code), "SetTopPhoto d.httpW.Post(%s,%d)", d.setTopPhotoURL+"?"+params.Encode(), res.Code)
	}
	return
}

const _topPhotoArcSQL = "SELECT aid,image_url,mid,ctime,mtime FROM topphoto_arc_%d WHERE mid=?"

func (d *Dao) RawTopPhotoArc(ctx context.Context, mid int64) (*model.TopPhotoArc, error) {
	row := d.db.QueryRow(ctx, fmt.Sprintf(_topPhotoArcSQL, mid%10), mid)
	data := &model.TopPhotoArc{}
	if err := row.Scan(&data.Aid, &data.ImageUrl, &data.Mid, &data.Ctime, &data.Mtime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "RawTopPhotoArc")
	}
	return data, nil
}

const _topPhotoArcAddSQL = "INSERT INTO topphoto_arc_%d(mid,aid,image_url) VALUES (?,?,?) ON DUPLICATE KEY UPDATE aid=?,image_url=?"

func (d *Dao) AddTopPhotoArc(ctx context.Context, mid, aid int64, imageURL string) (int64, error) {
	res, err := d.db.Exec(ctx, fmt.Sprintf(_topPhotoArcAddSQL, mid%10), mid, aid, imageURL, aid, imageURL)
	if err != nil {
		return 0, errors.Wrap(err, "AddTopPhotoArc")
	}
	return res.LastInsertId()
}

const _topPhotoArcCancelSQL = `UPDATE topphoto_arc_%d SET aid=0,image_url="" WHERE mid=?`

func (d *Dao) TopPhotoArcCancel(ctx context.Context, mid int64) (int64, error) {
	res, err := d.db.Exec(ctx, fmt.Sprintf(_topPhotoArcCancelSQL, mid%10), mid)
	if err != nil {
		return 0, errors.Wrap(err, "TopPhotoArcCancel")
	}
	return res.RowsAffected()
}

func keyTopPhotoArc(mid int64) string {
	return fmt.Sprintf("%d_topphoto_arc", mid)
}

func (d *Dao) cacheSFTopPhotoArc(mid int64) string {
	return keyTopPhotoArc(mid)
}

// memTopphotoHit .
func memTopphotoHit(mid int64) int64 {
	return mid % _memTopPhotoHit
}

func (d *Dao) CacheTopPhotoArc(ctx context.Context, mid int64) (*model.TopPhotoArc, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	key := keyTopPhotoArc(mid)
	bs, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "CacheTopPhotoArc key:%s", key)
	}
	res := &model.TopPhotoArc{}
	if err = res.Unmarshal(bs); err != nil {
		return nil, errors.Wrap(err, "CacheTopPhotoArc Unmarshal")
	}
	return res, nil
}

func (d *Dao) AddCacheTopPhotoArc(ctx context.Context, mid int64, data *model.TopPhotoArc) error {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	key := keyTopPhotoArc(mid)
	bs, err := data.Marshal()
	if err != nil {
		return errors.Wrap(err, "AddCacheUserTab Marshal")
	}
	if _, err = conn.Do("SETEX", key, d.redisTopPhotoArcExpire, bs); err != nil {
		return errors.Wrap(err, "AddCacheTopPhotoArc SETEX")
	}
	return nil
}

func (d *Dao) DelCacheTopPhotoArc(ctx context.Context, mid int64) (err error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	key := keyTopPhotoArc(mid)
	if _, err = conn.Do("DEL", key); err != nil {
		if err == redis.ErrNil {
			return nil
		}
		log.Errorc(ctx, "d.DelCacheTopPhotoArc(key: %v) err: %+v", key, err)
		return err
	}
	return nil
}

// MemSetTopPhoto .
func (d *Dao) MemSetTopPhoto(c context.Context, mid, sid, expire int64) (err error) {
	var (
		tx    *xsql.Tx
		count int64
	)
	if tx, err = d.db.Begin(c); err != nil {
		log.Error("MemSetTopPhoto: d.db.Begin error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			log.Error("MemSetTopPhoto recover mid(%d) sid(%d) recover(%v)", mid, sid, r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("MemSetTopPhoto Rollback mid(%d) sid(%d) error(%v),error1(%v)", mid, sid, err, err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("MemSetTopPhoto Commit mid(%d) sid(%d) error(%v)", mid, sid, err)
			return
		}
	}()
	if _, err = tx.Exec(fmt.Sprintf(_topPhotoActiveDel, memTopphotoHit(mid)), mid); err != nil {
		return
	}
	row := tx.QueryRow(fmt.Sprintf(_topPhotoCount, memTopphotoHit(mid)), mid, sid)
	if err = row.Scan(&count); err != nil {
		return
	}
	if count > 0 {
		//已存在 直接更新
		if _, err = tx.Exec(fmt.Sprintf(_topPhotoActiveYes, memTopphotoHit(mid)), mid, sid); err != nil {
			return
		}
	} else {
		if _, err = tx.Exec(fmt.Sprintf(_topPhotoInsert, memTopphotoHit(mid)), mid, sid, expire); err != nil {
			return
		}
	}
	//删除上传的头图
	if _, err = tx.Exec(_deletePhotoUploadSQL, mid, model.UploadTopPhotoWeb, model.UploadTopPhotoIpad); err != nil {
		return
	}
	//删除德德皮肤主题
	if _, err = tx.Exec(fmt.Sprintf(_themeUnSQL, themeHit(mid)), mid); err != nil {
		return
	}
	return
}

// RawMemberUploadTopphoto .
func (d *Dao) RawMemberUploadTopphoto(c context.Context, mid int64, platFrom int) (*model.MemberPhotoUpload, error) {
	var (
		rows   *xsql.Rows
		photos map[int][]*model.MemberPhotoUpload
	)
	defer func() {
		log.Info("RawMemberUploadTopphoto mid(%d) platFrom(%d) res(%+v)", mid, platFrom, photos)
	}()
	rows, err := d.db.Query(c, _topPhotoUpload, mid)
	if err != nil {
		log.Error("RawMemberUploadTopphoto mid(%d) platFrom(%d) db.Exec(%s) error(%v)", mid, platFrom, _topPhotoUpload, err)
		return nil, err
	}
	defer rows.Close()
	photos = make(map[int][]*model.MemberPhotoUpload)
	for rows.Next() {
		photo := new(model.MemberPhotoUpload)
		err = rows.Scan(&photo.ID, &photo.Mid, &photo.Platfrom, &photo.ImgPath, &photo.Status)
		if err != nil {
			err = nil
			log.Error("RawMemberUploadTopphoto  mid(%d) platFrom(%d) row.Scan error(%v)", mid, platFrom, err)
			continue
		}
		//按照平台分类
		photos[photo.Platfrom] = append(photos[photo.Platfrom], photo)
	}
	if err = rows.Err(); err != nil {
		log.Error("RawMemberUploadTopphoto mid(%d) platFrom(%d) error(%v)", mid, platFrom, err)
		return nil, err
	}
	if len(photos) == 0 {
		return nil, nil
	}
	datas := make(map[int]map[int]*model.MemberPhotoUpload)
	for plat, photo := range photos {
		//按照平台+审核状态分类
		tmp := make(map[int]*model.MemberPhotoUpload)
		for _, v := range photo {
			tmp[v.Status] = v
		}
		datas[plat] = tmp
	}
	//ios 和 android
	if platFrom == model.UploadTopPhotoIos || platFrom == model.UploadTopPhotoAndroid {
		//按照android,ios顺序 优先待审核
		v, ok := datas[model.UploadTopPhotoAndroid][model.UploadTopPhotoVerify]
		if ok {
			return v, nil
		}
		v, ok = datas[model.UploadTopPhotoIos][model.UploadTopPhotoVerify]
		if ok {
			return v, nil
		}
		v, ok = datas[model.UploadTopPhotoAndroid][model.UploadTopPhotoPass]
		if ok {
			return v, nil
		}
		v, ok = datas[model.UploadTopPhotoIos][model.UploadTopPhotoPass]
		if ok {
			return v, nil
		}
		return nil, nil
	}
	//按照web,ipad顺序 优先待审核
	v, ok := datas[model.UploadTopPhotoWeb][model.UploadTopPhotoVerify]
	if ok {
		return v, nil
	}
	v, ok = datas[model.UploadTopPhotoIpad][model.UploadTopPhotoVerify]
	if ok {
		return v, nil
	}
	v, ok = datas[model.UploadTopPhotoWeb][model.UploadTopPhotoPass]
	if ok {
		return v, nil
	}
	v, ok = datas[model.UploadTopPhotoIpad][model.UploadTopPhotoPass]
	if ok {
		return v, nil
	}
	return nil, nil
}

// webTopphoto  .
func (d *Dao) RawMemberTopphoto(c context.Context, mid int64) (res *model.MemberTopphoto, err error) {
	res = &model.MemberTopphoto{}
	topPhotoSQL := fmt.Sprintf(_topPhotoWeb, memTopphotoHit(mid))
	row := d.db.QueryRow(c, topPhotoSQL, mid, time.Now().Unix())
	if err = row.Scan(&res.Mid, &res.Sid, &res.Platfrom, &res.Expire); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			log.Error("MemPhotoMall:row.Scan Mid(%d) error(%v)", mid, err)
		}
	}
	return
}

// RawMemberUploadTopphotoByID .
func (d *Dao) RawMemberUploadTopphotoByID(c context.Context, id int64) (res *model.MemberPhotoUpload, err error) {
	res = &model.MemberPhotoUpload{}
	row := d.db.QueryRow(c, _topPhotoUploadByID, id)
	if err = row.Scan(&res.ImgPath); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("RawMemberUploadTopphotoByID:row.Scan error(%v)", err)
		}
	}
	return
}

// AddTopphotoUpload .
func (d *Dao) AddTopphotoUpload(c context.Context, mid int64, platfrom int, path string) (id int64, err error) {
	var (
		tx             *xsql.Tx
		res            sql.Result
		delOne, delTwo int
	)
	if tx, err = d.db.Begin(c); err != nil {
		log.Error("AddTopphotoUpload: d.db.Begin error(%v)", err)
		return
	}
	if platfrom == model.UploadTopPhotoIos || platfrom == model.UploadTopPhotoAndroid {
		delOne = model.UploadTopPhotoIos
		delTwo = model.UploadTopPhotoAndroid
	} else {
		delOne = model.UploadTopPhotoWeb
		delTwo = model.UploadTopPhotoIpad
	}
	if _, err = tx.Exec(_deletePhotoUploadSQL, mid, delOne, delTwo); err != nil {
		_ = tx.Rollback()
		log.Error("AddTopphotoUpload: tx.Exec(sql:%s mid:%d,platfrom:%d) error(%v)", _deletePhotoUploadSQL, mid, platfrom, err)
		return
	}
	if res, err = tx.Exec(_insertPhotoUploadSQL, path, mid, platfrom); err != nil {
		_ = tx.Rollback()
		log.Error("AddTopphotoUpload: tx.Exec(sql:%s mid:%d,platfrom:%d) error(%v)", _insertPhotoUploadSQL, mid, platfrom, err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("AddTopphotoUpload: tx.Commit error(%v)", err)
		return
	}
	id, err = res.LastInsertId()
	return
}

// UpdateMemberTopPhotoExpire
func (d *Dao) UpdateMemberTopPhotoExpire(c context.Context, id, mid, expire int64) error {
	updateSQL := fmt.Sprintf(_updateTopPhotoExpire, memTopphotoHit(mid))
	if _, err := d.db.Exec(c, updateSQL, expire, id); err != nil {
		return errors.Wrapf(ecode.ServerErr, "%+v", err)
	}
	return nil
}

// UpdateMemberTopPhotoExpire
func (d *Dao) GetMemberTopPhoto(c context.Context, mid int64) ([]*model.MemberTopphoto, error) {
	selectSQL := fmt.Sprintf(_allMemberTopPhoto, memTopphotoHit(mid))
	rows, err := d.db.Query(c, selectSQL, mid)
	if err != nil {
		return nil, errors.Wrapf(ecode.ServerErr, "%+v", err)
	}
	res := []*model.MemberTopphoto{}
	defer rows.Close()
	for rows.Next() {
		topPhoto := &model.MemberTopphoto{}
		if err := rows.Scan(&topPhoto.ID, &topPhoto.Mid, &topPhoto.Sid, &topPhoto.Expire, &topPhoto.Platfrom); err != nil {
			return nil, errors.Wrapf(ecode.ServerErr, "%+v", err)
		}
		res = append(res, topPhoto)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrapf(ecode.ServerErr, "%+v", err)
	}
	return res, nil
}
