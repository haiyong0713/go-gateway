package like

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	actecode "go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/job/model/like"
	"net/url"
	"strconv"
)

const (
	_upActReserveRelationBindInfo = "select `id`, `sid`, `rid` from up_act_reserve_relation_bind where oid = %s and o_type = %d and r_type = %d order by sid desc limit 1"
	_upActReserveRelationBind     = "insert into up_act_reserve_relation_bind (`sid`,`oid`,`o_type`,`rid`,`r_type`) values (?,?,?,?,?)"
	_upActReserveUpdateDynamicId  = "update up_act_reserve_relation set dynamic_id = ? where sid = ?"
)

func (d *Dao) GetUpActReserveRelationBindInfo(ctx context.Context, oid string, oType int64, from int64) (res *like.UpActReserveRelationBind, err error) {
	row := d.db.QueryRow(ctx, fmt.Sprintf(_upActReserveRelationBindInfo, oid, oType, from))
	res = new(like.UpActReserveRelationBind)
	if err = row.Scan(&res.ID, &res.Sid, &res.Rid); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = fmt.Errorf("GetUpActReserveRelationBindInfo d.db.Query _upActReserveRelationByOid rows.Scan error(%+v)", err)
		}
	}

	return
}

func (d *Dao) BuildDynamicData(ctx context.Context, relation *like.UpActReserveRelation, cover string, fileInfo *like.BFSFileInfo) (res *like.CreateDynamicCard) {
	res = new(like.CreateDynamicCard)

	// 组装动态数据
	createDynamicCardExtension := &like.CreateDynamicCardExtension{}
	createDynamicCardExtension.FlagCfg.Reserve.ReserveID = relation.Sid
	extend, _ := json.Marshal(createDynamicCardExtension)

	createDynamicCardImgs := make([]like.CreateDynamicCardImg, 0)
	createDynamicCardImg := like.CreateDynamicCardImg{
		ImgSrc:    cover,
		ImgWidth:  strconv.FormatInt(fileInfo.Width, 10),
		ImgHeight: strconv.FormatInt(fileInfo.Height, 10),
		ImgSize:   strconv.FormatInt(fileInfo.FileSize/1024, 10),
	}
	createDynamicCardImgs = append(createDynamicCardImgs, createDynamicCardImg)

	res.UID = relation.Mid
	res.Biz = like.CreateDynamicBiz
	res.Category = like.CreateDynamicCategory
	res.Type = like.CreateDynamicType
	res.Pictures = createDynamicCardImgs
	res.Description = "视频更新预告"
	res.From = like.CreateDynamicFrom
	res.Extension = string(extend)
	res.AuditLevel = like.CreateDynamicAuditLevel

	return
}

func (d *Dao) CreateDynamicData(ctx context.Context, data *like.CreateDynamicCard) (dynamicID string, err error) {
	params := url.Values{}
	pics, _ := json.Marshal(data.Pictures)

	params.Set("uid", strconv.FormatInt(data.UID, 10))
	params.Set("biz", strconv.FormatInt(data.Biz, 10))
	params.Set("category", strconv.FormatInt(data.Category, 10))
	params.Set("type", strconv.FormatInt(data.Type, 10))
	params.Set("pictures", string(pics))
	params.Set("description", data.Description)
	params.Set("from", data.From)
	params.Set("extension", data.Extension)
	params.Set("audit_level", strconv.FormatInt(data.AuditLevel, 10))

	reply := new(like.CreateDynamicReply)
	err = d.httpClient.Post(ctx, d.createDynamicURL, "", params, reply)
	log.Infoc(ctx, "CreateDynamicData url(%+v) params(%+v) reply(%+v) err(%+v)", d.createDynamicURL, params, reply, err)
	// 忽略创建动态失败的错误码
	if err != nil || reply.Code != 0 {
		err = nil
		return
	}

	if reply.Data.DynamicIDStr == "" {
		err = fmt.Errorf("CreateDynamicData d.httpClient.Post reply.Data.DynamicIDStr == ''")
		return
	}

	dynamicID = reply.Data.DynamicIDStr
	return
}

func (d *Dao) DeleteDynamicRelatedByLiveReserve(ctx context.Context, mid int64, dynamicID string) (err error) {
	params := url.Values{}

	params.Set("dynamic_id", dynamicID)
	params.Set("dynamic_uid", strconv.FormatInt(mid, 10))
	params.Set("from", like.DeleteDynamicFrom)

	var res struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	err = d.httpClient.Post(ctx, d.deleteDynamicURL, "", params, &res)
	log.Infoc(ctx, "DeleteDynamicRelatedByLiveReserve d.httpClient.Post req:(%+v), reply:(%+v), err:(%+v)", params, res, err)
	if err != nil {
		err = errors.Wrapf(err, "DeleteDynamicRelatedByLiveReserve d.httpClient.Post params(%+v) reply(%+v) error(%+v)", params, res, err)
		return
	}

	if res.Code != ecode.OK.Code() && res.Code != actecode.DynamicErrDynamicRemoved.Code() {
		err = errors.Wrapf(err, "DeleteDynamicRelatedByLiveReserve d.httpClient.Post res.Code != 0")
		return
	}
	return
}

func (d *Dao) TXCreateUpActReserveRelationBind(tx *sql.Tx, sid int64, oid string, oType int64, rid string, rType int64) (err error) {
	res, err := tx.Exec(_upActReserveRelationBind, sid, oid, oType, rid, rType)
	if err != nil {
		err = errors.Wrap(err, "CreateUpActReserveRelationBind _upActReserveRelationBind Exec")
		return
	}
	if eff, _ := res.RowsAffected(); eff <= 0 {
		err = errors.Wrap(err, "CreateUpActReserveRelationBind _upActReserveRelationBind RowsAffected")
		return
	}
	return
}

func (d *Dao) TXUpdateUpActReserveRelationDynamicId(tx *sql.Tx, sid int64, dynamicId string) (err error) {
	res, err := tx.Exec(_upActReserveUpdateDynamicId, dynamicId, sid)
	if err != nil {
		err = errors.Wrap(err, "UpdateUpActReserveRelationDynamicId _upActReserveUpdateDynamicId Exec err")
		return
	}
	if eff, _ := res.RowsAffected(); eff <= 0 {
		err = errors.Wrap(err, "UpdateUpActReserveRelationDynamicId _upActReserveUpdateDynamicId RowsAffected 0")
		return
	}
	return
}

// 事务修改两张表数据
func (d *Dao) TXUpdateSubjectAndRelationData(ctx context.Context, sid int64, oid string, oType int64, rid string, rType int64) (err error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		err = errors.Wrap(err, "d.db.Begin err")
		return
	}

	defer func() {
		if r := recover(); r != nil {
			if err = tx.Rollback(); err != nil {
				err = errors.Wrap(err, "tx.Rollback() err")
				return
			}
			return
		}
		if err != nil {
			if err = tx.Rollback(); err != nil {
				err = errors.Wrap(err, "tx.Rollback() err")
				return
			}
			return
		}
		if err = tx.Commit(); err != nil {
			err = errors.Wrap(err, "tx.Commit() err")
			return
		}
		return
	}()

	// 更新relation表的动态id字段
	if err = d.TXUpdateUpActReserveRelationDynamicId(
		tx, sid, rid); err != nil {
		return
	}
	// 添加bind表记录
	if err = d.TXCreateUpActReserveRelationBind(
		tx, sid, oid, oType, rid, rType); err != nil {
		return
	}

	return
}
