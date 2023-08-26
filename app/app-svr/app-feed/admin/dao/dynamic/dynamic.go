package dynamic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

// Dao .
type Dao struct {
	c              *conf.Config
	liveHTTPClient *bm.Client
	host           string
}

// New .
func New(c *conf.Config) *Dao {
	return &Dao{
		c:              c,
		liveHTTPClient: bm.NewClient(c.HTTPClient.Read),
		host:           c.Host.Dynamic,
	}
}

const (
	detailUrl        = "/dynamic_detail/v0/Dynamic/details"
	DynamicAuditPass = 0
	DynamicNotDel    = 0
)

type Dynamic struct {
	DynamicID    int64      `json:"dynamic_id,omitempty"`
	PublishTime  xtime.Time `json:"publish_time,omitempty"`
	Mid          int64      `json:"mid,omitempty"`
	RidType      int8       `json:"rid_type,omitempty"`
	Rid          int64      `json:"rid,omitempty"`
	ImgCount     int        `json:"img_count,omitempty"`
	Imgs         []string   `json:"imgs,omitempty"`
	DynamicText  string     `json:"dynamic_text,omitempty"`
	ViewCount    int64      `json:"view_count,omitempty"`
	Topics       []string   `json:"topics,omitempty"`
	NickName     string     `json:"nick_name,omitempty"`
	FaceImg      string     `json:"face_img,omitempty"`
	CommentCount int64      `json:"comment_count,omitempty"`
	LikeCount    int32      `json:"like_count,omitempty"`
	AuditStatus  int        `json:"audit_status"`
	DeleteStatus int        `json:"delete_status"`
}

// DynamicDetail .
func (d *Dao) DynamicDetail(c context.Context, ids []int64) (picm map[int64]*Dynamic, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	for _, id := range ids {
		params.Add("dynamic_ids[]", strconv.FormatInt(id, 10))
	}
	var res struct {
		Code int `json:"code"`
		Data *struct {
			List []*Dynamic `json:"list"`
		} `json:"data"`
	}
	url := d.host + detailUrl
	if err = d.liveHTTPClient.Get(c, url, ip, params, &res); err != nil {
		log.Error("DynamicDetail Req(%v) error(%v) res(%+v)", ids, err, res)
		return nil, fmt.Errorf(util.ErrorNetFmts, util.ErrorNet, url+"?"+params.Encode(), err.Error())
	}
	if res.Code != ecode.OK.Code() || res.Data == nil || len(res.Data.List) == 0 {
		log.Error("DynamicDetail Req(%v) error(%v) res(%+v)", ids, err, res)
		return nil, fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.c.UserFeed.Dynamic, url+"?"+params.Encode())
	}
	picm = make(map[int64]*Dynamic, len(res.Data.List))
	for _, pic := range res.Data.List {
		picm[pic.DynamicID] = pic
	}
	return
}
