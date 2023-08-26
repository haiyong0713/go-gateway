package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/library/ecode"
	"go-common/library/log"
	infocv2 "go-common/library/log/infoc.v2"
	"go-common/library/net/metadata"
	"go-common/library/text/translate/chinese.v2"
	"go-gateway/app/app-svr/app-search/internal/model/search"

	"github.com/pkg/errors"
)

func (d *dao) AiRecommend(c context.Context) (rs map[int64]struct{}, err error) {
	params := url.Values{}
	params.Set("cmd", "hot")
	params.Set("from", "10")
	timeout := time.Duration(d.c.Custom.RecommendTimeout) / time.Millisecond
	params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	params.Set("ignore_custom", "1")
	var res struct {
		Code int `json:"code"`
		Data []struct {
			Id int64 `json:"id"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.rcmdAi, "", params, &res); err != nil {
		return
	}
	if res.Code != 0 {
		// err = errors.Wrap(err, fmt.Sprintf("code(%d)", res.Code))
		err = errors.Wrapf(ecode.Int(res.Code), "recommend url(%s) code(%d)", d.rcmdAi, res.Code)
		return
	}
	rs = map[int64]struct{}{}
	for _, l := range res.Data {
		if l.Id > 0 {
			rs[l.Id] = struct{}{}
		}
	}
	return
}

func (d *dao) GetAiRecommendTags(ctx context.Context, style, numNot1st, nonPersonality int64, gt, id1st string, isHant bool) (*search.RecommendTagsRsp, error) {
	var (
		req *http.Request
		ip  = metadata.String(ctx, metadata.RemoteIP)
		err error
	)
	// 获取鉴权 mid
	au, _ := auth.FromContext(ctx)
	// 获取设备信息
	dev, _ := device.FromContext(ctx)
	params := url.Values{}
	params.Set("cmd", "tm_search_guide")
	params.Set("timeout", d.c.SearchRcmdTagsConfig.AiRcmdTimeout)
	params.Set("mid", strconv.FormatInt(au.Mid, 10))
	params.Set("buvid", dev.Buvid)
	params.Set("build", strconv.FormatInt(dev.Build, 10))
	params.Set("plat", strconv.Itoa(int(dev.Plat())))
	params.Set("style", strconv.FormatInt(style, 10))
	params.Set("ip", ip)
	params.Set("network", dev.Network)
	params.Set("mobi_app", dev.RawMobiApp)
	params.Set("goto", gt)
	params.Set("id_1st", id1st)
	params.Set("num_not_1st", strconv.FormatInt(numNot1st, 10))
	params.Set("non_personality", strconv.FormatInt(nonPersonality, 10))
	if req, err = d.client.NewRequest("GET", d.rcmdTag, ip, params); err != nil {
		return nil, err
	}
	var res struct {
		Code    int    `json:"code"`
		TrackId string `json:"trackid"`
		Data    struct {
			Title string                `json:"title"`
			State int64                 `json:"state"`
			Items []*search.TagItemList `json:"itemlist"`
		} `json:"data"`
		PvFeature string `json:"pv_feature"`
	}
	if err = d.client.Do(ctx, req, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.rcmdTag+"?"+params.Encode())
		return nil, err
	}
	if isHant {
		// 简繁体转换
		res.Data.Title = chinese.Convert(ctx, res.Data.Title)
		for _, v := range res.Data.Items {
			v.Query = chinese.Convert(ctx, v.Query)
		}
	}
	var tags []*search.RecommendTag
	for _, v := range res.Data.Items {
		if v == nil {
			continue
		}
		tags = append(tags, &search.RecommendTag{Query: v.Query, JumpUrl: fmt.Sprintf("bilibili://search?search_from_source=app_tm_guide&keyword=%s&direct_return=true", makeAiRecommendRealQuery(v))})
	}
	defer func() {
		dataJson, _ := json.Marshal(res.Data)
		payload := infocv2.NewLogStreamV("016332", log.String(params.Get("ip")),
			log.String(strconv.FormatInt(time.Now().Unix(), 10)),
			log.String(res.TrackId),
			log.String(params.Get("buvid")),
			log.String(params.Get("mid")),
			log.String(params.Get("id_1st")),
			log.String(params.Get("num_not_1st")),
			log.String(params.Get("goto")),
			log.String(params.Get("mobi_app")),
			log.String(params.Get("plat")),
			log.String(params.Get("build")),
			log.String(strconv.Itoa(res.Code)),
			log.String(string(dataJson)),
			log.String(res.PvFeature),
			log.String(params.Get("non_personality")),
		)
		d.uploadAiRecommendInfoc(ctx, payload)
	}()
	return &search.RecommendTagsRsp{Title: res.Data.Title, Tags: tags}, nil
}

func makeAiRecommendRealQuery(tagItem *search.TagItemList) string {
	if tagItem.SearchQuery != "" {
		return tagItem.SearchQuery
	}
	return tagItem.Query
}

func (d *dao) uploadAiRecommendInfoc(ctx context.Context, payload infocv2.Payload) {
	if err := d.infocv2.Info(ctx, payload); err != nil {
		log.Warn("uploadAiRecommendInfoc() d.pubInfocv2.Info() payload(%+v) error(%+v)", payload, err)
	}
}
