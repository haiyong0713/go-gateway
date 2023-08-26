package api

import (
	"fmt"
	"hash/crc32"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/stat/prom"
	"go-common/library/time"

	"go-gateway/app/app-svr/archive/service/model"

	batch "git.bilibili.co/bapis/bapis-go/video/vod/playurlugcbatch"

	"github.com/thoas/go-funk"
)

// 各属性地址见 http://syncsvn.bilibili.co/platform/doc/blob/master/archive/field/state.md
// all const
const (
	// open state
	StateOpen    = 0
	StateOrange  = 1
	AccessMember = int32(10000)
	// forbid state
	StateForbidWait       = -1
	StateForbidRecicle    = -2
	StateForbidPolice     = -3
	StateForbidLock       = -4
	StateForbidFixed      = -6
	StateForbidLater      = -7
	StateForbidAdminDelay = -10
	StateForbidXcodeFail  = -16
	StateForbidSubmit     = -30
	StateForbidUserDelay  = -40
	StateForbidUpDelete   = -100
	StateForbidSteins     = -20
	// copyright
	CopyrightUnknow   = int8(0)
	CopyrightOriginal = int8(1)
	CopyrightCopy     = int8(2)

	// attribute yes and no
	AttrYes = int32(1)
	AttrNo  = int32(0)
	// attribute bit
	AttrBitNoRank    = uint(0)
	AttrBitNoDynamic = uint(1)
	AttrBitNoWeb     = uint(2)
	AttrBitNoMobile  = uint(3)
	// AttrBitNoSearch    = uint(4)
	AttrBitOverseaLock = uint(5)
	AttrBitNoRecommend = uint(6)
	AttrBitNoReprint   = uint(7)
	AttrBitHasHD5      = uint(8)
	AttrBitIsPGC       = uint(9)
	AttrBitAllowBp     = uint(10)
	AttrBitIsBangumi   = uint(11)
	AttrBitIsPorder    = uint(12)
	AttrBitLimitArea   = uint(13)
	AttrBitAllowTag    = uint(14)
	// AttrBitIsFromArcApi  = uint(15)
	AttrBitJumpUrl       = uint(16)
	AttrBitIsMovie       = uint(17)
	AttrBitBadgepay      = uint(18)
	AttrBitUGCPay        = uint(22)
	AttrBitHasBGM        = uint(23)
	AttrBitIsCooperation = uint(24)
	AttrBitHasViewpoint  = uint(25)
	AttrBitHasArgument   = uint(26)
	AttrBitUGCPayPreview = uint(27)
	AttrBitTeenager      = uint(28)
	AttrBitSteinsGate    = uint(29)
	AttrBitIsPUGVPay     = uint(30)
	// attribute_v2
	AttrBitV2NoBackground = uint(0)
	AttrBitV2NoPublic     = uint(1)
	// 是否360全景视频
	AttrBitV2Is360 = uint(2)
	//是否云非编稿件
	AttrBitV2BsEditor = uint(3)
	//是否存量导入的小视频
	AttrBitV2IsImportSvideo = uint(4)
	//播放页干净模式
	AttrBitV2CleanMode = uint(5)
	//禁止特别关注push
	AttrBitV2NoFansPush = uint(6)
	//是否开启杜比音效
	AttrBitV2IsDolby = uint(7)
	// 仅收藏可见
	AttrBitV2OnlyFavView = uint(8)
	// 是否活动合集
	AttrBitV2ActSeason = uint(9)
	// 是否是首映稿件
	AttrBitV2Premiere = uint(12)
	// 是否是付费稿件
	AttrBitV2Pay = uint(13)
	//是否是稿件自见
	AttrBitV2OnlySely = uint(17)
	AttrBitV2Charging = uint(18)
	// staff attribute
	StaffAttrBitAdOrder = uint(0)
	// Max playurl count
	_MaxPlayCnt = 50
	// archive internal attr
	InterAttrBitOverseaLock = uint(0)

	// 付费稿件，是否合集付费
	PaySubTypeAttrBitSeason = uint(0)
)

