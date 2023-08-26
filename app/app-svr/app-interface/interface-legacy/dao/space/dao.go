package space

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/naming/discovery"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"
	spaceclient "go-gateway/app/web-svr/space/interface/api/v1"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

const (
	_uploadTopPhotoURL = "/api/member/getUploadTopPhoto"
	_resetTopPhotoURL  = "/api/member/clearTopPhoto"
	_report            = "/api/report/add"
	_blacklist         = "/x/internal/space/blacklist"
	_prinfo            = "/x/internal/space/system/notice"
)

// Dao is space dao
type Dao struct {
	client     *httpx.Client
	clientSync *httpx.Client
	report     string
	// space api
	uploadTop      string
	resetTop       string
	blacklist      string
	prinfo         string
	spaceClient    spaceclient.SpaceClient
	photoArcClient spaceclient.SpaceClient
	upRcmdClient   spaceclient.SpaceClient
}

// New initial space dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:     httpx.NewClient(c.HTTPClient, httpx.SetResolver(resolver.New(nil, discovery.Builder()))),
		clientSync: httpx.NewClient(c.HTTPWrite),
		report:     c.Host.Space + _report,
		uploadTop:  c.Host.Space + _uploadTopPhotoURL,
		resetTop:   c.Host.Space + _resetTopPhotoURL,
		blacklist:  c.Host.APICo + _blacklist,
		prinfo:     c.Host.APICo + _prinfo,
	}
	var err error
	if d.spaceClient, err = spaceclient.NewClient(c.SpaceClient); err != nil {
		panic(err)
	}
	if d.photoArcClient, err = spaceclient.NewClient(c.SpacePhotoClient); err != nil {
		panic(err)
	}
	if d.upRcmdClient, err = spaceclient.NewClient(c.SpaceUpRcmdClient); err != nil {
		panic(err)
	}
	return
}

// Setting get setting data from api.
func (d *Dao) Setting(c context.Context, mid int64) (*space.Setting, error) {
	res, err := d.spaceClient.SpaceSetting(c, &spaceclient.SpaceSettingReq{Mid: mid})
	if err != nil {
		return nil, errors.Wrapf(err, "d.spaceClient.SpaceSetting mid=%d", mid)
	}
	return &space.Setting{
		Channel:           int(res.Channel),
		FavVideo:          int(res.FavVideo),
		CoinsVideo:        int(res.CoinsVideo),
		LikesVideo:        int(res.LikesVideo),
		Bangumi:           int(res.Bangumi),
		PlayedGame:        int(res.PlayedGame),
		Groups:            int(res.Groups),
		Comic:             int(res.Comic),
		BBQ:               int(res.BBQ),
		DressUp:           int(res.DressUp),
		DisableFollowing:  int(res.DisableFollowing),
		LivePlayback:      int(res.LivePlayback),
		CloseSpaceMedal:   int(res.CloseSpaceMedal),
		OnlyShowWearing:   int(res.OnlyShowWearing),
		DisableShowSchool: int(res.DisableShowSchool),
		DisableShowNft:    int(res.DisableShowNft),
	}, nil
}

// SpaceMob space mobile
func (d *Dao) SpaceMob(c context.Context, mid, vmid int64, platform, device string) (us string, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("vmid", strconv.FormatInt(vmid, 10))
	params.Set("platform", platform)
	params.Set("device", device)
	var res struct {
		Code int             `json:"code"`
		Data *space.TopPhoto `json:"data"`
	}
	if err = d.client.Get(c, d.uploadTop, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.uploadTop+"?"+params.Encode())
		return
	}
	if res.Data == nil { // if data is nil, regard it as not found
		err = errors.Wrap(ecode.NothingFound, d.uploadTop+"?"+params.Encode())
		return
	}
	us = res.Data.ImgURL
	return
}

// Report report
func (d *Dao) Report(c context.Context, mid int64, reason, ak string) (err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("access_key", ak)
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("reason", reason)
	var res struct {
		Code int `json:"code"`
	}
	if err = d.client.Post(c, d.report, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.report+"?"+params.Encode())
	}
	return
}

// UpRcmdBlockMap is used to block up-recommend in space.
func (d *Dao) UpRcmdBlockMap(c context.Context) (map[int64]struct{}, error) {
	reply, err := d.upRcmdClient.UpRcmdBlackList(c, &empty.Empty{})
	if err != nil {
		return nil, err
	}
	blockMap := make(map[int64]struct{})
	for _, v := range reply.BannedMids {
		blockMap[v] = struct{}{}
	}
	return blockMap, nil
}

