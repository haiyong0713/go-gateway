package archive

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-feed/admin/conf"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/util"

	"git.bilibili.co/bapis/bapis-go/archive/service"

	flowCtrlGrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
)

const (
	_maxAids     = 100
	_arcBanURL   = "/va/archive/attrs"
	_arcAuditURL = "/videoup/view"
)

// 流量管控业务类型
const (
	// FlowCtrlBizArchive 稿件
	FlowCtrlBizArchive = 1
	// FlowCtrlBizArticle 专栏
	FlowCtrlBizArticle = 2
)

// 流量管控禁止项字段
// https://info.bilibili.co/pages/viewpage.action?pageId=150125941#grpc&http%E6%8E%A5%E5%8F%A3%E6%96%87%E6%A1%A3-%E5%B1%9E%E6%80%A7%E4%BD%8D%E7%A6%81%E6%AD%A2%E9%A1%B9%E8%BF%81%E7%A7%BB%E6%8E%92%E6%9F%A5%E8%AF%B4%E6%98%8E%EF%BC%9A
const (
	// FlowCtrlKeyNoSearch 搜索禁止
	FlowCtrlKeyNoSearch = "53"
)

// Arc get archive.
func (d *Dao) Arc(c context.Context, aid int64) (a *api.Arc, err error) {
	var (
		arg   = &api.ArcRequest{Aid: aid}
		reply *api.ArcReply
	)
	if reply, err = d.arcClient.Arc(c, arg); err != nil {
		log.Error("d.arcRPC.Archive3(%v) error(%+v)", arg, err)
		return
	}
	a = reply.Arc
	return
}

// Arcs gets archives
func (d *Dao) Arcs(c context.Context, aids []int64) (res map[int64]*api.Arc, err error) {
	if len(aids) == 0 {
		return
	}
	var (
		arg   = &api.ArcsRequest{Aids: aids}
		reply *api.ArcsReply
	)
	if reply, err = d.arcClient.Arcs(c, arg); err != nil {
		log.Error("d.arcRPC.Archive3(%v) error(%+v)", arg, err)
		return
	}
	res = reply.Arcs
	return
}

// Arcs gets archives
func (d *Dao) AidsToCids(c context.Context, aids []int64) (cids []int64, err error) {
	var (
		arg   = &api.ViewsRequest{Aids: aids}
		reply *api.ViewsReply
	)
	if reply, err = d.arcClient.Views(c, arg); err != nil || reply == nil || reply.Views == nil {
		log.Error("d.arcRPC.Archive3(%v) error(%+v)", arg, err)
		return
	}
	for _, v := range reply.Views {
		for _, p := range v.Pages {
			if p != nil && p.Cid != 0 {
				cids = append(cids, p.Cid)
			}
		}
	}
	return
}

func (d *Dao) ArcsWithPage(c context.Context, aids []int64) (map[int64]*api.Arc, error) {
	res := make(map[int64]*api.Arc)
	pag := len(aids)/_maxAids + 1
	for i := 0; i < pag; i++ {
		maxIndex := (i + 1) * _maxAids
		if maxIndex > len(aids) {
			maxIndex = len(aids)
		}
		aidTemp := aids[i*_maxAids : maxIndex]
		arcsTemp, err := d.Arcs(c, aidTemp)
		if err != nil {
			return nil, err
		}
		if len(arcsTemp) > 0 {
			for k, v := range arcsTemp {
				if _, ok := res[k]; !ok {
					res[k] = v
				}
			}
		}
	}
	return res, nil
}

// Types def.
func (d *Dao) Types(c context.Context) (result map[int32]*api.Tp, err error) {
	var response *api.TypesReply
	if response, err = d.arcClient.Types(c, &api.NoArgRequest{}); err != nil {
		return
	}
	result = response.Types
	return
}

// FlowJudge picks the forbidden arcs
func (d *Dao) FlowJudge(c context.Context, aids []int64, flowCtrlConf *conf.FlowCtrl) (noHotAids map[int64]struct{}, hotDownAids map[int64]struct{}, err error) {
	if flowCtrlConf == nil {
		return noHotAids, hotDownAids, ecode.Error(-500, "FlowCtr 配置错误")
	}

	var flowResp *flowCtrlGrpc.FlowCtlInfosReply
	noHotAids = make(map[int64]struct{})
	hotDownAids = make(map[int64]struct{})

	for {
		var oids []int64

		if len(aids) == 0 {
			break
		} else if len(aids) > flowCtrlConf.OidLength {
			oids = aids[0:flowCtrlConf.OidLength]
			aids = aids[flowCtrlConf.OidLength:]
		} else {
			oids = aids
			aids = []int64{}
		}

		now := time.Now().Unix()
		ts := strconv.FormatInt(time.Now().Unix(), 10)

		params := url.Values{}
		params.Set("source", flowCtrlConf.Source)
		params.Set("oids", xstr.JoinInts(oids)) //批量
		params.Set("business_id", "1")
		params.Set("ts", ts)

		req := &flowCtrlGrpc.FlowCtlInfosReq{
			Oids:       oids,
			BusinessId: 1,
			Source:     flowCtrlConf.Source,
			Ts:         now,
			Sign:       getSign(params, flowCtrlConf.Secret),
		}

		if flowResp, err = d.flowControlClient.Infos(c, req); err != nil {
			log.Error("s.creativeClient.FlowJudge error(%v)", err)
			return
		}

		for oid, fbItems := range flowResp.ForbiddenItemMap {
			for _, item := range fbItems.ForbiddenItems {
				if item.Key == "nohot" && item.Value == 1 {
					noHotAids[oid] = struct{}{}
				}
				if item.Key == "hot_down" && item.Value == 1 {
					hotDownAids[oid] = struct{}{}
				}
			}
		}
	}

	return
}