var (
	_emptyTags = []string{}
	promInfo   = prom.BusinessInfoCount
)

func (a *Arc) IsNormalPremiere() bool {
	return a.AttrValV2(AttrBitV2Premiere) == AttrYes && (a.State >= 0 || a.State == StateForbidUserDelay)
}

func (a *SimpleArc) IsNormalPremiere() bool {
	return a.AttrValV2(AttrBitV2Premiere) == AttrYes && (a.State >= 0 || a.State == StateForbidUserDelay)
}

// IsNormal is
func (a *Arc) IsNormal() bool {
	return a.State >= StateOpen
}

// (正常稿件state>=0  首映稿件会有state=-40的情况)
func (a *Arc) IsNormalV2() bool {
	return a.State >= StateOpen || (a.AttrValV2(AttrBitV2Premiere) == AttrYes && a.State == StateForbidUserDelay)
}

func (pay *PayInfo) AttrVal(bit uint) int32 {
	return int32((pay.PayAttr >> bit) & int64(1))
}

func (a *SimpleArc) IsNormal() bool {
	return a.State >= StateOpen
}

// RegionArc RegionArc
type RegionArc struct {
	Aid       int64
	Attribute int32
	Copyright int8
	PubDate   time.Time
}

// AllowShow AllowShow
func (ra *RegionArc) AllowShow() bool {
	return ra.attrVal(AttrBitNoWeb) == AttrNo && ra.attrVal(AttrBitNoMobile) == AttrNo
}

func (ra *RegionArc) attrVal(bit uint) int32 {
	return (ra.Attribute >> bit) & int32(1)
}

// AttrVal get attr val by bit.
func (a *Arc) AttrVal(bit uint) int32 {
	return (a.Attribute >> bit) & int32(1)
}

// AttrVal get attr val by bit.
func (sa *SimpleArc) AttrVal(bit uint) int32 {
	return (sa.Attribute >> bit) & int32(1)
}

// AttrVal get attr val by bit.
func (sa *SimpleArc) AttrValV2(bit uint) int32 {
	return int32((sa.AttributeV2 >> bit) & int64(1))
}

// AttrValV2 get attr v2 val by bit.
func (a *Arc) AttrValV2(bit uint) int32 {
	return int32((a.AttributeV2 >> bit) & int64(1))
}

// StaffAttrVal get staff attr val by bit.
func (staff *StaffInfo) StaffAttrVal(bit uint) int32 {
	return int32((staff.Attribute >> bit) & int64(1))
}

// FillDimensionAndFF is build dimension and first_frame
func (a *Arc) FillDimensionAndFF(d string) {
	a.FirstFrame = FFURL(a.FirstFrame)
	if d == "" || d == "0,0,0" {
		return
	}
	ds := strings.Split(d, ",")
	dsLen := 3
	if len(ds) != dsLen {
		return
	}
	var (
		width, height, rotate int64
		err                   error
	)
	if width, err = strconv.ParseInt(ds[0], 10, 64); err != nil {
		return
	}
	if height, err = strconv.ParseInt(ds[1], 10, 64); err != nil {
		return
	}
	if rotate, err = strconv.ParseInt(ds[2], 10, 64); err != nil {
		return
	}
	a.Dimension = Dimension{
		Width:  width,
		Height: height,
		Rotate: rotate,
	}
}

// FillDimensionAndFF is build dimension and first_frame
func (v *Page) FillDimensionAndFF(d string) {
	v.FirstFrame = FFURL(v.FirstFrame)
	if d == "" || d == "0,0,0" {
		return
	}
	ds := strings.Split(d, ",")
	dsLen := 3
	if len(ds) != dsLen {
		return
	}
	var (
		width, height, rotate int64
		err                   error
	)
	if width, err = strconv.ParseInt(ds[0], 10, 64); err != nil {
		return
	}
	if height, err = strconv.ParseInt(ds[1], 10, 64); err != nil {
		return
	}
	if rotate, err = strconv.ParseInt(ds[2], 10, 64); err != nil {
		return
	}
	v.Dimension = Dimension{
		Width:  width,
		Height: height,
		Rotate: rotate,
	}
}

