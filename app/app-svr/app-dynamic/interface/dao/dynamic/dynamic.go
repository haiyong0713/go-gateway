package dynamic

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-dynamic/interface/api"
	dynmdl "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"

	xmetadata "go-common/library/net/metadata"

	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dynSvrFeedGrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	"github.com/pkg/errors"
)

const (
	_videoListURL  = "/dynamic_svr/v0/dynamic_svr/vd_new"
	_videoHistory  = "/dynamic_svr/v0/dynamic_svr/vd_history"
	_topicInfos    = "/topic_ext_svr/v0/topic_ext_svr/dynamic_topics"
	_likeIcon      = "/dynamic_like_icon/v0/dynamic_like_icon/query_icon"
	_videoPersonal = "/dynamic_svr/v0/dynamic_svr/vd_personal"
	_dynUpdOffset  = "/dynamic_svr/v0/dynamic_svr/vd_upd_offset"
	_getBottom     = "/dynamic_mix/v0/dynamic_mix/get_bottom"
	_vdUpList      = "/dynamic_svr/v0/dynamic_svr/vd_uplist"
	_dynBriefs     = "/dynamic_svr/v0/dynamic_svr/dyn_briefs"
	_sVideo        = "/dynamic_svr/v0/dynamic_svr/mix_video"
)

