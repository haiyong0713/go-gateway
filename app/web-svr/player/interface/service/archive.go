package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/web-svr/player/interface/model"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	ugcmdl "git.bilibili.co/bapis/bapis-go/account/service/ugcpay"
)

const (
	_maxLevel  = 6
	_hasUGCPay = 1
)

// View get view info
func (s *Service) View(c context.Context, aid int64) (view *model.View, err error) {
	var viewReply *arcmdl.ViewReply
	if viewReply, err = s.arcGRPC.View(c, &arcmdl.ViewRequest{Aid: aid}); err != nil {
		log.Error("View s.arcGRPCView3(%d) error(%v)", aid, err)
		return
	}
	if viewReply == nil || viewReply.Arc == nil || !viewReply.Arc.IsNormal() {
		log.Warn("View warn aid(%d) vi nil or state not allow", aid)
		err = ecode.NothingFound
		return
	}
	view = &model.View{Arc: viewReply.Arc, Bvid: s.avToBv(viewReply.Arc.Aid), Pages: viewReply.Pages}
	return
}

// Matsuri get matsuri info
func (s *Service) Matsuri(c context.Context, now time.Time) (view *model.View) {
	if now.Unix() < s.matTime.Unix() {
		return s.pastView
	}
	if s.matOn || len(s.matView.Pages) < 1 {
		return s.matView
	}
	view = new(model.View)
	*view = *s.matView
	view.Pages = view.Pages[0 : len(view.Pages)-1]
	return
}

// PageList many p video pages
func (s *Service) PageList(c context.Context, aid int64) ([]*arcmdl.Page, error) {
	reply, err := s.arcGRPC.View(c, &arcmdl.ViewRequest{Aid: aid})
	if err != nil {
		return nil, s.slbRetryCode(err)
	}
	if reply == nil || reply.Arc == nil || !reply.Arc.IsNormal() {
		return nil, ecode.NothingFound
	}
	rs := reply.Pages
	if reply.Arc.AttrVal(arcmdl.AttrBitSteinsGate) == arcmdl.AttrYes {
		reply, err := s.steinsGateGRPC.View(c, &api.ViewReq{Aid: aid})
		if err != nil {
			return nil, err
		}
		rs = []*arcmdl.Page{model.ArchivePage(reply.Page)}
	}
	return rs, nil
}

// VideoShot get archive video shot data
func (s *Service) VideoShot(c context.Context, mid, aid, cid int64, index bool, buvid string) (res *model.Videoshot, err error) {
	var (
		viewReply      *arcmdl.SteinsGateViewReply
		videoShotReply *arcmdl.VideoShotReply
	)
	if viewReply, err = s.arcGRPC.SteinsGateView(c, &arcmdl.SteinsGateViewRequest{Aid: aid}); err != nil {
		log.Error("VideoShot s.arcGRPC.View(%d) error(%v)", aid, err)
		return
	}
	if viewReply == nil || viewReply.Arc == nil || !viewReply.Arc.IsNormal() {
		log.Warn("VideoShot warn aid(%d) viewReply(%+v)", aid, viewReply)
		err = ecode.NothingFound
		return
	}
	if viewReply.Arc.Rights.UGCPay == _hasUGCPay {
		if mid == 0 {
			err = ecode.NothingFound
			return
		} else if mid != viewReply.Arc.Author.Mid {
			var relation *ugcmdl.AssetRelationResp
			if relation, err = s.ugcPayGRPC.AssetRelation(c, &ugcmdl.AssetRelationReq{Mid: mid, Oid: aid, Otype: _ugcPayOtypeArc}); err != nil {
				log.Error("Player AssetRelation mid:%d aid:%d error(%+v)", mid, aid, err)
				err = ecode.NothingFound
				return
			} else if relation.State != _relationPaid {
				err = ecode.NothingFound
				return
			}
		}
	}
	if cid == 0 { // steins-gate archive replaces the cid
		if len(viewReply.Pages) == 0 {
			err = ecode.NothingFound
			return
		}
		cid = viewReply.Pages[0].Cid
	}
	res = &model.Videoshot{}
	res.Index = make([]uint16, 0)
	if videoShotReply, err = s.arcGRPC.VideoShot(c, &arcmdl.VideoShotRequest{Aid: aid, Cid: cid}); err != nil {
		log.Error("s.arc.Videoshot2(%d,%d) err(%v)", aid, cid, err)
		return
	}
	res.VideoShot = videoShotReply.GetVs()
	if Hdvs := videoShotReply.GetHdVs(); Hdvs != nil {
		res.VideoShot = Hdvs
	}
	if index && res.PvData != "" {
		if pv, e := s.dao.PvData(c, res.PvData); e != nil {
			log.Error("s.dao.PvData(aid:%d,cid:%d) err(%+v)", aid, cid, e)
		} else if len(pv) > 0 {
			var (
				v   uint16
				pvs []uint16
				buf = bytes.NewReader(pv)
			)
			for {
				if e := binary.Read(buf, binary.BigEndian, &v); e != nil {
					if e != io.EOF {
						log.Warn("binary.Read pvdata(%s) err(%v)", res.PvData, e)
					}
					break
				}
				pvs = append(pvs, v)
			}
			res.Index = pvs
		}
	}
	s.fmtVideshot(res, aid)
	return
}