// Fill file archive some field.
func (a *Arc) Fill() {
	a.Tags = _emptyTags
	a.Pic = coverURL(a.Pic)
	a.Rights.Bp = a.AttrVal(AttrBitAllowBp)
	a.Rights.Movie = a.AttrVal(AttrBitIsMovie)
	a.Rights.Pay = a.AttrVal(AttrBitBadgepay)
	a.Rights.HD5 = a.AttrVal(AttrBitHasHD5)
	a.Rights.NoReprint = a.AttrVal(AttrBitNoReprint)
	a.Rights.UGCPay = a.AttrVal(AttrBitUGCPay)
	a.Rights.IsCooperation = a.AttrVal(AttrBitIsCooperation)
	a.Rights.UGCPayPreview = a.AttrVal(AttrBitUGCPayPreview)
	a.Rights.NoBackground = a.AttrValV2(AttrBitV2NoBackground)
	a.Rights.Autoplay = CalcAutoplay(a) //autoplay并不是最终结果
	a.Rights.ArcPay = a.AttrValV2(AttrBitV2Pay)
}

type autoplayOpt struct {
	skipLimitArea bool
}

type autoplayOptFunc func(opt *autoplayOpt)

func SkipLimitArea() autoplayOptFunc {
	return func(opt *autoplayOpt) {
		opt.skipLimitArea = true
	}
}

func CalcAutoplayV2(a *Arc, inc *ArcInternal, opts ...autoplayOptFunc) int32 {
	ao := &autoplayOpt{}
	for _, opt := range opts {
		opt(ao)
	}
	if a.FirstCid == 0 {
		return 0
	}
	if a.Access == AccessMember {
		return 0
	}
	if a.AttrVal(AttrBitIsPGC) == AttrYes {
		return 0
	}
	if a.AttrVal(AttrBitAllowBp) == AttrYes {
		return 0
	}
	if a.AttrVal(AttrBitBadgepay) == AttrYes {
		return 0
	}
	if inc != nil && inc.attrVal(InterAttrBitOverseaLock) == int64(AttrYes) { //这里使用inc中的attr
		return 0
	}
	if a.AttrVal(AttrBitUGCPay) == AttrYes {
		return 0
	}
	if !ao.skipLimitArea {
		if a.AttrVal(AttrBitLimitArea) == AttrYes {
			return 0
		}
	}
	if a.AttrVal(AttrBitSteinsGate) == AttrYes {
		return 0
	}
	if a.AttrVal(AttrBitIsPUGVPay) == AttrYes {
		return 0
	}
	if a.AttrValV2(AttrBitV2OnlyFavView) == AttrYes {
		return 0
	}
	if a.AttrValV2(AttrBitV2OnlySely) == AttrYes {
		return 0
	}
	if a.AttrValV2(AttrBitV2Pay) == AttrYes {
		return 0
	}
	return 1
}

// CalcAutoplay 此func不可信，历史原因不可删除
func CalcAutoplay(a *Arc, opts ...autoplayOptFunc) int32 {
	ao := &autoplayOpt{}
	for _, opt := range opts {
		opt(ao)
	}
	if a.FirstCid == 0 {
		return 0
	}
	if a.Access == AccessMember {
		return 0
	}
	if a.AttrVal(AttrBitIsPGC) == AttrYes {
		return 0
	}
	if a.AttrVal(AttrBitAllowBp) == AttrYes {
		return 0
	}
	if a.AttrVal(AttrBitBadgepay) == AttrYes {
		return 0
	}
	if a.AttrVal(AttrBitOverseaLock) == AttrYes {
		return 0
	}
	if a.AttrVal(AttrBitUGCPay) == AttrYes {
		return 0
	}
	if !ao.skipLimitArea {
		if a.AttrVal(AttrBitLimitArea) == AttrYes {
			return 0
		}
	}
	if a.AttrVal(AttrBitSteinsGate) == AttrYes {
		return 0
	}
	if a.AttrVal(AttrBitIsPUGVPay) == AttrYes {
		return 0
	}
	if a.AttrValV2(AttrBitV2OnlyFavView) == AttrYes {
		return 0
	}
	return 1
}

