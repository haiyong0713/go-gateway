package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/database/sql"
	"go-common/library/log"
	model "go-gateway/app/app-svr/app-feed/admin/model/frontpage"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	"strings"
	"time"
)

const (
	dataFormat       = "2006-01-02 15:04:00"
	_queryDefaultSQL = `select id, name, litpic, pic, url,is_split_layer, ifnull(split_layer,"[]") from default_one where id = 1`
	_queryOnlineSQL  = `select id, name, litpic, pic, url, stime, etime, state,is_split_layer, ifnull(split_layer,"[]"), rule, resource_id 
                       from resource_assignment where resource_id = ? AND stime <= ? AND etime >= ? AND state = 0 order by etime desc`
	_queryHiddenSQL = `select id, name, litpic, pic, url, stime, etime, state,is_split_layer, ifnull(split_layer,"[]"), rule, resource_id
                       from resource_assignment where resource_id = ? AND stime > ? AND state = 0`
	_queryAllMenuResourceIDsSQL = "SELECT id FROM resource WHERE mark in (%s)"
)

func (d *Dao) RawDefaultPage(c context.Context, req *pb.FrontPageReq) (ret *pb.FrontPage, err error) {
	ret = &pb.FrontPage{}
	if err = d.db.QueryRow(c, _queryDefaultSQL).Scan(&ret.Id, &ret.Title, &ret.Logo, &ret.Litpic, &ret.JumpUrl, &ret.IsSplitLayer, &ret.SplitLayer); err != nil {
		log.Error("dao.GetDefault query(%+v) error(%+v)", _queryDefaultSQL, err)
		return
	}
	return
}

func (d *Dao) RawOnlinePage(c context.Context, req *pb.FrontPageReq) (ret []*pb.FrontPage, err error) {
	if ret, err = d.GetFrontPage(c, _queryOnlineSQL, req); err != nil {
		log.Error("dao.RawOnlinePage error(%+v)", err)
	}
	return
}

func (d *Dao) RawHiddenPage(c context.Context, req *pb.FrontPageReq) (ret []*pb.FrontPage, err error) {
	if ret, err = d.GetFrontPage(c, _queryHiddenSQL, req); err != nil {
		log.Error("dao.RawHiddenPage error(%+v)", err)
	}
	return
}

func (d *Dao) GetFrontPage(c context.Context, querySql string, req *pb.FrontPageReq) (ret []*pb.FrontPage, err error) {
	now := time.Now().Format(dataFormat)
	var rows *sql.Rows
	if querySql == _queryOnlineSQL {
		if rows, err = d.db.Query(c, querySql, req.ResourceId, now, now); err != nil {
			log.Error("dao.GetOnline query(%+v) error(%+v)", querySql, err)
			return
		}
	} else {
		if rows, err = d.db.Query(c, querySql, req.ResourceId, now); err != nil {
			log.Error("dao.GetOnline query(%+v) error(%+v)", querySql, err)
			return
		}
	}
	defer rows.Close()
	ret = make([]*pb.FrontPage, 0)
	indx := 1
	for rows.Next() {
		item := &pb.FrontPage{}
		var rule string
		if err = rows.Scan(&item.Id, &item.Title, &item.Logo, &item.Litpic, &item.JumpUrl, &item.Stime, &item.Etime,
			&item.State, &item.IsSplitLayer, &item.SplitLayer, &rule, &item.ResourceId); err != nil {
			log.Error("dao.GetOnline Scan row error(%+v)", err)
			return
		}
		tmpStyle := &struct {
			IsCover int   `json:"is_cover"`
			Style   int32 `json:"style"`
		}{}
		if err = json.Unmarshal([]byte(rule), &tmpStyle); err != nil {
			log.Error("dao.GetFrontPage Unmarshal rule(%+v) error(%+v)", rule, err)
			return
		}
		item.Style = tmpStyle.Style
		item.Pos = int32(indx)
		indx += 1
		ret = append(ret, item)
	}
	if err = rows.Err(); err != nil {
		log.Error("GetOnline rows error: %s", err)
	}
	return
}

func (d *Dao) GetEffectiveFrontPage(c context.Context, req *pb.FrontPageReq) (ret *pb.FrontPageResp, err error) {
	ret = &pb.FrontPageResp{}
	if ret.Default, err = d.DefaultPage(c, req); err != nil {
		log.Error("dao.GetEffectiveFrontPage get defaultPage req(%+v) error(%+v)", req, err)
		return
	}
	if ret.Online, err = d.OnlinePage(c, req); err != nil {
		log.Error("dao.GetEffectiveFrontPage get onlinePage req(%+v) error(%+v)", req, err)
		return
	}
	if ret.Hidden, err = d.HiddenPage(c, req); err != nil {
		log.Error("dao.GetEffectiveFrontPage get hiddenPage req(%+v) error(%+v)", req, err)
		return
	}
	return
}

// 重构后
func (d *Dao) GetAllMenuResourceIDs(ctx context.Context) (res []int64, err error) {
	allMarks := d.GetCategoriesKeys()

	allMarksVals := make([]string, 0, len(allMarks))
	for _, mark := range allMarks {
		allMarksVals = append(allMarksVals, "'"+mark+"'")
	}

	res = make([]int64, 0)
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, fmt.Sprintf(_queryAllMenuResourceIDsSQL, strings.Join(allMarksVals, ","))); err != nil {
		res = nil
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			if err == sql.ErrNoRows {
				err = nil
				break
			}
			return nil, err
		}
		res = append(res, id)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	res = append(res, 0) // 添加默认区
	return
}

// GetCategoriesKeys
func (d *Dao) GetCategoriesKeys() (res []string) {
	res = make([]string, 0, len(model.CategoriesMap))
	for key := range model.CategoriesMap {
		res = append(res, key)
	}
	return
}