// Blacklist is.
func (d *Dao) Blacklist(c context.Context) (list map[int64]struct{}, err error) {
	var res struct {
		Code int                `json:"code"`
		Data map[int64]struct{} `json:"data"`
	}
	if err = d.clientSync.Get(c, d.blacklist, "", nil, &res); err != nil {
		err = errors.Wrap(ecode.Int(res.Code), d.blacklist)
		return
	}
	b, _ := json.Marshal(res)
	log.Error("Blacklist url(%s) response(%s)", d.blacklist, b)
	list = res.Data
	return
}

// PRInfo isinfo.
func (d *Dao) PRInfo(c context.Context, mid int64) (*space.PRInfo, error) {
	params := url.Values{}
	params.Set("id", strconv.FormatInt(mid, 10))
	var res struct {
		Code int           `json:"code"`
		Data *space.PRInfo `json:"data"`
	}
	var err error
	if err = d.client.Get(c, d.prinfo, "", params, &res); err != nil {
		err = errors.Wrap(ecode.Int(res.Code), d.blacklist)
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.prinfo+"?"+params.Encode())
		return nil, err
	}
	return res.Data, nil
}

// TopphotoReset resets the vip's top photo
func (d *Dao) TopphotoReset(c context.Context, accesskey, platform, device string) (err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("access_key", accesskey)
	params.Set("platform", platform)
	params.Set("device", device)
	var res struct {
		Code int `json:"code"`
	}
	if err = d.client.Get(c, d.resetTop, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.resetTop+"?"+params.Encode())
	}
	return
}

func (d *Dao) OfficialDownload(c context.Context, mid int64, plat int8) (res *space.OfficialItem, err error) {
	var reply *spaceclient.OfficialReply
	if reply, err = d.spaceClient.Official(c, &spaceclient.OfficialRequest{Mid: mid}); err != nil {
		if ecode.EqualError(ecode.NothingFound, err) {
			err = nil
			return
		}
		log.Error("OfficialDownload d.spaceClient.Official(%d) error(%v)", mid, err)
		return
	}
	if reply == nil {
		return
	}
	res = &space.OfficialItem{
		Uid:    reply.Uid,
		Name:   reply.Name,
		Icon:   reply.Icon,
		Scheme: reply.Scheme,
		Rcmd:   reply.Rcmd,
		Button: reply.Button,
	}
	res.URL = reply.AndroidUrl
	if model.IsIOS(plat) || model.IsPad(plat) || model.IsIPhoneB(plat) {
		res.URL = reply.IosUrl
	}
	return
}

// PhotoMallList 获取app端空间头图列表
func (d *Dao) PhotoMallList(c context.Context, mobiApp, device string, mid int64) (res []*space.PhotoMallItem, err error) {
	reply, err := d.spaceClient.PhotoMallList(c, &spaceclient.PhotoMallListReq{Mobiapp: mobiApp, Device: device, Mid: mid})
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if reply == nil {
		return
	}
	for _, v := range reply.List {
		i := &space.PhotoMallItem{}
		i.FromPhotoMallItem(v)
		res = append(res, i)
	}
	return
}

// PhotoTopSet 设置app端空间默认头图
func (d *Dao) PhotoTopSet(c context.Context, mobiApp string, id, mid, typ int64) (err error) {
	if _, err = d.photoArcClient.SetTopPhoto(c, &spaceclient.SetTopPhotoReq{Mobiapp: mobiApp, ID: id, Mid: mid, Type: spaceclient.TopPhotoType(typ)}); err != nil {
		log.Error("%+v", err)
	}
	return
}

// TopPhoto 获取头图
func (d *Dao) TopPhoto(c context.Context, mobiApp, device string, build int, vmid, mid int64) (res *spaceclient.TopPhoto, arc *spaceclient.TopPhotoArc, err error) {
	reply, err := d.spaceClient.TopPhoto(c, &spaceclient.TopPhotoReq{Mobiapp: mobiApp, Device: device, Mid: vmid, Build: int32(build), LoginMid: mid})
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if reply == nil {
		return
	}
	res = reply.TopPhoto
	arc = reply.TopPhotoArc
	return
}

func (d *Dao) ActivityTab(ctx context.Context, mid int64, plat, build int32) (*spaceclient.UserTabReply, error) {
	reply, err := d.spaceClient.UserTab(ctx, &spaceclient.UserTabReq{
		Mid:   mid,
		Plat:  plat,
		Build: build,
	})
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) TopPhotoArcCancel(ctx context.Context, mid int64) error {
	_, err := d.spaceClient.TopPhotoArcCancel(ctx, &spaceclient.TopPhotoArcCancelReq{Mid: mid})
	if err != nil {
		log.Error("%+v", err)
	}
	return err
}