// IsSteinsGate def.
func (a *Arc) IsSteinsGate() bool {
	return a.AttrVal(AttrBitSteinsGate) == AttrYes
}

func (a *Arc) IsBitV2Pay() bool {
	return a.AttrValV2(AttrBitV2Pay) == AttrYes
}

// coverURL convert cover url to full url.
func coverURL(uri string) (cover string) {
	if uri == "" {
		cover = "http://static.hdslb.com/images/transparent.gif"
		return
	}
	cover = uri
	if strings.Index(uri, "http://") == 0 {
		return
	}
	if len(uri) >= 10 && uri[:10] == "/templets/" {
		return
	}
	if strings.HasPrefix(uri, "group1") {
		cover = "http://i0.hdslb.com/" + uri
		return
	}
	if pos := strings.Index(uri, "/uploads/"); pos != -1 && (pos == 0 || pos == 3) {
		cover = uri[pos+8:]
	}
	cover = strings.Replace(cover, "{IMG}", "", -1)
	cover = "http://i" + strconv.FormatInt(int64(crc32.ChecksumIEEE([]byte(cover)))%3, 10) + ".hdslb.com" + cover
	return
}

func FFURL(uri string) string {
	if uri == "" {
		return ""
	}
	if strings.Index(uri, "http://") == 0 {
		return uri
	}
	return "http://i" + strconv.FormatInt(int64(crc32.ChecksumIEEE([]byte(uri)))%3, 10) + ".hdslb.com" + uri
}

// FillStat file stat, check access.
func (a *Arc) FillStat() {
	if a.Access > 0 {
		a.Stat.View = 0
	}
}

func (out *BvcVideoItem) FromBatch(in *batch.ResponseItem, askQn int64) {
	//粉6.8版本开始使用新字段AcceptFormats
	for _, v := range in.AcceptFormats {
		if v == nil {
			continue
		}
		var attr int64
		out.AcceptFormats = append(out.AcceptFormats, &FormatDescription{
			Quality:        v.Quality,
			Format:         v.Format,
			Description:    v.Description,
			Attribute:      setQnAttr(attr, v.Quality),
			NewDescription: v.NewDescription,
			DisplayDesc:    v.DisplayDesc,
			Superscript:    v.Superscript,
		})
	}
	out.ExpireTime = in.ExpireTime
	out.Cid = uint32(in.Cid)
	out.SupportQuality = append(out.SupportQuality, in.SupportQuality...)
	out.SupportFormats = in.SupportFormats
	out.SupportDescription = in.SupportDescription
	out.Url = in.Url
	out.FileInfo = make(map[uint32]*VideoFormatFileInfo, len(in.FileInfo))
	if len(in.FileInfo) > 0 {
		for cid, file := range in.FileInfo {
			tmpFile := new(VideoFormatFileInfo)
			tmpFile.fromFileInfo(file)
			out.FileInfo[cid] = tmpFile
		}
	}
	out.VideoCodecid = in.VideoCodecid
	out.VideoProject = in.VideoProject
	out.Fnver = in.Fnver
	out.Fnval = in.Fnval
	var (
		newDash  *ResponseDash
		qnChange bool
	)
	if in.Dash != nil {
		newDash, qnChange = fromDash(in.Dash, askQn)
		out.Dash = newDash
	}
	out.Quality = in.Quality
	//如果是story && 有720p
	if askQn > 0 && qnChange {
		out.Quality = uint32(askQn)
	}
	out.NoRexcode = in.NoRexcode
	out.BackupUrl = in.BackupUrl
	if in.Url != "" {
		promInfo.Incr(fmt.Sprintf("durl_num_%d", len(in.BackupUrl)))
	}
}

