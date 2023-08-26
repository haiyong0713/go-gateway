package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"hash/crc32"
	"html/template"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/log/infoc.v2"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	playsvcgrpc "go-gateway/app/app-svr/playurl/service/api/v2"
	v1 "go-gateway/app/app-svr/resource/service/api/v1"
	resmdl "go-gateway/app/app-svr/resource/service/model"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/web-svr/player/interface/model"
	mecode "go-gateway/ecode"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	memmdl "git.bilibili.co/bapis/bapis-go/account/service/member"
	ugcmdl "git.bilibili.co/bapis/bapis-go/account/service/ugcpay"
	assitmdl "git.bilibili.co/bapis/bapis-go/assist/service"
	pugvmdl "git.bilibili.co/bapis/bapis-go/cheese/service/auth"
	ansmdl "git.bilibili.co/bapis/bapis-go/community/interface/answer"
	dmmdl "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	hismdl "git.bilibili.co/bapis/bapis-go/community/interface/history"
	appcfgmdl "git.bilibili.co/bapis/bapis-go/community/service/appconfig"
	locmdl "git.bilibili.co/bapis/bapis-go/community/service/location"
	videoupmdl "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

const (
	_content = `<a href="%s" target="_blank"><font color="#FFFFFF">%s</font></a>`
	_china   = "中国"
	_local   = "局域网"

	_accBanNor      = 0 // no block
	_accBanSta      = 1 // block MSpacesta
	_accBlockAll    = 1
	_accBlockLimit  = 2
	_dmMaskPlatWeb  = 0
	_mockBlockTime  = 100
	_previewToast   = "为创作付费，购买观看完整视频|购买观看"
	_member         = 10000
	_notAnswer      = 1
	_guideTypeAtten = 1
)

var (
	_copyRightMap = map[int32]string{
		0: "Nnknown",
		1: "Original",
		2: "Copy",
	}

	// if typeid in this xml add bottom = 1
	_bottomMap = map[int32]struct{}{
		// 番剧
		33:  {},
		32:  {},
		153: {},
		// 电影
		82:  {},
		85:  {},
		145: {},
		146: {},
		147: {},
		83:  {},
		// note 电视剧存在三级分区
		15:  {},
		34:  {},
		86:  {},
		128: {},
		// 三级分区
		110: {},
		111: {},
		112: {},
		113: {},
		87:  {},
		88:  {},
		89:  {},
		90:  {},
		91:  {},
		92:  {},
		73:  {},
	}
)

func (s *Service) PlayerCardClick(ctx context.Context, arg *model.PlayerCardClickArg, mid int64) error {
	req := &appcfgmdl.ClickPlayerCardReq{
		Id:      arg.ID,
		Mid:     mid,
		OidType: appcfgmdl.OidType(arg.OidType),
		Oid:     arg.Oid,
		Pid:     arg.Pid,
		Common:  &appcfgmdl.CommonParam{Platform: _platformPC, Buvid: arg.Buvid},
		Action:  arg.Action,
	}
	_, err := s.appConfigGRPC.ClickPlayerCard(ctx, req)
	if err != nil {
		log.Error("PlayerCardClick req:%+v error:%+v", req, err)
		return err
	}
	return nil
}

// Carousel return carousel items.
func (s *Service) Carousel(c context.Context) (items []*model.Item, err error) {
	items = s.caItems
	return
}

// Player return player info.
func (s *Service) Player(c context.Context, mid, aid int64, arg *model.PlayerArg, cdnIP, refer, innerSign string, now time.Time) (res []byte, err error) {
	var (
		ip     = metadata.String(c, metadata.RemoteIP)
		vi     *arcmdl.ViewReply
		cuPage *arcmdl.Page
		pi     = &model.Player{
			IP:           ip,
			Login:        mid > 0,
			Time:         now.Unix(),
			ZoneIP:       cdnIP,
			Upermission:  "1000,1001",
			PreviewToast: _previewToast,
		}
		withU bool
	)
	if vi, err = s.view(c, aid); err != nil {
		log.Error("Player s.arcGRPCView3(%d) error(%v)", aid, err)
		return
	} else if vi == nil || vi.Arc == nil || !vi.Arc.IsNormal() {
		log.Warn("aid(%d) vi nil or state not allow", aid)
		err = ecode.NothingFound
		return
	}
	for _, page := range vi.Pages {
		if arg.Cid == page.Cid {
			cuPage = page
			break
		}
	}
	if cuPage == nil {
		log.Warn("cuPage is nil aid(%d) cid(%d) refer(%s)", aid, arg.Cid, refer)
	}
	if err = s.fillArc(c, mid, arg.Cid, arg.GraphVersion, arg.SeasonID, arg.EpID, pi, vi, cuPage, arg.Buvid, refer, innerSign, ip, now); err != nil {
		err = ecode.NothingFound
		return
	}
	withU = s.fillAcc(c, pi, vi, mid, arg.Cid, now)
	// bvid
	pi.Bvid = s.avToBv(pi.Aid)
	// template
	var doc = bytes.NewBuffer(nil)
	if withU {
		_ = s.tWithU.Execute(doc, pi)
	} else {
		_ = s.tNoU.Execute(doc, pi)
	}
	if s.params != "" {
		doc.WriteString(s.params)
	}
	res = doc.Bytes()
	return
}

