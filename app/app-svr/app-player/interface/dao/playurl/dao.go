package player

import (
	"context"
	"fmt"
	"go-common/library/log"
	api "go-gateway/app/app-svr/app-player/interface/api/playurl"
	"go-gateway/app/app-svr/app-player/interface/conf"
	"go-gateway/app/app-svr/app-player/interface/model"
	v2 "go-gateway/app/app-svr/playurl/service/api/v2"

	ott "git.bilibili.co/bapis/bapis-go/ott/service"
	"github.com/pkg/errors"
)

var (
	_fhMobiAppMap = map[string]struct{}{
		"android":    {},
		"android_tv": {},
		"android_G":  {},
		"android_i":  {},
		"iphone":     {},
		"ipad":       {},
		"white":      {},
	}
)

// Dao is player dao.
type Dao struct {
	conf *conf.Config
	// rpc
	playURLRPCV2 v2.PlayURLClient
	ottCli       ott.OTTServiceClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		conf: c,
	}
	var err error
	d.playURLRPCV2, err = v2.NewClient(c.PlayURLClient)
	if err != nil {
		panic(fmt.Sprintf("player v2 NewClient error(%v)", err))
	}
	if d.ottCli, err = ott.NewClient(c.OttClient); err != nil {
		panic(err)
	}
	return
}

// PlayURLV2 is
func (d *Dao) PlayURLV2(c context.Context, params *model.Param, mid, upgradeAid, upgradeCid int64) (player *v2.PlayURLReply, err error) {
	// 以下参数转换见视频云tapd地址https://www.tapd.cn/20095661/prong/stories/view/1120095661001131850
	fh := int32(1)            //force_host默认是1
	if params.ForceHost > 0 { //客户端有值即透传
		fh = params.ForceHost
	} else if _, ok := _fhMobiAppMap[params.MobiApp]; ok { //未传值判断platform
		fh = 0
	}
	req := &v2.PlayURLReq{
		Aid:        params.AID,
		Cid:        params.CID,
		Qn:         params.Qn,
		Platform:   params.Platform,
		Fnver:      params.Fnver,
		Fnval:      params.Fnval,
		ForceHost:  fh,
		Mid:        mid,
		UpgradeAid: upgradeAid,
		UpgradeCid: upgradeCid,
		Fourk:      d.checkFourk(params.MobiApp, params.Build, params.FourkBool),
		Device:     params.Device,
		MobiApp:    params.MobiApp,
		Download:   params.Download,
		BackupNum:  d.conf.Custom.BackupNum, //客户端请求默认2个
		Build:      params.Build,
		Buvid:      params.Buvid,
		NetType:    v2.NetworkType(params.NetType),
		TfType:     v2.TFType(params.TfType),
	}
	if needVerify := d.verifySteins(params); needVerify {
		req.VerifySteins = 1
	}
	if params.Download > 0 {
		req.ForceHost = 2 //离线下载默认https
	}
	// 只对粉版做限制
	if d.conf.Switch.VipControl && (params.MobiApp == "iphone" || params.MobiApp == "android") {
		req.VerifyVip = 1
	}
	if player, err = d.playURLRPCV2.PlayURL(c, req); err != nil {
		err = errors.Wrapf(err, "d.playURLRPCV2.PlayURL args(%v)", req)
		return
	}
	return
}

// feature PlayurlSteins
func (d *Dao) verifySteins(params *model.Param) (needVerify bool) { // 低于x版本，要求playurl-service校验互动视频
	if (params.MobiApp == "android" && params.Build < d.conf.Custom.SteinsBuild.Android) ||
		(params.MobiApp == "iphone" && params.Build <= d.conf.Custom.SteinsBuild.Iphone) ||
		(params.MobiApp == "iphone_b" && params.Build < d.conf.Custom.SteinsBuild.IphoneB) ||
		(params.MobiApp == "ipad" && params.Build < d.conf.Custom.SteinsBuild.IpadHD) ||
		(params.MobiApp == "android_i" && params.Build < d.conf.Custom.SteinsBuild.AndroidI) ||
		(params.MobiApp == "iphone_i" && params.Build <= d.conf.Custom.SteinsBuild.IphoneI) ||
		(params.MobiApp == "android_hd") {
		return true
	}
	return
}