func (out *VideoFormatFileInfo) fromFileInfo(in *batch.VideoFormatFileInfo) {
	for _, v := range in.Infos {
		if v == nil {
			continue
		}
		out.Infos = append(out.Infos, &VideoFileInfo{
			Filesize:   v.Filesize,
			Timelength: v.Timelength,
			Ahead:      v.Ahead,
			Vhead:      v.Vhead,
		})
	}
}

func fromDash(in *batch.ResponseDash, askQn int64) (out *ResponseDash, qnChange bool) {
	out = new(ResponseDash)
	for _, v := range in.Video {
		if v == nil {
			continue
		}
		videoItem := &DashItem{
			Id:        v.Id,
			BaseUrl:   v.BaseUrl,
			Bandwidth: v.Bandwidth,
			Codecid:   v.Codecid,
			Size_:     v.Size_,
			BackupUrl: v.BackupUrl,
		}
		promInfo.Incr(fmt.Sprintf("dash_video_num_%d", len(v.BackupUrl)))
		if askQn > 0 && videoItem.Id == uint32(askQn) {
			qnChange = true
		}
		out.Video = append(out.Video, videoItem)
	}
	for _, v := range in.Audio {
		if v == nil {
			continue
		}
		audioItem := &DashItem{
			Id:        v.Id,
			BaseUrl:   v.BaseUrl,
			Bandwidth: v.Bandwidth,
			Codecid:   v.Codecid,
			Size_:     v.Size_,
			BackupUrl: v.BackupUrl,
		}
		promInfo.Incr(fmt.Sprintf("dash_audio_num_%d", len(v.BackupUrl)))
		out.Audio = append(out.Audio, audioItem)
	}
	return
}

// FromBatchV2 如果是dash 从所有清晰度中过滤出用户清晰度+秒开清晰度返回
// 用户清晰度：降路后的可观看清晰度
// 秒开清晰：视频云返回quality
func (out *BvcVideoItem) FromBatchV2(in *batch.ResponseItem, userQn, askQn int64, formatsNew, vipFree bool, cdnScore map[string]map[string]string, from string, netType NetworkType, isHighQn bool) {
	for _, v := range in.AcceptFormats {
		if v == nil {
			continue
		}
		var (
			attr               int64
			needLogin, needVip bool
		)
		//未登录最高清晰度 480
		if IsLoginQuality(v.Quality) {
			needLogin = true
		}
		if IsVipQuality(v.Quality) && !vipFree {
			needVip = true
		}
		out.AcceptFormats = append(out.AcceptFormats, &FormatDescription{
			Quality:        v.Quality,
			Format:         v.Format,
			Description:    v.Description,
			Attribute:      setQnAttr(attr, v.Quality),
			NewDescription: v.NewDescription,
			DisplayDesc:    v.DisplayDesc,
			Superscript:    v.Superscript,
			NeedLogin:      needLogin,
			NeedVip:        needVip,
		})
	}
	// 6.8版本开始不返回以下3个字段 改为使用AcceptFormats
	if !formatsNew {
		out.SupportQuality = in.SupportQuality
		out.SupportFormats = in.SupportFormats
		out.SupportDescription = in.SupportDescription
	}
	out.ExpireTime = in.ExpireTime
	out.Cid = uint32(in.Cid)
	out.Url = in.Url
	out.FileInfo = make(map[uint32]*VideoFormatFileInfo, len(in.FileInfo))
	if len(in.FileInfo) > 0 {
		for cid, file := range in.FileInfo {
			tmpFile := new(VideoFormatFileInfo)
			tmpFile.fromFileInfo(file)
			out.FileInfo[cid] = tmpFile
		}
	}
	out.VideoCodecid = in.VideoCodecid
	out.VideoProject = in.VideoProject
	out.Fnver = in.Fnver
	out.Fnval = in.Fnval
	out.NoRexcode = in.NoRexcode
	out.Quality = in.Quality
	if in.Dash != nil {
		out.Dash, out.Quality = fromDashV2(in.Dash, userQn, askQn, int64(in.Quality), cdnScore, from, netType, isHighQn)
	}
	out.BackupUrl = in.BackupUrl
	if in.Url != "" {
		promInfo.Incr(fmt.Sprintf("durl_num_%d", len(in.BackupUrl)))
	}
}

