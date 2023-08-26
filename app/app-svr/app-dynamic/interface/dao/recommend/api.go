package recommend

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"

	"github.com/pkg/errors"
)

const (
	_recommand = "/recommand"
)

func (d *Dao) Recommend(c context.Context, general *mdlv2.GeneralParam, prevCampusId int64, userCampusId int64, reqArgs map[string]string, avidInfocM bool) (*mdlv2.RcmdReply, bool, error) {
	var (
		params = url.Values{}
	)
	// https://info.bilibili.co/pages/viewpage.action?pageId=416278182
	params.Set("cmd", "campus_nearby")
	timeout := time.Duration(d.c.HTTPData.Timeout)
	if ddl, ok := c.Deadline(); ok && ddl.Unix() > 0 {
		// timeout和context ddl取短的那个
		if dur := time.Until(ddl); dur > 0 && dur < timeout {
			timeout = dur
		}
	}
	params.Set("timeout", strconv.FormatInt(timeout.Milliseconds(), 10))
	params.Set("mid", strconv.FormatInt(general.Mid, 10))
	params.Set("buvid", general.GetBuvid())
	params.Set("build", general.GetBuildStr())
	params.Set("plat", strconv.Itoa(int(general.Plat)))
	params.Set("request_cnt", "10")
	params.Set("previous_campus_id", strconv.FormatInt(prevCampusId, 10))
	params.Set("user_campus_id", strconv.FormatInt(userCampusId, 10))
	params.Set("network", general.GetNetWork())
	params.Set("ip", general.Network.RemoteIP)
	params.Set("mobi_app", general.GetMobiApp())
	params.Set("page_type", "nearby")
	for k, v := range reqArgs {
		params.Set(k, v)
	}

	res := new(mdlv2.RcmdReply)
	if err := d.client.Get(c, d.recommand, "", params, res); err != nil {
		return nil, false, err
	}
	// infoc
	infocTmp := &mdlv2.RcmdInfo{}
	if avidInfocM {
		infocTmp.FromRcmdInfoAvID(res)
	} else {
		infocTmp.FromRcmdInfoDynID(res)
	}
	res.Infoc = infocTmp
	// code
	const (
		expectedNoMoreRcmd = -11
	)
	if code := ecode.Int(res.Code); !ecode.Equal(code, ecode.OK) && code != expectedNoMoreRcmd {
		return nil, false, errors.WithMessagef(code, d.recommand+"?"+params.Encode())
	}
	// 正常情况下只要 code != -11 就是hasMore=True
	return res, res.Code != expectedNoMoreRcmd, nil
}