// Project is
func (d *Dao) Project(c context.Context, params *model.Param, mid int64) (res *v2.ProjectReply, err error) {
	// 以下参数转换见视频云tapd地址https://www.tapd.cn/20095661/prong/stories/view/1120095661001131850
	fh := int32(1)            //force_host默认是1
	if params.ForceHost > 0 { //客户端有值即透传
		fh = params.ForceHost
	} else if _, ok := _fhMobiAppMap[params.MobiApp]; ok { //未传值判断platform
		fh = 0
	}
	req := &v2.ProjectReq{
		Aid:        params.AID,
		Cid:        params.CID,
		Qn:         params.Qn,
		Platform:   params.Platform,
		Fnver:      params.Fnver,
		Fnval:      params.Fnval,
		ForceHost:  fh,
		Mid:        mid,
		Fourk:      d.checkFourk(params.MobiApp, params.Build, params.FourkBool),
		Device:     params.Device,
		MobiApp:    params.MobiApp,
		Download:   params.Download,
		BackupNum:  2, //客户端请求默认2个
		Protocol:   params.Protocol,
		DeviceType: params.DeviceType,
		Business:   v2.Business_UGC,
		Buvid:      params.Buvid,
	}
	if params.Download > 0 {
		req.ForceHost = 2 //离线下载默认https
	}
	if res, err = d.playURLRPCV2.Project(c, req); err != nil {
		err = errors.Wrapf(err, "d.playURLRPCV2.Project args(%v)", req)
		return
	}
	return
}

// PlayView is
func (d *Dao) PlayView(c context.Context, params *model.Param, mid, upgradeAid, upgradeCid int64) (player *v2.PlayViewReply, err error) {
	// 以下参数转换见视频云tapd地址https://www.tapd.cn/20095661/prong/stories/view/1120095661001131850
	fh := int32(1)            //force_host默认是1
	if params.ForceHost > 0 { //客户端有值即透传
		fh = params.ForceHost
	} else if _, ok := _fhMobiAppMap[params.MobiApp]; ok { //未传值判断platform
		fh = 0
	}
	req := &v2.PlayViewReq{
		Aid:            params.AID,
		Cid:            params.CID,
		Qn:             params.Qn,
		Platform:       params.Platform,
		Fnver:          params.Fnver,
		Fnval:          params.Fnval,
		ForceHost:      fh,
		Mid:            mid,
		UpgradeAid:     upgradeAid,
		UpgradeCid:     upgradeCid,
		Fourk:          d.checkFourk(params.MobiApp, params.Build, params.FourkBool),
		Device:         params.Device,
		MobiApp:        params.MobiApp,
		Download:       params.Download,
		BackupNum:      d.conf.Custom.BackupNum, //客户端请求默认2个
		Build:          params.Build,
		Buvid:          params.Buvid,
		TeenagersMode:  params.TeenagersMode,
		NetType:        v2.NetworkType(params.NetType),
		TfType:         v2.TFType(params.TfType),
		LessonsMode:    params.LessonsMode,
		BusinessSource: v2.BusinessSource(params.Business),
		VoiceBalance:   params.VoiceBalance,
	}
	if needVerify := d.verifySteins(params); needVerify {
		req.VerifySteins = 1
	}
	if params.Download > 0 {
		req.ForceHost = 2 //离线下载默认https
	}
	// 只对粉版做限制
	if d.conf.Switch.VipControl && (params.MobiApp == "iphone" || params.MobiApp == "android") {
		req.VerifyVip = 1
	}
	if player, err = d.playURLRPCV2.PlayView(c, req); err != nil {
		log.Error("d.playURLRPCV2.PlayView req(%+v) error(%+v)", req, err)
		return
	}
	return
}