// nolint:gocognit,gomnd
func (s *Service) fillAcc(c context.Context, pi *model.Player, vi *arcmdl.ViewReply, mid, cid int64, now time.Time) (withU bool) {
	if mid == 0 {
		return
	}
	var (
		proReply *accmdl.ProfileStatReply
		pro      *hismdl.ProgressReply
		err      error
	)
	if proReply, err = s.accGRPC.ProfileWithStat3(c, &accmdl.MidReq{Mid: mid}); err != nil {
		log.Error("fillAcc s.accGRPC.ProfileWithStat3(%d) error(%v)", mid, err)
		return
	}
	if proReply == nil {
		return
	}
	withU = true
	var nameBu = bytes.NewBuffer(nil)
	if err = xml.EscapeText(nameBu, []byte(proReply.Profile.Name)); err != nil {
		log.Error("xml.EscapeText(%s) error(%v)", proReply.Profile.Name, err)
	} else {
		pi.Name = nameBu.String()
	}
	pi.User = proReply.Profile.Mid
	pi.UserHash = midCrc(proReply.Profile.Mid)
	pi.Money = fmt.Sprintf("%.2f", proReply.Coins)
	pi.Face = strings.Replace(proReply.Profile.Face, "http://", "//", 1)
	var bs []byte
	if bs, err = json.Marshal(proReply.LevelInfo); err != nil {
		log.Error("json.Marshal(%v) error(%v)", proReply.LevelInfo, err)
	} else {
		// nolint:gosec
		pi.LevelInfo = template.HTML(bs)
	}
	vip := model.VIPInfo{Type: proReply.Profile.Vip.Type, DueDate: proReply.Profile.Vip.DueDate, VipStatus: proReply.Profile.Vip.Status}
	if bs, err = json.Marshal(vip); err != nil {
		log.Error("json.Marshal(%v) error(%v)", vip, err)
	} else {
		// nolint:gosec
		pi.Vip = template.HTML(bs)
	}
	off := &model.Official{Type: -1}
	if proReply.Profile.Official.Role != 0 {
		if proReply.Profile.Official.Role <= 2 {
			off.Type = 0
		} else {
			off.Type = 1
		}
		off.Desc = proReply.Profile.Official.Title
	}
	if bs, err = json.Marshal(off); err != nil {
		log.Error("json.Marshal(%v) error(%v)", off, err)
	} else {
		// nolint:gosec
		pi.OfficialVerify = template.HTML(bs)
	}
	group := errgroup.WithContext(c)
	if vi.Arc != nil {
		pi.Upermission = userPermission(proReply)
		// NOTE: if vInfo==nil, no admin
		if mid == vi.Arc.Author.Mid {
			pi.IsAdmin = true
		}
		group.Go(func(ctx context.Context) error {
			arg := &hismdl.ProgressReq{Mid: mid, Aids: []int64{vi.Arc.Aid}}
			if pro, err = s.hisGRPC.Progress(ctx, arg); err != nil {
				log.Error("fillAcc s.his.Progress(%d,%d) error(%v)", mid, vi.Arc.Aid, err)
			} else if progress, ok := pro.Res[vi.Arc.Aid]; ok && progress != nil && progress.Cid > 0 && progress.Cid == cid {
				if progress.Pro >= 0 {
					pi.LastPlayTime = 1000 * progress.Pro
					pi.LastCid = progress.Cid
				} else if len(vi.Pages) != 0 {
					for _, page := range vi.Pages {
						if page.Cid == progress.Cid {
							pi.LastPlayTime = 1000 * progress.Pro
							pi.LastCid = progress.Cid
							break
						}
					}
				}
			}
			return nil
		})
	}
	if proReply.Profile.Rank < _member {
		group.Go(func(ctx context.Context) error {
			if statusReply, e := s.ansGRPC.Status(ctx, &ansmdl.StatusReq{Mid: mid}); e != nil {
				log.Error("s.ansRPC.Status(%d) error(%v)", mid, e)
				pi.AnswerStatus = _notAnswer
			} else if statusReply != nil {
				pi.AnswerStatus = statusReply.Status.Status
			}
			return nil
		})
	}
	if s.c.Rule.NoAssistMid != vi.Arc.Author.Mid {
		group.Go(func(ctx context.Context) error {
			if assist, e := s.assistGRPC.Assist(ctx, &assitmdl.AssistReq{Mid: vi.Arc.Author.Mid, AssistMid: proReply.Profile.Mid, Tp: 2}); e != nil {
				log.Error("fillAcc s.ass.Assist(%d,%d) error(%v)", vi.Arc.Author.Mid, proReply.Profile.Mid, e)
			} else {
				pi.Role = strconv.FormatInt(assist.Ar.Assist, 10)
			}
			return nil
		})
	}
	if proReply.Profile.Silence == _accBanSta {
		group.Go(func(ctx context.Context) error {
			blockTime, e := s.memberGRPC.BlockInfo(ctx, &memmdl.MemberMidReq{Mid: mid})
			if e != nil {
				log.Error("fillAcc s.memberGRPC.BlockInfo(%d) error(%v)", mid, e)
				return nil
			}
			switch blockTime.GetBlockStatus() {
			case _accBlockLimit: //限时封禁
				pi.BlockTime = blockTime.EndTime - now.Unix()
				if pi.BlockTime < 0 {
					pi.BlockTime = 0
				}
			case _accBlockAll: //永久封禁
				pi.BlockTime = _mockBlockTime
			default:
				pi.BlockTime = 0
			}
			return nil
		})
	}
	if pi.IsPayPreview {
		if mid == vi.Arc.Author.Mid {
			pi.IsPayPreview = false
		} else {
			group.Go(func(ctx context.Context) error {
				var relation *ugcmdl.AssetRelationResp
				if relation, err = s.ugcPayGRPC.AssetRelation(ctx, &ugcmdl.AssetRelationReq{Mid: mid, Oid: vi.Arc.Aid, Otype: _ugcPayOtypeArc}); err != nil {
					log.Error("Player s.ugcPayGRPC.AssetRelation mid:%d aid:%d error(%+v)", mid, vi.Arc.Aid, err)
				} else if relation.State == _relationPaid {
					pi.IsPayPreview = false
				}
				return nil
			})
		}
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}

func (s *Service) fillArc(c context.Context, mid, cid, graphVersion, seasonID, epID int64, pi *model.Player, vi *arcmdl.ViewReply, page *arcmdl.Page, buvid, refer, innerSign, ip string, now time.Time) (err error) {
	// 稿件和其弹幕信息
	pi.Aid = vi.Arc.Aid
	pi.Typeid = vi.Arc.TypeID
	if page != nil {
		if page.From != "sina" {
			pi.Vtype = page.From
		} else {
			pi.Vtype = ""
		}
		pi.Maxlimit = dmLimit(page.Duration)
		pi.Chatid = page.Cid
		pi.Oriurl = oriURL(page.From, page.Vid)
		pi.Pid = int64(page.Page)
	} else {
		pi.Chatid = cid
		pi.Maxlimit = 1500
		pi.Pid = 1
	}
	pi.Arctype = _copyRightMap[vi.Arc.Copyright]
	pi.SuggestComment = false
	group := errgroup.WithContext(c)
	pi.OnlineCount = 1
	group.Go(func(ctx context.Context) error {
		if onlineCount, e := s.dao.OnlineCount(ctx, pi.Aid, cid); e == nil && onlineCount > 1 {
			pi.OnlineCount = onlineCount
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		_, pi.MaskNew = s.dmMask(ctx, cid)
		return nil
	})
	group.Go(func(ctx context.Context) error {
		_, pi.Subtitle = s.dmSubtitle(ctx, pi.Aid, cid)
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if ipInfo, e := s.locGRPC.Info(ctx, &locmdl.InfoReq{Addr: ip}); e != nil {
			log.Error("fillArc s.locGRPC.Info(%s) error(%v)", ip, e)
		} else if ipInfo != nil {
			pi.Zoneid = ipInfo.ZoneId
			pi.Country = ipInfo.Country
			pi.Acceptaccel = ipInfo.Country != _china && ipInfo.Country != _local
			pi.Cache = ipInfo.Country != _china && ipInfo.Country != _local
		}
		return nil
	})
	// 高能看点和章节
	group.Go(func(ctx context.Context) error {
		_, pi.ViewPoints = s.viewPoints(ctx, pi.Aid, cid)
		return nil
	})
	group.Go(func(ctx context.Context) error {
		_, pi.PlayerIcon = s.tagPlayerIcon(ctx, pi.Aid, seasonID, pi.Typeid, mid)
		return nil
	})
	if vi.Arc.AttrVal(arcmdl.AttrBitSteinsGate) == arcmdl.AttrYes {
		if _, ok := s.steinGuideCids[cid]; !ok {
			group.Go(func(ctx context.Context) error {
				var e error
				if _, pi.Interaction, e = s.interaction(ctx, pi.Aid, cid, mid, graphVersion, buvid); e != nil {
					log.Error("s.interaction(%d,%d) error(%v)", pi.Aid, mid, e)
				}
				return e
			})
		}
	}
	if vi.Arc.AttrVal(arcmdl.AttrBitIsPUGVPay) == arcmdl.AttrYes {
		group.Go(func(ctx context.Context) error {
			pugvStatus, e := s.pugvGRPC.SeasonPlayStatus(ctx, &pugvmdl.SeasonPlayStatusReq{Aid: pi.Aid, Mid: mid})
			if e != nil {
				log.Error("s.pugvGRPC.SeasonPlayStatus aid (%d),mid (%d) error(%v)", pi.Aid, mid, e)
				return e
			}
			pi.PugvPayStatus = pugvStatus.PayStatus
			pi.PugvWatchStatus = pugvStatus.WatchStatus
			pi.PugvSeasonStatus = pugvStatus.SeasonStatus
			return nil
		})
	}
	group.Go(func(ctx context.Context) error {
		_, pi.PcdnLoader = s.pcdnLoader(ctx, cid, mid, refer, innerSign)
		return nil
	})
	group.Go(func(ctx context.Context) error {
		_, _, _, pi.GuideAttention, pi.JumpCard, pi.OperationCard = s.playerCards(ctx, pi.Aid, cid, epID, mid)
		return nil
	})
	if err = group.Wait(); err != nil {
		return
	}
	pi.Duration = formatDuration(vi.Arc.Duration)
	pi.AllowBp = vi.Arc.AttrVal(arcmdl.AttrBitAllowBp) == 1
	if _, ok := _bottomMap[vi.Arc.TypeID]; ok {
		pi.Bottom = 1
	}
	pi.Acceptguest = false
	if s.BrBegin.Unix() <= now.Unix() && now.Unix() <= s.BrEnd.Unix() {
		pi.BrTCP = s.c.Broadcast.TCPAddr
		pi.BrWs = s.c.Broadcast.WsAddr
		pi.BrWss = s.c.Broadcast.WssAddr
	}
	for index, pa := range vi.Pages {
		if pa != nil && cid == pa.Cid && index+1 < len(vi.Pages) {
			pi.HasNext = 1
		}
	}
	if vi.Arc.AttrVal(arcmdl.AttrBitUGCPay) == arcmdl.AttrYes && vi.Arc.AttrVal(arcmdl.AttrBitUGCPayPreview) == arcmdl.AttrYes {
		pi.IsPayPreview = true
	}
	_, pi.Options = s.playerOptions(vi.Arc)
	return
}

func (s *Service) playerOptions(arc *arcmdl.Arc) (*model.Option, template.HTML) {
	option := &model.Option{}
	if arc.AttrValV2(arcmdl.AttrBitV2Is360) == arcmdl.AttrYes {
		option.Is360 = true
	}
	const isLimitFree = 1
	if value, ok := s.limitFreeMap[arc.Aid]; ok && value.LimitFree == isLimitFree {
		option.WithoutVip = true
	}
	bs, _ := json.Marshal(option)
	// nolint:gosec
	return option, template.HTML(bs)
}

func isAdmin(uRank int32) bool {
	// 32000 -> admin
	// 31300 -> 评论管理员
	if uRank == 31300 || uRank == 32000 {
		return true
	}
	return false
}

func userPermission(u *accmdl.ProfileStatReply) (permission string) {
	if u.Profile.Silence == _accBanNor || isAdmin(u.Profile.Rank) {
		permission = strings.Join([]string{strconv.FormatInt(int64(u.Profile.Rank), 10), "1001"}, ",")
	} else {
		permission = "0"
	}
	// if a.AttrVal(arcmdl.AttrBitNoMission) == 0 && a.Author.Mid == u.Mid {
	// 	permission = strings.Join([]string{permission, "20000"}, ",")
	// }
	return
}

// nolint:gomnd
func oriURL(dmType, dmIndex string) (url string) {
	switch dmType {
	case "sina":
		url = "http://p.you.video.sina.com.cn/swf/bokePlayer20131203_V4_1_42_33.swf?vid=" + dmIndex
	case "youku":
		url = "http://v.youku.com/v_show/id_" + dmIndex + ".html"
	case "qq":
		if len(dmIndex) >= 3 {
			url = "http://v.qq.com/page/" + dmIndex[0:1] + "/" + dmIndex[1:2] + "/" + dmIndex[2:3] + "/" + dmIndex + ".html"
		}
	default:
		url = ""
	}
	return
}

// nolint:gomnd
func formatDuration(duration int64) (du string) {
	if duration == 0 {
		du = "00:00"
	} else {
		var duFen, duMiao string
		duFen = strconv.Itoa(int(duration / 60))
		if int(duration%60) < 10 {
			duMiao = "0" + strconv.Itoa(int(duration%60))
		} else {
			duMiao = strconv.Itoa(int(duration % 60))
		}
		du = duFen + ":" + duMiao
	}
	return
}

func midCrc(mid int64) string {
	midStr := strconv.FormatInt(mid, 10)
	return fmt.Sprintf("%08x", crc32.ChecksumIEEE([]byte(midStr)))
}

// nolint:gomnd
func dmLimit(duration int64) (limit int64) {
	switch {
	case duration > 3600:
		limit = 8000
	case duration > 2400:
		limit = 6000
	case duration > 900:
		limit = 3000
	case duration > 600:
		limit = 1500
	case duration > 150:
		limit = 1000
	case duration > 60:
		limit = 500
	case duration > 30:
		limit = 300
	case duration <= 30:
		limit = 100
	default:
		limit = 1500
	}
	return
}

func (s *Service) dmMask(c context.Context, cid int64) (*model.PlayerDmMask, template.HTML) {
	if reply, err := s.dmGRPC.Mask(c, &dmmdl.MaskReq{Cid: cid, Plat: _dmMaskPlatWeb}); err != nil {
		log.Error("dmMask s.dm2.MaskList cid(%d) error(%v)", cid, err)
	} else if reply != nil && reply.Mask != nil && reply.Mask.MaskUrl != "" {
		reply.Mask.MaskUrl = strings.Replace(reply.Mask.MaskUrl, "http://", "//", 1)
		if bs, err := json.Marshal(reply.Mask); err != nil {
			log.Error("dmMask json.Marshal(%+v) error(%v)", reply.Mask, err)
		} else {
			// nolint:gosec
			tmp := template.HTML(bs)
			return &model.PlayerDmMask{
				Cid:     reply.Mask.Cid,
				Plat:    reply.Mask.Plat,
				Fps:     reply.Mask.Fps,
				Time:    reply.Mask.Time,
				MaskUrl: reply.Mask.MaskUrl,
			}, tmp
		}
	}
	return nil, ""
}

func (s *Service) dmSubtitle(c context.Context, aid, cid int64) (*dmmdl.VideoSubtitles, template.HTML) {
	if reply, err := s.dmGRPC.SubtitleGet(c, &dmmdl.SubtitleGetReq{Aid: aid, Oid: cid, Type: 1}); err != nil {
		log.Error("s.dm2.SubtitleGet aid(%d) cid(%d) error(%v)", aid, cid, err)
	} else {
		if reply != nil && reply.Subtitle != nil {
			if len(reply.Subtitle.Subtitles) == 0 {
				reply.Subtitle.Subtitles = make([]*dmmdl.VideoSubtitle, 0)
			}
			for _, v := range reply.Subtitle.Subtitles {
				v.SubtitleUrl = strings.Replace(v.SubtitleUrl, "http://", "//", 1)
			}
			if bs, err := json.Marshal(reply.Subtitle); err != nil {
				log.Error("dmSubject json.Marshal(%+v) error(%v)", reply.Subtitle, err)
			} else {
				// nolint:gosec
				return reply.Subtitle, template.HTML(bs)
			}
		}
	}
	return nil, ""
}

func (s *Service) tagPlayerIcon(c context.Context, aid, seasonID int64, typeID int32, mid int64) (*resmdl.PlayerIcon, template.HTML) {
	var tagIDs []int64
	tags, err := s.tag.ArcTags(c, &model.ArgAid{Aid: aid})
	if err != nil {
		log.Error("tagPlayerIcon s.tag.ArcTags aid(%d) error(%v)", aid, err)
	} else {
		for _, v := range tags {
			tagIDs = append(tagIDs, v.ID)
		}
	}
	req := &v1.WebPlayerIconRequest{Aid: aid, TagIDs: tagIDs, TypeID: typeID, SeasonID: seasonID, Mid: mid}
	webIconReply, err := s.resGRPC.WebPlayerIcon(c, req)
	if err != nil {
		if !ecode.EqualError(ecode.NothingFound, err) {
			log.Error("tagPlayerIcon s.resGRPC.WebPlayerIcon req(%+v) error(%v)", req, err)
		}
		return nil, ""
	}
	webIcon := webIconReply.GetIcon()
	if webIcon == nil {
		return nil, ""
	}
	playerIcon := &resmdl.PlayerIcon{
		URL1:  webIcon.URL1,
		Hash1: webIcon.Hash1,
		URL2:  webIcon.URL2,
		Hash2: webIcon.Hash2,
		CTime: webIcon.Ctime,
	}
	playerIcon.URL1 = strings.Replace(playerIcon.URL1, "http://", "https://", 1)
	playerIcon.URL2 = strings.Replace(playerIcon.URL2, "http://", "https://", 1)
	bs, err := json.Marshal(playerIcon)
	if err != nil {
		log.Error("tagPlayerIcon json.Marshal(%v) error(%v)", playerIcon, err)
		return nil, ""
	}
	// nolint:gosec
	return playerIcon, template.HTML(bs)
}

// nolint:gosec
func (s *Service) viewPoints(c context.Context, aid, cid int64) ([]*model.Point, template.HTML) {
	reply, err := s.videoUpGRPC.VideoViewPoints(c, &videoupmdl.VideoPointsReq{Aid: aid, Cid: cid})
	if err != nil {
		log.Error("viewPoints videoUpGRPC.VideoViewPoints aid:%d cid:%d error:%v", aid, cid, err)
		return []*model.Point{}, ""
	}
	var res []*model.Point
	for _, v := range reply.GetPoints() {
		if v == nil {
			continue
		}
		res = append(res, &model.Point{Type: v.Type, From: v.From, To: v.To, Content: v.Content, ImgUrl: v.ImgUrl, LogoUrl: v.LogoUrl})
	}
	for _, v := range reply.GetChapters() {
		if v == nil {
			continue
		}
		res = append(res, &model.Point{Type: v.Type, From: v.From, To: v.To, Content: v.Content, ImgUrl: v.ImgUrl, LogoUrl: v.LogoUrl})
	}
	if len(res) == 0 {
		return []*model.Point{}, ""
	}
	bs, err := json.Marshal(res)
	if err != nil {
		log.Error("viewPoints json.Marshal:%v, error:%v", res, err)
		return []*model.Point{}, ""
	}
	return res, template.HTML(bs)
}

func (s *Service) view(c context.Context, aid int64) (*arcmdl.ViewReply, error) {
	if view, ok := s.bnjViewMap[aid]; ok && view != nil {
		return view, nil
	}
	reply, err := s.arcGRPC.View(c, &arcmdl.ViewRequest{Aid: aid})
	if err != nil {
		return nil, s.slbRetryCode(err)
	}
	return reply, nil
}

func (s *Service) interaction(c context.Context, aid, cid, mid, graphVersion int64, buvid string) (*model.Interaction, template.HTML, error) {
	steinView, err := s.steinsGateGRPC.View(c, &api.ViewReq{Aid: aid, Mid: mid, Buvid: buvid})
	if err != nil {
		if ecode.EqualError(mecode.NonValidGraph, err) {
			err = ecode.NothingFound
		}
		return nil, "", err
	}
	if steinView.Graph == nil {
		log.Warn("interaction steinView.Graph is nil")
		return nil, "", ecode.NothingFound
	}
	stein := &model.Interaction{
		GraphVersion: steinView.Graph.Id,
		ErrorToast:   mecode.GraphInvalid.Message(),
		Mark:         steinView.Mark,
	}
	if graphVersion != 0 && steinView.Graph.Id != graphVersion {
		stein.Msg = mecode.GraphInvalid.Message()
		stein.NeedReload = 1
	} else if steinView.CurrentNode != nil {
		stein.HistoryNode = &model.Node{
			CID:    steinView.CurrentNode.Cid,
			Title:  steinView.CurrentNode.Name,
			NodeID: steinView.CurrentNode.Id,
		}
	}
	// 根节点才展示提示语
	if steinView.Graph.FirstCid == cid && graphVersion == 0 {
		stein.Msg = steinView.ToastMsg
	}
	bs, err := json.Marshal(stein)
	if err != nil {
		return nil, "", err
	}
	// nolint:gosec
	return stein, template.HTML(bs), nil
}

func (s *Service) pcdnLoader(c context.Context, cid, mid int64, refer, innerSign string) (json.RawMessage, template.HTML) {
	pcdn, err := s.dao.PcdnLoader(c, cid, mid, refer, innerSign)
	if err != nil {
		log.Error("pcdnData s.dao.PcdnLoader cid(%d) refer(%s) error(%v)", cid, refer, err)
		return []byte("{}"), ""
	}
	if len(pcdn) > 0 {
		// nolint:gosec
		return pcdn, template.HTML(pcdn)
	}
	return []byte("{}"), ""
}

func (s *Service) loadGuideCid() {
	if s.loadGuideCidRunning {
		return
	}
	defer func() {
		s.loadGuideCidRunning = false
	}()
	s.loadGuideCidRunning = true
	if s.c.Rule.SteinsGuideAid == 0 {
		return
	}
	view, err := s.arcGRPC.View(context.Background(), &arcmdl.ViewRequest{Aid: s.c.Rule.SteinsGuideAid})
	if err != nil {
		log.Error("loadGuideCid s.arcGRPC.View(%d) error(%v)", s.c.Rule.SteinsGuideAid, err)
		return
	}
	if view == nil || view.Arc == nil || !view.Arc.IsNormal() {
		log.Error("loadGuideCid view(%+v) data not allowed", view)
		return
	}
	tmp := make(map[int64]struct{}, len(view.Pages))
	for _, v := range view.Pages {
		if v == nil {
			continue
		}
		tmp[v.Cid] = struct{}{}
	}
	log.Warn("loadGuideCid success cids(%+v)", tmp)
	s.steinGuideCids = tmp
}

func (s *Service) loadLimitFreeList() {
	if s.loadLimitFreeRunning {
		return
	}
	defer func() {
		s.loadLimitFreeRunning = false
	}()
	s.loadLimitFreeRunning = true
	res, err := s.dao.LimitFree(context.Background())
	if err != nil {
		log.Error("loadLimitFreeList s.dao.LimitFree() error(%v)", err)
		return
	}
	s.limitFreeMap = res
}

// nolint:gocognit
func (s *Service) playerCards(ctx context.Context, aid, cid, epID, mid int64) (attens []*model.GuideAttention, cards []*model.JumpCard, operCard []*model.OperationCard, atten, skip, oper template.HTML) {
	cardsReply, err := func() (*appcfgmdl.PlayerCardsReply, error) {
		if epID > 0 {
			reply, err := s.appConfigGRPC.PlayerCardsOGV(ctx, &appcfgmdl.PlayerCardsOGVReq{Epid: epID, Mid: mid})
			if err != nil {
				return nil, err
			}
			return &appcfgmdl.PlayerCardsReply{Operations: reply.GetOperations()}, nil
		}
		reply, err := s.appConfigGRPC.PlayerCards(ctx, &appcfgmdl.PlayerCardsReq{Aid: aid, Cid: cid, Mid: mid})
		if err != nil {
			return nil, err
		}
		return reply, nil
	}()
	if err != nil {
		log.Error("playerCards s.appConfigGRPC.PlayerCards aid:%d cid:%d epID:%d mid:%d error(%+v)", aid, cid, epID, mid, err)
		return
	}
	for _, v := range cardsReply.GetAttentions() {
		if v != nil {
			attens = append(attens, &model.GuideAttention{
				Type: _guideTypeAtten,
				From: v.From,
				To:   v.To,
				PosX: v.PosX,
				PosY: v.PosY,
			})
		}
	}
	for _, v := range cardsReply.GetSkips() {
		if v != nil && v.Web != "" {
			cards = append(cards, &model.JumpCard{
				From:    v.From,
				To:      v.To,
				Icon:    v.Icon,
				Label:   v.Label,
				Content: v.Content,
				Button:  v.Button,
				Link:    v.Web,
			})
		}
	}
	for _, v := range cardsReply.GetOperations() {
		if v == nil {
			continue
		}
		tmp := &model.OperationCard{
			ID:       v.Id,
			From:     v.From,
			To:       v.To,
			Status:   v.Status,
			CardType: int32(v.CardType),
			BizType:  int32(v.BizType),
		}
		switch v.CardType {
		case appcfgmdl.OperationCardTypeStandard:
			if v.GetStandard() == nil {
				log.Warn("playerCards standard card nil aid:%d cid:%d mid:%d", aid, cid, mid)
				continue
			}
			tmp.StandardCard = &model.OperationStandardCard{
				Title:               v.GetStandard().Title,
				ButtonTitle:         v.GetStandard().ButtonTitle,
				ButtonSelectedTitle: v.GetStandard().ButtonSelectedTitle,
				ShowSelected:        v.GetStandard().ShowSelected,
			}
			switch v.BizType {
			case appcfgmdl.BizTypeFollowVideo:
				if v.GetFollow() != nil {
					tmp.ParamFollow = &model.ParamFollow{SeasonID: v.GetFollow().SeasonID}
				}
			case appcfgmdl.BizTypeReserveActivity:
				if v.GetReserve() != nil {
					tmp.ParamReserve = &model.ParamReserve{ActivityID: v.GetReserve().ActivityID}
				}
			default:
			}
		case appcfgmdl.OperationCardTypeSkip:
			if v.GetSkip() == nil {
				log.Warn("playerCards skip card nil aid:%d cid:%d mid:%d", aid, cid, mid)
				continue
			}
			if v.GetSkip().Web == "" {
				continue
			}
			tmp.SkipCard = &model.OperationSkipCard{
				From:    v.GetSkip().From,
				To:      v.GetSkip().To,
				Icon:    v.GetSkip().Icon,
				Label:   v.GetSkip().Label,
				Content: v.GetSkip().Content,
				Button:  v.GetSkip().Button,
				Link:    v.GetSkip().Web,
			}
			if v.BizType == appcfgmdl.BizTypeJumpLink && v.GetJump() != nil {
				tmp.ParamJump = &model.ParamJump{URL: v.GetJump().Url}
			}
		default:
			continue
		}
		operCard = append(operCard, tmp)
	}
	if len(attens) > 0 {
		bs, err := json.Marshal(attens)
		if err != nil {
			log.Error("playerCards guideAttentions json.Marshal error(%v)", err)
			return
		}
		// nolint:gosec
		atten = template.HTML(bs)
	}
	if len(cards) > 0 {
		bs, err := json.Marshal(cards)
		if err != nil {
			log.Error("playerCards skipCard json.Marshal error(%v)", err)
			return
		}
		// nolint:gosec
		skip = template.HTML(bs)
	}
	if len(operCard) > 0 {
		bs, err := json.Marshal(operCard)
		if err != nil {
			log.Error("playerCards operCard json.Marshal error(%v)", err)
			return
		}
		// nolint:gosec
		oper = template.HTML(bs)
	}
	if len(attens) == 0 {
		attens = []*model.GuideAttention{}
	}
	if len(cards) == 0 {
		cards = []*model.JumpCard{}
	}
	if len(operCard) == 0 {
		operCard = []*model.OperationCard{}
	}
	return
}

func (s *Service) ShowInfoc(ctx context.Context, ip string, now time.Time, buvid string, aid, mid int64) {
	const (
		_api    = "player"
		_client = "web"
	)
	info := struct {
		ip        string
		ctime     string
		api       string
		buvid     string
		mid       string
		client    string
		itemID    string
		displayID string
		errCode   string
		from      string
		build     string
		trackID   string
	}{
		ip:        ip,
		ctime:     strconv.FormatInt(now.Unix(), 10),
		api:       _api,
		buvid:     buvid,
		mid:       strconv.FormatInt(mid, 10),
		client:    _client,
		itemID:    strconv.FormatInt(aid, 10),
		displayID: "",
		errCode:   "",
		from:      "",
		build:     "",
		trackID:   "",
	}
	payload := infoc.NewLogStream(s.c.InfocLog.ShowLogID, info.ip, info.ctime, info.api, info.buvid, info.mid, info.client, info.itemID, info.displayID, info.errCode, info.from, info.build, info.trackID)
	if err := s.playInfoc.Info(ctx, payload); err != nil {
		log.Error("playInfoc.Info(%s,%s,%s,%s,%s) error:%v", info.ip, info.ctime, info.buvid, info.mid, info.itemID, err)
		return
	}
	log.Info("playInfoc.Info(%s,%s,%s,%s,%s) success", info.ip, info.ctime, info.buvid, info.mid, info.itemID)
}

func (s *Service) OnlineTotal(ctx context.Context, mid int64, buvid string, aid, cid int64, business int32) (*model.OnlineTotal, error) {
	arcReply, err := s.arcGRPC.SimpleArc(ctx, &arcmdl.SimpleArcRequest{Aid: aid})
	if err != nil {
		return nil, err
	}
	if arcReply.GetArc() == nil || !arcReply.GetArc().IsNormal() {
		return nil, ecode.NothingFound
	}
	var ok bool
	for _, val := range arcReply.GetArc().GetCids() {
		if cid == val {
			ok = true
			break
		}
	}
	if !ok {
		return nil, ecode.NothingFound
	}
	reply, err := s.playsvcGRPC.PlayOnline(ctx, &playsvcgrpc.PlayOnlineReq{Aid: aid, Cid: cid, Business: playsvcgrpc.OnlineBusiness(business)})
	if err != nil {
		return nil, err
	}
	group := s.abTest(mid, buvid, s.c.OnlineGray.RealWhitelist, s.c.OnlineGray.RealBucket)
	res := &model.OnlineTotal{
		Abtest: &model.Abtest{Group: group},
	}
	if reply.GetIsHide() {
		return res, nil
	}
	count := reply.GetCount()["web"]
	if count < 1 { // 兜底，避免初始化0在线
		count = 1
	}
	total := reply.GetCount()["total"]
	if total < 1 {
		total = 1
	}
	var onlineTotal string
	switch group {
	case model.GroupA:
		onlineTotal = totalString(total)
	case model.GroupB:
		onlineTotal = totalString2(total)
	}
	res.Count = countString(count)
	res.ShowSwitch.Count = true
	if !s.onlineGray(mid, buvid) {
		return res, nil
	}
	res.Total = onlineTotal
	res.ShowSwitch.Total = true
	return res, nil
}

func (s *Service) abTest(mid int64, buvid string, whitelist []int64, bucket uint64) model.Group {
	// a:对照组，b：实验组
	if mid != 0 {
		for _, val := range whitelist {
			if mid == val {
				return model.GroupB
			}
		}
	}
	h := md5.New()
	_, _ = h.Write([]byte(buvid))
	b, err := strconv.ParseUint(hex.EncodeToString(h.Sum(nil))[18:], 16, 64)
	if err != nil {
		log.Error("日志告警 分组错误,error:%+v", err)
		return model.GroupA
	}
	if b%100 < bucket {
		return model.GroupB
	}
	return model.GroupA
}

func (s *Service) onlineGray(mid int64, buvid string) bool {
	if s.c.OnlineGray.Open {
		return true
	}
	if mid < 1 {
		return false
	}
	for _, val := range s.c.OnlineGray.Whitelist {
		if mid == val {
			return true
		}
	}
	h := md5.New()
	_, _ = h.Write([]byte(buvid))
	b, err := strconv.ParseUint(hex.EncodeToString(h.Sum(nil))[18:], 16, 64)
	if err != nil {
		log.Error("日志告警 分组错误,error:%+v", err)
		return false
	}
	if b%100 < s.c.OnlineGray.Bucket {
		return true
	}
	return false
}

// nolint:gomnd
func countString(number int64) string {
	if number < 100000 {
		return strconv.FormatInt(number, 10)
	}
	return "10万+"
}

// nolint:gomnd
func totalString(number int64) string {
	if number < 10 {
		return "<10"
	}
	if number < 100 {
		return strconv.FormatInt(number/10*10, 10) + "+"
	}
	if number < 1000 {
		return strconv.FormatInt(number/100*100, 10) + "+"
	}
	if number < 10000 {
		return strconv.FormatInt(number/1000*1000, 10) + "+"
	}
	if number < 100000 {
		return strings.TrimSuffix(strconv.FormatFloat(float64(number)/10000, 'f', 1, 64), ".0") + "万+"
	}
	if number < 1000000 {
		return strings.TrimSuffix(strconv.FormatFloat(float64(number)/10000, 'f', 0, 64), ".0") + "万+"
	}
	return "100万+"
}

// nolint:gomnd
func totalString2(number int64) string {
	if number < 1000 {
		return strconv.FormatInt(number, 10)
	}
	if number < 10000 {
		return strconv.FormatInt(number/1000*1000, 10) + "+"
	}
	if number < 100000 {
		return strings.TrimSuffix(strconv.FormatFloat(float64(number)/10000, 'f', 1, 64), ".0") + "万+"
	}
	if number < 1000000 {
		return strings.TrimSuffix(strconv.FormatFloat(float64(number)/10000, 'f', 0, 64), ".0") + "万+"
	}
	return "100万+"
}