func (s *Service) fmtVideshot(res *model.Videoshot, aid int64) {
	if res.PvData != "" {
		res.PvData = strings.Replace(res.PvData, "http://", "//", 1)
	}
	for i, v := range res.Image {
		v = s.grayVideoShot(v, aid)
		res.Image[i] = strings.Replace(v, "http://", "//", 1)
	}
}

func (s *Service) grayVideoShot(img string, aid int64) string {
	group := aid % int64(s.c.GrayVideoShot.Group)
	if group < int64(s.c.GrayVideoShot.Gray) {
		u, err := url.Parse(img)
		if err != nil {
			log.Error("s.grayVideoShot error:%v", err)
			return img
		}
		if strings.Contains(u.Host, "boss") {
			u.Host = "bimp.hdslb.com"
			return u.Scheme + "://" + u.Host + u.Path
		}
	}
	return img
}

// PlayURLToken get playurl token
func (s *Service) PlayURLToken(c context.Context, mid, aid, cid int64) (res *model.PlayURLToken, err error) {
	var (
		arcReply    *arcmdl.ArcReply
		ui          *accmdl.CardReply
		owner, svip int
		vip         int32
	)
	if arcReply, err = s.arcGRPC.Arc(c, &arcmdl.ArcRequest{Aid: aid}); err != nil {
		log.Error("PlayURLToken s.arcGRPC.Arc(%d) error(%v)", aid, err)
		err = ecode.NothingFound
		return
	}
	if arcReply == nil || arcReply.Arc == nil || !arcReply.Arc.IsNormal() {
		log.Warn("PlayURLToken warn aid(%d) arcReply(%+v)", aid, arcReply)
		err = ecode.NothingFound
		return
	}
	if mid == arcReply.Arc.Author.Mid {
		owner = 1
	}
	if ui, err = s.accGRPC.Card3(c, &accmdl.MidReq{Mid: mid}); err != nil {
		log.Error("PlayURLToken s.accGRPC.Card3(%d) error(%v)", mid, err)
		err = ecode.AccessDenied
		return
	}
	if vip = ui.Card.Level; vip > _maxLevel {
		vip = _maxLevel
	}
	if ui.Card.Vip.Type != 0 && ui.Card.Vip.Status == 1 {
		svip = 1
	}
	res = &model.PlayURLToken{
		From:  "pc",
		Ts:    time.Now().Unix(),
		Aid:   aid,
		Bvid:  s.avToBv(aid),
		Cid:   cid,
		Mid:   mid,
		Owner: owner,
		VIP:   int(vip),
		SVIP:  svip,
	}
	params := url.Values{}
	params.Set("from", res.From)
	params.Set("ts", strconv.FormatInt(res.Ts, 10))
	params.Set("aid", strconv.FormatInt(res.Aid, 10))
	params.Set("cid", strconv.FormatInt(res.Cid, 10))
	params.Set("mid", strconv.FormatInt(res.Mid, 10))
	params.Set("vip", strconv.Itoa(res.VIP))
	params.Set("svip", strconv.Itoa(res.SVIP))
	params.Set("owner", strconv.Itoa(res.Owner))
	tmp := params.Encode()
	if strings.IndexByte(tmp, '+') > -1 {
		tmp = strings.Replace(tmp, "+", "%20", -1)
	}
	mh := md5.Sum([]byte(strings.ToLower(tmp) + s.c.PlayURLToken.Secret))
	res.Fcs = hex.EncodeToString(mh[:])
	res.Token = base64.StdEncoding.EncodeToString([]byte(tmp + "&fcs=" + res.Fcs))
	return
}