func fromDashV2(in *batch.ResponseDash, userQn, askQn, miaoQn int64, cdnScore map[string]map[string]string, from string, netType NetworkType, isHighQn bool) (*ResponseDash, uint32) {
	out := new(ResponseDash)
	var qn uint32
	for _, v := range in.Audio {
		if v == nil {
			continue
		}
		audioItem := &DashItem{
			Id:        v.Id,
			BaseUrl:   v.BaseUrl,
			Bandwidth: v.Bandwidth,
			Codecid:   v.Codecid,
			Size_:     v.Size_,
			BackupUrl: v.BackupUrl,
			FrameRate: v.FrameRate,
		}
		promInfo.Incr(fmt.Sprintf("dash_audio_num_%d", len(v.BackupUrl)))
		audioItem.BaseUrl = ChooseCdnUrl(audioItem.BaseUrl, cdnScore)
		for k, bku := range audioItem.BackupUrl {
			audioItem.BackupUrl[k] = ChooseCdnUrl(bku, cdnScore)
		}
		out.Audio = append(out.Audio, audioItem)
	}
	out.Video, qn = fromDashVideo(in.Video, userQn, askQn, miaoQn, cdnScore, from, netType, isHighQn)
	return out, qn
}

func fromDashVideo(video []*batch.DashItem, userQn, askQn, miaoQn int64, cdnScore map[string]map[string]string, from string, netType NetworkType, isHighQn bool) ([]*DashItem, uint32) {
	var (
		tmpVideo []*DashItem
		dashMap  = make(map[uint32][]*DashItem)
		minQn    = uint32(0)
	)
	for _, v := range video {
		if v == nil {
			continue
		}
		promInfo.Incr(fmt.Sprintf("dash_video_num_%d", len(v.BackupUrl)))
		tmpDash := &DashItem{
			Id:        v.Id,
			BaseUrl:   v.BaseUrl,
			Bandwidth: v.Bandwidth,
			Codecid:   v.Codecid,
			Size_:     v.Size_,
			NoRexcode: v.NoRexcode,
			FrameRate: v.FrameRate,
			BackupUrl: v.BackupUrl,
		}
		tmpDash.BaseUrl = ChooseCdnUrl(tmpDash.BaseUrl, cdnScore)
		for k, bku := range tmpDash.BackupUrl {
			tmpDash.BackupUrl[k] = ChooseCdnUrl(bku, cdnScore)
		}
		tmpVideo = append(tmpVideo, tmpDash)
		dashMap[v.Id] = append(dashMap[v.Id], tmpDash)
		if minQn == 0 || minQn > v.Id {
			minQn = v.Id
		}
	}
	if len(tmpVideo) == 0 {
		return nil, uint32(miaoQn)
	}
	res, quality := chooseQn(dashMap, tmpVideo, userQn, askQn, miaoQn, from, netType, isHighQn)
	// 如果最终过滤完没有结果，需要兜底返回最低清晰度
	if len(res) == 0 {
		res = dashMap[uint32(minQn)]
		quality = uint32(minQn)
	} else {
		// 保证最终吐出顺序
		sort.SliceStable(res, func(i, j int) bool {
			return res[i].Id > res[j].Id
		})
	}
	return res, quality
}

func chooseQn(dashMap map[uint32][]*DashItem, tmpVideo []*DashItem, userQn, askQn, miaoQn int64, from string, netType NetworkType, isHighQn bool) ([]*DashItem, uint32) {
	var res []*DashItem
	//story清晰度走单独的逻辑
	if from == "story" {
		return StoryDash(dashMap, netType, isHighQn)
	}
	//执行降路逻辑选择用户清晰度
	if askQn > 0 {
		miaoQn = askQn
	}
	selectUserQn := selectVideo(tmpVideo, userQn)
	if uVideo, ok := dashMap[uint32(selectUserQn)]; ok {
		res = append(res, uVideo...)
		// 非二压和二压清晰度之间无法无缝切换，所以如果用户设置qn是非二压的，那秒开qn也只返回这一路，保证半屏切全屏时不会黑屏
		// 出现过同一清晰度既有h264为非二压，又有h265为全二压的数据，认为该清晰度非二压且2路都返回，保证是否二压和编码解耦
		for _, v := range uVideo {
			if v.NoRexcode == 1 {
				return res, uint32(selectUserQn)
			}
		}
	}
	selectMiaoQn := selectVideo(tmpVideo, miaoQn)
	//选择秒开清晰度（如果要求story720p优先选择720，否则用视频云返回quality的清晰度）
	if mVideo, ok := dashMap[uint32(selectMiaoQn)]; ok && selectMiaoQn != selectUserQn {
		res = append(res, mVideo...)
	}
	return res, uint32(selectMiaoQn)
}

