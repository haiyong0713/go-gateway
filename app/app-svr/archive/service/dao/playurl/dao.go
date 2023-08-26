package playurl

import (
	"context"
	"fmt"
	"hash/crc32"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/archive/service/conf"
	"go-gateway/app/app-svr/archive/service/model"

	steampunkgrpc "git.bilibili.co/bapis/bapis-go/pcdn/steampunk"
	batch "git.bilibili.co/bapis/bapis-go/video/vod/playurlugcbatch"
	volume "git.bilibili.co/bapis/bapis-go/video/vod/playurlvolume"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	MaxGray = 1000
)

// Dao dao
type Dao struct {
	c               *conf.Config
	batchClient     batch.PlayurlServiceClient
	volumeClient    volume.PlayurlServiceClient
	steampunkClient steampunkgrpc.PcdnClient
}

// New init mysql db
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c: c,
	}
	var (
		err        error
		opts       []grpc.DialOption
		WindowSize = int32(65535000)
	)
	opts = append(opts, grpc.WithInitialWindowSize(WindowSize))
	opts = append(opts, grpc.WithInitialConnWindowSize(WindowSize))
	if dao.batchClient, err = batch.NewClient(c.PlayurlClient, opts...); err != nil {
		panic(err)
	}
	if dao.volumeClient, err = volume.NewClientPlayurlService(c.VolumeClient, opts...); err != nil {
		panic(err)
	}
	if dao.steampunkClient, err = steampunkgrpc.NewClientPcdn(c.SteampunkClient); err != nil {
		panic(err)
	}
	return
}

// PlayurlBatch .
func (d *Dao) PlayurlBatch(c context.Context, cidArr []*batch.RequestVideoItem, batchArg *api.BatchPlayArg, backupNum int, v2 bool) (res map[uint64]*batch.ResponseItem, err error) {
	var req *batch.RequestMsg
	// 该判断为了兼容6.15 ios默认force_host传1 导致起播时间增大问题，预计202103删除
	if batchArg.MobiApp == "iphone" && batchArg.ForceHost == 1 {
		batchArg.ForceHost = 0
	}
	//4k需求由于android老版本缓存时缺少fnval和fnver，不出4k，所以增加版本控制
	fourk := int64(0)
	if (batchArg.MobiApp == "iphone" && batchArg.Build > d.c.Custom.FourkIOSBuild) || (batchArg.MobiApp == "android" && batchArg.Build >= d.c.Custom.FourkAndBuild) || (batchArg.MobiApp == "ipad" && batchArg.Build >= d.c.Custom.FourkIPadHDBuild) {
		fourk = batchArg.Fourk
	}
	req = &batch.RequestMsg{
		Keys:      cidArr,
		Qn:        uint32(batchArg.Qn),
		Uip:       batchArg.Ip,
		Platform:  batchArg.MobiApp,
		Fnver:     uint32(batchArg.Fnver),
		Fnval:     uint32(batchArg.Fnval),
		Mid:       uint64(batchArg.Mid),
		ForceHost: uint32(batchArg.ForceHost),
		Fourk:     fourk == 1,
		FlvProj:   d.getFlvProject(batchArg.Buvid),
		BackupNum: uint32(backupNum),
		NetType:   batch.NetworkType(batchArg.NetType),
		TfType:    batch.TFType(batchArg.TfType),
	}
	// 将版本号大于等于66000100的粉ipad的platform改为ipad
	if batchArg.MobiApp == "iphone" && batchArg.Device == "pad" && batchArg.Build >= 66000100 && d.ClarityGrayControl(batchArg.Mid, batchArg.Buvid) {
		req.Platform = "ipad"
	}

	if batchArg.From == model.PlayurlFromStory {
		req.ReqSource = batch.RequestSource_STORY
	}
	var batchRes *batch.ResponseMsg
	if v2 { // 返回所有路所有路清晰度及该路是否二压
		batchRes, err = d.batchClient.Playurl2(c, req)
	} else {
		batchRes, err = d.batchClient.ProtobufPlayurl(c, req)
	}
	if err != nil {
		return nil, err
	}
	if batchRes == nil {
		err = errors.New("res is nil")
		return nil, err
	}
	if int(batchRes.Code) != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(int(batchRes.Code)), fmt.Sprintf("ProtobufPlayurl errcode arg(%v)", req))
		return nil, err
	}
	return batchRes.Data, nil
}

func (d *Dao) getFlvProject(buvid string) bool {
	return crc32.ChecksumIEEE([]byte(buvid+"_project_flv"))%100 < d.c.Custom.FlvProjectGray
}

// PlayurlVolume .
func (d *Dao) PlayurlVolume(c context.Context, cids []uint64, batchArg *api.BatchPlayArg) (map[uint64]*volume.VolumeItem, error) {
	volInfo := make(map[uint64]*volume.VolumeItem, len(cids))
	if len(cids) == 0 {
		return volInfo, nil
	}

	req := &volume.RequestMsg{
		Cids: cids,
		Uip:  batchArg.Ip,
		Mid:  uint64(batchArg.Mid),
	}
	res, err := d.volumeClient.Volume(c, req)
	if err != nil {
		return nil, err
	}
	if res == nil {
		err = errors.New("volumeClient.Volume res is nil")
		return nil, err
	}
	if int(res.Code) != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(int(res.Code)), fmt.Sprintf("volumeClient.Volume errcode(%v) arg(%v)", res.Code, req))
		return nil, err
	}
	//对于measured_i  <= -9的稿件不进行处理
	result := make(map[uint64]*volume.VolumeItem, len(res.Data))
	for i, re := range res.Data {
		if re.MeasuredI <= -9 {
			continue
		}
		result[i] = re
	}
	return result, nil
}

// ClarityGrayControl 白名单和灰度控制
func (d *Dao) ClarityGrayControl(mid int64, buvid string) bool {
	// 白名单
	_, ok := d.c.IpadClarityGrayControl.Mid[strconv.FormatInt(mid, 10)]
	// 灰度控制
	group := crc32.ChecksumIEEE([]byte(buvid)) % MaxGray
	return ok || group < uint32(d.c.IpadClarityGrayControl.Gray)
}

func (d *Dao) BatchGetPcdnUrl(ctx context.Context, cids []uint64, arg *api.BatchPlayArg) (map[uint64]*steampunkgrpc.CidResources, error) {
	req := &steampunkgrpc.BatchPlayRequest{
		Cid:      cids,
		Mid:      uint64(arg.Mid),
		Qn:       uint32(arg.Qn),
		Platform: arg.MobiApp,
		Uip:      metadata.String(ctx, metadata.RemoteIP),
	}
	res, err := d.steampunkClient.GetUrlsByCids(ctx, req)
	if err != nil {
		return nil, err
	}
	if res == nil {
		err = errors.New("res is nil")
		return nil, err
	}
	if int(res.Code) != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(int(res.Code)), fmt.Sprintf("d.steampunkClient.GetUrlsByCids req(%v)", req))
		return nil, err
	}
	return res.Data, nil
}