// ArchiveSearchBan
// 2022M3W1: 安全性考虑，使用 ArchiveSearchBan2 接口
func (d *Dao) ArchiveSearchBan(c context.Context, aid int64) (err error) {
	params := url.Values{}
	params.Add("aid", strconv.FormatInt(aid, 10))
	var res struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    *struct {
			ArcAttr *struct {
				Nosearch *struct {
					Val bool `json:"val"`
				} `json:"nosearch"`
			} `json:"arc_attr"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.archiveBanURL, "", params, &res); err != nil {
		log.Error("ArchiveSearchBan Req(%v) error(%v)", aid, err)
		return fmt.Errorf(util.ErrorNetFmts, util.ErrorNet, d.archiveBanURL+"?"+params.Encode(), err.Error())
	}
	if res.Code != ecode.OK.Code() {
		return fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.userFeed.Archive, d.archiveBanURL+"?"+params.Encode())
	}
	if res.Data == nil || res.Data.ArcAttr == nil || res.Data.ArcAttr.Nosearch == nil {
		return fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.userFeed.Archive, d.archiveBanURL+"?"+params.Encode())
	}
	if res.Data.ArcAttr.Nosearch.Val {
		return fmt.Errorf("该稿件（%d）被禁止在搜索中搜索，无法提交，如需提交，请联系审核，校验或修改稿件限制", aid)
	}
	return
}

// ArchiveSearchBan2 .
func (d *Dao) ArchiveSearchBan2(c context.Context, aid int64) (err error) {
	if d.feedFlowCtrlConf == nil {
		return ecode.Error(ecode.RequestErr, "FlowCtrl 配置错误")
	}
	now := time.Now().Unix()
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	params := url.Values{}

	params.Set("source", d.feedFlowCtrlConf.Source)
	params.Set("oid", fmt.Sprintf("%d", aid))
	params.Set("business_id", "1")
	params.Set("ts", ts)

	req := &flowCtrlGrpc.FlowCtlInfoReq{
		Oid:        aid,
		BusinessId: FlowCtrlBizArchive,
		Source:     d.feedFlowCtrlConf.Source,
		Ts:         now,
		Sign:       getSign(params, d.feedFlowCtrlConf.Secret),
	}
	var resp *flowCtrlGrpc.FlowCtlInfoV2Reply
	if resp, err = d.flowControlClient.InfoV2(c, req); err != nil {
		log.Error("s.flowControlClient.InfoV2 error aid(%v) (%v)", aid, err)
		return
	}
	for _, item := range resp.Items {
		if item.Key == FlowCtrlKeyNoSearch && item.Value == 1 {
			return ecode.Errorf(ecode.RequestErr,
				"该稿件（%d）被禁止在搜索中搜索，无法提交，如需提交，请联系审核，校验或修改稿件限制", aid)
		}
	}
	return
}

// ArchiveAudit
func (d *Dao) ArchiveAudit(c context.Context, aid int64) (*common.Archive, error) {
	params := url.Values{}
	params.Add("aid", strconv.FormatInt(aid, 10))
	res := &common.ArchiveResult{}
	err := d.client.Get(c, d.archiveAuditURL, "", params, &res)
	if err != nil {
		log.Error("ArchiveSearchBan Req(%v) error(%v)", aid, err)
		return nil, fmt.Errorf(util.ErrorNetFmts, util.ErrorNet, d.archiveAuditURL+"?"+params.Encode(), err.Error())
	}
	if res.Code != ecode.OK.Code() {
		return nil, fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.userFeed.Archive, d.archiveAuditURL+"?"+params.Encode())
	}
	if res.Data == nil || res.Data.Archive == nil {
		return nil, fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.userFeed.Archive, d.archiveAuditURL+"?"+params.Encode())
	}
	return res.Data.Archive, nil
}

// get sign for flowCtrl
func getSign(params url.Values, secret string) string {
	tmp := params.Encode()
	if strings.IndexByte(tmp, '+') > -1 {
		tmp = strings.Replace(tmp, "+", "%20", -1)
	}

	var buf bytes.Buffer
	buf.WriteString(tmp)
	buf.WriteString(secret)
	mh := md5.Sum(buf.Bytes())
	return hex.EncodeToString(mh[:])
}