func StoryDash(dashMap map[uint32][]*DashItem, netType NetworkType, isHighQn bool) ([]*DashItem, uint32) {
	var res []*DashItem
	storyDefaultQn := uint32(0)
	//story下强制下发 1080p + 720p + 480p
	storyForceQn1080 := int64(model.Qn1080)
	storyForceQn720 := int64(model.QnFlv720)
	storyForceQn480 := int64(model.Qn480)
	storyForceQn1080Plus := int64(model.Qn1080Plus)
	if isHighQn { //是否要高清晰度1080P+
		if storyVideo, ok := dashMap[uint32(storyForceQn1080Plus)]; ok {
			res = append(res, storyVideo...)
		}
	}
	if storyVideo, ok := dashMap[uint32(storyForceQn1080)]; ok {
		res = append(res, storyVideo...)
		//story wifi默认qn为1080
		if netType == NetworkType_WIFI {
			storyDefaultQn = model.Qn1080
		}
	}
	if storyVideo, ok := dashMap[uint32(storyForceQn720)]; ok {
		res = append(res, storyVideo...)
		//story 4g 默认qn为720
		if netType == NetworkType_CELLULAR {
			storyDefaultQn = model.QnFlv720
		}
	}
	if storyVideo, ok := dashMap[uint32(storyForceQn480)]; ok {
		res = append(res, storyVideo...)
	}
	if storyDefaultQn == 0 && len(res) > 0 {
		storyDefaultQn = res[0].Id //第一个
		return res, storyDefaultQn
	}
	return res, storyDefaultQn
}

// 获取降路后可用清晰度
// 需要返回等级更低的且qn小于请求qn的（例如视频有qn:[116,74,80,64,21]，请求112，此时返回74）
func selectVideo(video []*DashItem, sourceQn int64) int64 {
	sort.SliceStable(video, func(i, j int) bool {
		return video[i].Id > video[j].Id
	})
	for _, uv := range video {
		if uv.Id <= uint32(sourceQn) {
			return int64(uv.Id)
		}
	}
	return 0
}

func setQnAttr(attr int64, qn uint32) int64 {
	var res int64
	if qn == model.QnHDR {
		res = attr | 1<<model.AttrIsHDR
	}
	if qn == model.QnDolbyHDR {
		res = attr | 1<<model.AttrIsDolbyHDR
	}
	return res
}

func IsVipQuality(qn uint32) bool {
	return qn >= model.Qn1080Plus
}

func IsLoginQuality(qn uint32) bool {
	return qn > model.Qn480
}

func (m *ArcsWithPlayurlRequest) ResetBatchArg() {
	if m.BatchPlayArg != nil {
		return
	}
	m.BatchPlayArg = &BatchPlayArg{
		Build:          m.Build,
		Device:         m.Device,
		NetType:        m.NetType,
		Qn:             m.Qn,
		MobiApp:        m.Platform,
		Fnver:          m.Fnver,
		Fnval:          m.Fnval,
		Ip:             m.Ip,
		Session:        m.Session,
		ForceHost:      m.ForceHost,
		Buvid:          m.Buvid,
		Mid:            m.Mid,
		Fourk:          m.Fourk,
		TfType:         m.TfType,
		From:           m.From,
		ShowPgcPlayurl: m.ShowPgcPlayurl,
	}
}