func (d *Dao) DynVideoList(c context.Context, updateBaseLine string, teenager int, uid int64) (*dynmdl.DynVideoListRes, error) {
	params := url.Values{}
	params.Set("teenagers_mode", strconv.Itoa(teenager))
	params.Set("update_baseline", updateBaseLine)
	params.Set("uid", strconv.FormatInt(uid, 10))
	videoListURL := d.videoList
	var ret struct {
		Code int                     `json:"code"`
		Msg  string                  `json:"msg"`
		Data *dynmdl.DynVideoListRes `json:"data"`
	}
	if err := d.client.Get(c, videoListURL, "", params, &ret); err != nil {
		log.Errorc(c, "DynVideoList http GET(%s) failed, params:(%s), error(%+v)", videoListURL, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "DynVideoList http GET(%s) failed, params:(%s), code: %v, msg: %v", videoListURL, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "videoList url(%v) code(%v) msg(%v)", videoListURL, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}

func (d *Dao) DynVideoHistory(c context.Context, offset string, teenager, page int, uid int64) (*dynmdl.DynVideoListRes, error) {
	params := url.Values{}
	params.Set("teenagers_mode", strconv.Itoa(teenager))
	params.Set("offset", offset)
	params.Set("page", strconv.Itoa(page))
	params.Set("uid", strconv.FormatInt(uid, 10))
	vdHistoryURL := d.videoHistory
	var ret struct {
		Code int                     `json:"code"`
		Msg  string                  `json:"msg"`
		Data *dynmdl.DynVideoListRes `json:"data"`
	}
	if err := d.client.Get(c, vdHistoryURL, "", params, &ret); err != nil {
		log.Errorc(c, "DynVideoHistory http GET(%s) failed, params:(%s), error(%v)", vdHistoryURL, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "DynVideoHistory http GET(%s) failed, params:(%s), code: %v, msg: %v", vdHistoryURL, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "DynVideoHistory url(%v) code(%v) msg(%v)", vdHistoryURL, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}

func (d *Dao) TopicInfos(c context.Context, dynIDs []int64, mobiApp, platform, device string, build int) (*dynmdl.TopicRes, error) {
	flexes := dynmdl.TopicParasParams{
		Paras: []dynmdl.Paras{
			{
				List: []dynmdl.TopicParasList{
					{
						Key:   "build",
						Value: strconv.Itoa(build),
					},
					{
						Key:   "device",
						Value: device,
					},
					{
						Key:   "mobi_app",
						Value: mobiApp,
					},
					{
						Key:   "platform",
						Value: platform,
					},
				},
				Type: 268435455,
			},
		},
	}
	flexesJosn, err := json.Marshal(flexes)
	dynTopicReq := &dynmdl.DynTopicsParams{
		DynIDs: dynIDs,
		Flexes: string(flexesJosn),
	}
	if err != nil {
		log.Errorc(c, "TopicInfos json.Marshal() failed. error(%v)", err)
		return nil, err
	}
	topicInfo := d.topicInfo
	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(dynTopicReq); err != nil {
		log.Error("TopicInfos json.NewEncoder() error(%v)", err)
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, topicInfo, b)
	if err != nil {
		log.Errorc(c, "TopicInfos error(%v)", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	var ret struct {
		Code int              `json:"code"`
		Msg  string           `json:"msg"`
		Data *dynmdl.TopicRes `json:"data"`
	}
	if err := d.client.Do(c, req, &ret); err != nil {
		log.Errorc(c, "TopicInfos http POST(%s) failed, params:(%s), error(%v)", topicInfo, b.String(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "TopicInfos http POST(%s) failed, params:(%s), code: %v, msg: %v", topicInfo, b.String(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "TopicInfos url(%v) code(%v) msg(%v)", topicInfo, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}

func (d *Dao) LikeIcon(c context.Context, params dynmdl.LikeIconReq) (*dynmdl.LikeIcon, error) {
	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(params); err != nil {
		log.Errorc(c, "LikeIcon json.NewEncoder() error(%+v)", err)
		return nil, err
	}
	lkIconURL := d.likeIcon
	req, err := http.NewRequest(http.MethodPost, lkIconURL, b)
	if err != nil {
		log.Errorc(c, "LikeIcon error(%+v)", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	var ret struct {
		Code int              `json:"code"`
		Msg  string           `json:"msg"`
		Data *dynmdl.LikeIcon `json:"data"`
	}
	if err := d.client.Do(c, req, &ret); err != nil {
		log.Errorc(c, "LikeIcon http POST(%s) failed, params:(%s), error(%+v)", lkIconURL, b.String(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "LikeIcon http POST(%s) failed, params:(%s), code: %v, msg: %v", lkIconURL, b.String(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "LikeIcon url(%v) code(%v) msg(%v)", lkIconURL, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}

func (d *Dao) VideoPersonal(c context.Context, req *dynmdl.DynVideoPersonalReq) (*dynmdl.VideoPersonalRes, error) {
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(req.Mid, 10))
	params.Set("host_uid", strconv.FormatInt(req.HostUID, 10))
	params.Set("offset", req.Offset)
	params.Set("page", strconv.Itoa(req.Page))
	params.Set("is_preload", strconv.Itoa(req.IsPreload))
	videoPersonalURL := d.videoPersonal
	var ret struct {
		Code int                      `json:"code"`
		Msg  string                   `json:"msg"`
		Data *dynmdl.VideoPersonalRes `json:"data"`
	}
	if err := d.client.Get(c, videoPersonalURL, "", params, &ret); err != nil {
		log.Errorc(c, "VideoPersonal http GET(%s) failed, params:(%s), error(%+v)", videoPersonalURL, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "VideoPersonal http GET(%s) failed, params:(%s), code: %v, msg: %v", videoPersonalURL, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "VideoPersonal url(%v) code(%v) msg(%v)", videoPersonalURL, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}

func (d *Dao) DynUpdOffset(c context.Context, req *dynmdl.DynUpdOffsetReq) error {
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(req.Mid, 10))
	params.Set("host_uid", strconv.FormatInt(req.HostUID, 10))
	params.Set("offset", req.ReadOffset)
	updOffsetURL := d.dynUpdOffset
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := d.client.Get(c, updOffsetURL, "", params, &ret); err != nil {
		log.Errorc(c, "DynUpdOffset http GET(%s) failed, params(%s), error(%+v)", updOffsetURL, params.Encode(), err)
		return err
	}
	if ret.Code != 0 {
		log.Errorc(c, "DynUpdOffset http GET(%s) failed. params(%s), code: %v, msg: %v", updOffsetURL, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "DynUpdOffset url(%v) code(%v) msg(%v)", updOffsetURL, ret.Code, ret.Msg)
		return err
	}
	return nil
}

func (d *Dao) GetBottom(c context.Context, params *dynmdl.DynBottomReq) (*dynmdl.BottomRes, error) {
	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(params); err != nil {
		log.Errorc(c, "GetBottom json.NewEncoder() error(%+v)", err)
		return nil, err
	}
	bottomURL := d.bottom
	req, err := http.NewRequest(http.MethodPost, bottomURL, b)
	if err != nil {
		log.Errorc(c, "GetBottom error(%+v)", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	var ret struct {
		Code int               `json:"code"`
		Msg  string            `json:"msg"`
		Data *dynmdl.BottomRes `json:"data"`
	}
	if err := d.client.Do(c, req, &ret); err != nil {
		log.Errorc(c, "GetBottom http POST(%s) failed, params:(%s), error(%+v)", bottomURL, b.String(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "GetBottom http POST(%s) failed, params:(%s), code: %v, msg: %v", bottomURL, b.String(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "GetBottom url(%v) code(%v) msg(%v)", bottomURL, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}

func (d *Dao) VdUpList(c context.Context, teenager int, uid int64, buvid string) (*dynmdl.VdUpListRsp, error) {
	params := url.Values{}
	params.Set("teenagers_mode", strconv.Itoa(teenager))
	params.Set("uid", strconv.FormatInt(uid, 10))
	params.Set("buvid", buvid)
	upList := d.vdUpList
	var ret struct {
		Code int                 `json:"code"`
		Msg  string              `json:"msg"`
		Data *dynmdl.VdUpListRsp `json:"data"`
	}
	if err := d.client.Get(c, upList, "", params, &ret); err != nil {
		log.Errorc(c, "VdUpList http GET(%s) failed, params(%s), error(%+v)", upList, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "VdUpList http GET(%s) failed. params(%s), code: %v, msg: %v", upList, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "VdUpList url(%v) code(%v) msg(%v)", upList, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}

func (d *Dao) DynBriefs(c context.Context, teenager int, uid int64, dynIDs []int64) (*dynmdl.DynDetailRsp, error) {
	params := dynmdl.DynBriefsReq{
		Teenager: teenager,
		DynIDs:   dynIDs,
		Uid:      uid,
	}
	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(params); err != nil {
		log.Errorc(c, "DynBriefs json.NewEncoder() error(%+v)", err)
		return nil, err
	}
	dynBriefs := d.dynBriefs
	req, err := http.NewRequest(http.MethodPost, dynBriefs, b)
	if err != nil {
		log.Errorc(c, "DynBriefs error(%+v)", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	var ret struct {
		Code int                  `json:"code"`
		Msg  string               `json:"msg"`
		Data *dynmdl.DynDetailRsp `json:"data"`
	}
	if err := d.client.Do(c, req, &ret); err != nil {
		log.Errorc(c, "DynBriefs http POST(%s) failed, params:(%s), error(%+v)", dynBriefs, b.String(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "DynBriefs http POST(%s) failed, params:(%s), code: %v, msg: %v", dynBriefs, b.String(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "DynBriefs url(%v) code(%v) msg(%v)", dynBriefs, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}

func (d *Dao) SVideo(c context.Context, offset string, needOffset int, uid int64) (*dynmdl.DynSVideoList, error) {
	params := url.Values{}
	params.Set("offset", offset)
	params.Set("need_offset", strconv.Itoa(needOffset))
	params.Set("uid", strconv.FormatInt(uid, 10))
	var ret struct {
		Code int                   `json:"code"`
		Msg  string                `json:"msg"`
		Data *dynmdl.DynSVideoList `json:"data"`
	}
	if err := d.client.Get(c, d.svideo, "", params, &ret); err != nil {
		err = errors.Wrapf(err, "SVideo d.client.Get url(%s)", d.svideo+"?"+params.Encode())
		return nil, err
	}
	if ret.Code != 0 {
		err := errors.Wrapf(ecode.Int(ret.Code), "SVideo code(%d) url(%s)", ret.Code, d.svideo+"?"+params.Encode())
		return nil, err
	}
	return ret.Data, nil
}

// nolint:gomnd
func (d *Dao) UpdateNum(c context.Context, mid int64, req *api.DynRedReq, mobiApp string, header *dynmdl.Header) (res *dynSvrFeedGrpc.UpdateNumResp, err error) {
	var arg = &dynSvrFeedGrpc.UpdateNumReq{
		Uid: mid,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:    strconv.Itoa(header.Build),
			Platform: header.Platform,
			MobiApp:  header.MobiApp,
			Buvid:    header.Buvid,
			Device:   header.Device,
			Ip:       xmetadata.String(c, xmetadata.RemoteIP),
		},
	}
	for _, r := range req.GetTabOffset() {
		argInfo := &dynSvrFeedGrpc.OffsetInfo{
			Tab:    r.Tab,
			Offset: r.Offset,
		}
		if argInfo.Tab == 1 {
			argInfo.TypeList = "268435455"
			if mobiApp == "ipad" || mobiApp == "android_hd" {
				argInfo.TypeList = "1,2,4,8,512,2048,2049,4097,4098,4099,4100,4101,4307" // ipad HD单独计算
			}
		} else if argInfo.Tab == 2 {
			argInfo.TypeList = "8,512,4097,4098,4099,4100,4101,4303"
			if mobiApp == "ipad" || mobiApp == "android_hd" {
				argInfo.TypeList = "8,512,4097,4098,4099,4100,4101" // ipad HD单独计算
			}
		} else {
			log.Error("UpdateNum unknow mid %v tab %v", mid, r.GetTab())
			continue
		}
		arg.Offsets = append(arg.Offsets, argInfo)
	}
	if res, err = d.dynaGRPC.UpdateNum(c, arg); err != nil {
		log.Error("%v", err)
	}
	return
}
