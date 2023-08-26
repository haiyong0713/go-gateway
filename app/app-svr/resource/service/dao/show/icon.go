package show

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"

	api "go-gateway/app/app-svr/resource/service/api/v1"

	"github.com/pkg/errors"
)

const (
	_iconsSQL = "SELECT id,module,icon,global_red_dot,effect_group,effect_url,stime,etime FROM mng_icon WHERE stime<=? AND etime>=? AND state=1 ORDER BY id DESC"
)

// Icons is
func (d *Dao) Icons(c context.Context, startTime, endTime time.Time) (icons map[int64]*api.MngIcon, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(c, _iconsSQL, endTime, startTime); err != nil {
		err = errors.Wrapf(err, "d.db.Query err")
		return
	}
	icons = make(map[int64]*api.MngIcon)
	defer rows.Close()
	for rows.Next() {
		var (
			module string
			ic     = &api.MngIcon{}
		)
		if err = rows.Scan(&ic.Id, &module, &ic.Icon, &ic.GlobalRed, &ic.EffectGroup, &ic.EffectUrl, &ic.Stime, &ic.Etime); err != nil {
			log.Error("rows.Scan err(%+v)", err)
			err = nil
			continue
		}
		if err = json.Unmarshal([]byte(module), &ic.Module); err != nil {
			log.Error("json.Unmarshal err(%+v) module(%s)", err, module)
			err = nil
			continue
		}
		// 每个模块只会展示一个icon，有多个生效时优先取后配置的（按ID倒序
		for _, v := range ic.Module {
			if _, ok := icons[v.Oid]; ok {
				continue
			}
			icons[v.Oid] = ic
		}
	}
	err = rows.Err()
	return
}

// EffectUrl 指定用户生效 由业务方提供白名单接口
func (d *Dao) EffectUrl(c context.Context, mid int64, checkURL string) (bool, error) {
	var (
		params = url.Values{}
		res    struct {
			Code int `json:"code"`
			Data struct {
				Display bool `json:"display"`
			} `json:"data"`
		}
		err error
	)
	params.Set("mid", strconv.FormatInt(mid, 10))
	if err = d.httpClient.Get(c, checkURL, "", params, &res); err != nil {
		return false, err
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), checkURL+"?"+params.Encode())
		return false, err
	}
	return res.Data.Display, nil
}