// PlayEdit .
func (d *Dao) PlayEdit(c context.Context, cloudParam *model.CloudEditParam, args *api.PlayConfEditReq) error {
	req := &v2.PlayConfEditReq{
		Buvid:    cloudParam.Buvid,
		Platform: cloudParam.Platform,
		Build:    int32(cloudParam.Build),
		Brand:    cloudParam.Brand,
		Model:    cloudParam.Model,
		FMode:    1,
	}
	req.PlayConf = model.ConfConvert(args.PlayConf)
	//没有合适的编辑，不需要变更，version不变
	if len(req.PlayConf) == 0 {
		return nil
	}
	if _, err := d.playURLRPCV2.PlayConfEdit(c, req); err != nil {
		err = errors.Wrapf(err, "d.playURLRPCV2.PlayConfEdit args(%v)", req)
		return err
	}
	return nil
}

// PlayConf .
func (d *Dao) PlayConf(c context.Context, cParam *model.CloudEditParam, mid int64) (*v2.PlayConfReply, error) {
	req := &v2.PlayConfReq{
		Buvid:    cParam.Buvid,
		Mid:      mid,
		Platform: cParam.Platform,
		Build:    int32(cParam.Build),
		Brand:    cParam.Brand,
		Model:    cParam.Model,
		FMode:    1,
	}
	return d.playURLRPCV2.PlayConf(c, req)
}

// 4k需求由于android老版本缓存时缺少fnval和fnver，不出4k，所以增加版本控制
func (d *Dao) checkFourk(mobiApp string, build int32, fourk bool) bool {
	// feature CheckFourk
	if (mobiApp == "iphone" && build > d.conf.Custom.FourkIOSBuild) || (mobiApp == "android" && build >= d.conf.Custom.FourkAndBuild) || (mobiApp == "ipad" && build >= d.conf.Custom.FourkIPadHDBuild) {
		return fourk
	}
	return false
}

func (d *Dao) Bubble(c context.Context, param *model.BubbleParam, mid int64) (*ott.ProjectionActivityReply, error) {
	req := &ott.ProjectionActivityReq{
		Mid:      mid,
		Aid:      param.Aid,
		Cid:      param.Cid,
		SeasonId: param.SeasonId,
		EpId:     param.EpId,
		MobiApp:  param.MobiApp,
		Build:    param.Build,
	}
	reply, err := d.ottCli.ProjectionActivity(c, req)
	if err != nil {
		return nil, errors.Wrapf(err, "Bubble req %v", req)
	}
	if reply == nil {
		return &ott.ProjectionActivityReply{Show: false}, nil
	}
	return reply, nil
}

func (d *Dao) BubbleSubmit(c context.Context, req *ott.ProjectionActivitySubmitReq) (*ott.ProjectionActivitySubmitReply, error) {
	return d.ottCli.ProjectionActivitySubmit(c, req)
}

func (d *Dao) ProjPageAct(c context.Context, params *model.ProjPageParam) (*ott.ProjPageActReply, error) {
	req := &ott.ProjPageActReq{
		PlayurlType: params.PlayurlType,
		Aid:         params.Aid,
		Cid:         params.Cid,
		EpId:        params.EpId,
		SeasonId:    params.SeasonId,
		Mid:         params.Mid,
		Channel:     params.Channel,
		Platform:    params.Platform,
		MobiApp:     params.MobiApp,
		Build:       params.Build,
	}
	reply, err := d.ottCli.ProjPageActivity(c, req)
	if err != nil {
		return nil, errors.Wrapf(err, "dao.ProjPageAct req %v", req)
	}
	return reply, nil
}

func (d *Dao) ProjActAll(c context.Context, params *model.ProjActAllParam) (*ott.ProjActivityAllReply, error) {
	req := &ott.ProjActivityAllReq{
		ActTypeBits: params.ActTypeBits,
		PlayurlType: params.PlayurlType,
		Aid:         params.Aid,
		Cid:         params.Cid,
		SeasonId:    params.SeasonId,
		EpId:        params.EpId,
		RoomId:      params.RoomId,
		PartitionId: params.PartitionId,
		Mid:         params.Mid,
		Channel:     params.Channel,
		Platform:    params.Platform,
		MobiApp:     params.MobiApp,
		Build:       params.Build,
		UserType:    params.NewUser,
	}
	reply, err := d.ottCli.ProjActivityAll(c, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