func (m *BatchPlayArg) AllowRequestBvc() bool {
	if m.MobiApp == "" || m.MobiApp == "unknown" || m.Ip == "" || m.Ip == "0.0.0.0" {
		return false
	}
	return true
}

func (m *ArcsPlayerRequest) Parse() (aids, noPlayAids []int64, aToC map[int64][]int64, aidHighQn map[int64]bool, err error) {
	if m.PlayAvs == nil {
		return nil, nil, nil, nil, ecode.RequestErr
	}
	aToC = make(map[int64][]int64)
	aidHighQn = make(map[int64]bool)
	var cids []int64
	for _, p := range m.PlayAvs {
		if p.Aid <= 0 {
			continue
		}
		if p.NoPlayer {
			noPlayAids = append(noPlayAids, p.Aid)
		}
		for _, v := range p.PlayVideos {
			if v.Cid <= 0 {
				continue
			}
			aToC[p.Aid] = append(aToC[p.Aid], v.Cid)
			cids = append(cids, v.Cid)
		}
		if p.HighQnExtra {
			aidHighQn[p.Aid] = p.HighQnExtra //高清晰度稿件
		}
		aids = append(aids, p.Aid)
	}
	aids = funk.UniqInt64(aids)
	if len(aids) > _MaxPlayCnt || len(aids) <= 0 || len(cids) > _MaxPlayCnt {
		return nil, nil, nil, aidHighQn, ecode.RequestErr
	}
	return aids, noPlayAids, aToC, aidHighQn, nil
}

func GetThirdDomain(dashUrl string) string {
	purl, err := url.Parse(dashUrl)
	if err != nil {
		log.Error("url.Parse(%s) error(%+v)", dashUrl, err)
		return ""
	}
	// 判断是否第三方cdn
	if !strings.Contains(purl.Host, "upos") {
		return ""
	}
	return purl.Host
}

type IpScore struct {
	Ip    string
	Score float64
}

// 第三方CDN选择优质IP
func ChooseCdnUrl(dashUrl string, cdnScore map[string]map[string]string) string {
	if len(cdnScore) == 0 {
		promInfo.Incr("无第三方cdn评分")
		return dashUrl
	}

	promInfo.Incr("有第三方cdn评分")
	baseDomain := GetThirdDomain(dashUrl)
	if baseDomain == "" {
		return dashUrl
	}

	promInfo.Incr("是第三方域名")
	ip, ok := cdnScore[baseDomain]
	if !ok || (len(ip["wwan"]) == 0 && len(ip["wifi"]) == 0) {
		return dashUrl
	}

	newDashUrl := dashUrl
	if strings.Contains(dashUrl, "?") {
		newDashUrl += "&"
	} else {
		newDashUrl += "?"
	}
	if len(ip["wwan"]) != 0 && len(ip["wifi"]) != 0 {
		promInfo.Incr("是第三方域名&指定ip&wwan&wifi")
		return fmt.Sprintf("%sclient_assign_ijk_ip_wwan=%s&client_assign_ijk_ip_wifi=%s", newDashUrl, ip["wwan"], ip["wifi"])
	}
	if len(ip["wwan"]) != 0 {
		promInfo.Incr("是第三方域名&指定ip&wwan")
		return fmt.Sprintf("%sclient_assign_ijk_ip_wwan=%s", newDashUrl, ip["wwan"])
	}
	if len(ip["wifi"]) != 0 {
		promInfo.Incr("是第三方域名&指定ip&wifi")
		return fmt.Sprintf("%sclient_assign_ijk_ip_wifi=%s", newDashUrl, ip["wifi"])
	}
	promInfo.Incr("unknown")
	return dashUrl
}

func (m *ArcInternal) UnfoldLimit() *ArcInnerLimit {
	rt := &ArcInnerLimit{}
	rt.OverseaBlock = m.attrVal(InterAttrBitOverseaLock) == int64(AttrYes)
	return rt
}

func (m *ArcInternal) attrVal(bit uint) int64 {
	return (m.Attribute >> bit) & int64(1)
}
