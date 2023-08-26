package manager

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"go-common/library/database/sql"
	"go-common/library/ecode"
	pb2 "go-gateway/app/app-svr/resource/service/api/v2"
	"go-gateway/app/app-svr/resource/service/model"
)

const (
	_refreshBWListWithGroup = `
		select
		  s.id as scene_id,
		  IFNULL(g.id, 0) as group_id,
		  IFNULL(g.low, 0) as low,
		  IFNULL(g.high, 0) as high,
		  IFNULL(g.token, '') as group_token,
		  IFNULL(g.is_deleted, 0) as is_group_deleted,
		  s.token as scene_token,
		  s.default_value,
		  s.large_oid_type,
		  s.large_list_url,
		  s.list_type,
		  IFNULL(g.show_without_login, 0) as show_without_login,
		  IFNULL(g.special_op, 0) as special_op,
		  IFNULL(g.white_list, '') as white_list
		from
		   black_white_scene as s left join black_white_gray_groups as g on g.scene_id = s.id
		where
		  s.status = 0
		  and s.is_online = 1`
)

func (d *Dao) GetBWListWithGroupFromDB() (list []*model.BWListWithGroup, err error) {
	var rows *sql.Rows
	ctx := context.Background()

	rows, err = d.db.Query(
		ctx,
		_refreshBWListWithGroup,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list = make([]*model.BWListWithGroup, 0)

	for rows.Next() {
		row := new(model.BWListWithGroup)
		var whiteList string
		if err = rows.Scan(
			&row.SceneId, &row.GroupID, &row.Low, &row.High, &row.GroupToken,
			&row.IsGroupDeleted, &row.SceneToken, &row.DefaultValue, &row.LargeOidType,
			&row.LargeListUrl, &row.ListType, &row.ShowWithoutLogin, &row.SpecialOp, &whiteList); err != nil {
			return nil, err
		}
		if len(whiteList) > 0 {
			row.WhiteList = strings.Split(whiteList, ",")
		}
		list = append(list, row)
	}

	if err = rows.Err(); err != nil {
		list = nil
	}

	return list, err
}

func (d *Dao) CheckLargeList(c context.Context, oids *pb2.LargeOidContent, checkURL string) (ok bool, err error) {
	if checkURL == "" || oids == nil {
		return false, nil
	}
	params := url.Values{}

	if uri, err := url.Parse(checkURL); err == nil {
		params = uri.Query()
		checkURL = "http://" + uri.Host + uri.Path
	}

	// 为了兼容现存接口
	params.Set("uid", strconv.FormatInt(oids.Mid, 10))
	params.Set("mid", strconv.FormatInt(oids.Mid, 10))
	params.Set("buvid", oids.Buvid)

	var res struct {
		Code int `json:"code"`
		Data struct {
			Status int `json:"status"`
		} `json:"data"`
	}
	if err = d.httpClient.Get(c, checkURL, "", params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), checkURL+"?"+params.Encode())
		return
	}
	//log.Error("[bw-list]第三方接口url(%+v), 参数(%+v)，返回结果(%+v)", checkURL, params, res)
	if res.Data.Status == 1 {
		ok = true
	}
	return
}
