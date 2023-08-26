package model

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	"go-common/library/stat/prom"

	api "go-gateway/app/app-svr/app-player/interface/api/playurl"
	playurlApi "go-gateway/app/app-svr/playurl/service/api"
	vod "go-gateway/app/app-svr/playurl/service/api/v2"
)

const (
	_appKey   = "appkey"
	_ts       = "ts"
	_baseUri  = "https://app.bilibili.com/x/playurl"
	_maxBkNum = 2
)

var promInfo = prom.BusinessInfoCount

// Playurl is http://git.bilibili.co/video/playurl_doc/blob/master/PlayurlV2%E6%8E%A5%E5%8F%A3%E6%96%87%E6%A1%A3.md
type Playurl struct {
	From              string   `json:"from"`
	Result            string   `json:"result"`
	Quality           int64    `json:"quality"`
	Format            string   `json:"format"`
	Timelength        int64    `json:"timelength"`
	NoRexcode         int32    `json:"no_rexcode,omitempty"`
	AcceptFormat      string   `json:"accept_format"`
	AcceptDescription []string `json:"accept_description,omitempty"`
	AcceptQuality     []int64  `json:"accept_quality"`
	VideoCodecid      int      `json:"video_codecid"`
	Fnver             int      `json:"fnver"`
	Fnval             int      `json:"fnval"`
	VideoProject      bool     `json:"video_project"`
	SeekParam         string   `json:"seek_param"`
	SeekType          string   `json:"seek_type"`
	Abtid             int      `json:"abtid,omitempty"`
	Durl              []*Durl  `json:"durl,omitempty"`
	Dash              *Dash    `json:"dash,omitempty"`
}

type ParamHls struct {
	AID int64 `form:"aid" validate:"min=1"`
	CID int64 `form:"cid" validate:"min=1"`
	//qn清晰度
	Qn int64 `form:"qn" validate:"min=0"`
	// fnver和fnval标识视频格式
	Fnver  int32  `form:"fnver"`
	Fnval  int32  `form:"fnval"`
	Device string `form:"device"`
	// 返回url是否强制使用域名(非ip地址), 1-http域名 2-https域名
	ForceHost int32  `form:"force_host"`
	MobiApp   string `form:"mobi_app"`
	Build     int32  `form:"build"`
	Buvid     string `form:"-"` //从header获取
	Platform  string `form:"platform"`
	NetType   int32  `form:"-"`
	TfType    int32  `form:"-"`
	XTfIsp    string `form:"-"` //从header获取
	// 投屏设备 默认其他=0，OTT设备=1
	DeviceType int32 `form:"device_type"`
	//0:投屏 1:画中画
	RequestType int32 `form:"request_type"`
	//登陆态
	AccessKey string `form:"access_key"`
	//qn 0:mix 1:video 2:audio
	QnCategory int64 `form:"qn_category"`
	Dolby      int64 `form:"dolby"`
	//青少年模式
	TeenagersMode int64 `form:"teenagers_mode"`
	//课堂模式
	LessonsMode int64 `form:"lessons_mode"`
}

// sign calc appkey and appsecret sign.
func sign(params url.Values, key, secret string) (query string) {
	if params == nil {
		params = url.Values{}
	}
	params.Set(_appKey, key)
	params.Set(_ts, strconv.FormatInt(time.Now().Unix(), 10))
	tmp := params.Encode()
	if strings.IndexByte(tmp, '+') > -1 {
		tmp = strings.Replace(tmp, "+", "%20", -1)
	}
	var b bytes.Buffer
	b.WriteString(tmp)
	b.WriteString(secret)
	mh := md5.Sum(b.Bytes())
	// query
	var qb bytes.Buffer
	qb.WriteString(tmp)
	qb.WriteString("&sign=")
	qb.WriteString(hex.EncodeToString(mh[:]))
	query = qb.String()
	return
}

func formatNextRequest(in *ParamHls, qn int64, key, secret string, qnCategory vod.QnCategory) string {
	out := url.Values{}
	out.Set("aid", strconv.FormatInt(in.AID, 10))
	out.Set("cid", strconv.FormatInt(in.CID, 10))
	out.Set("qn", strconv.FormatInt(qn, 10))
	out.Set("fnver", strconv.FormatInt(int64(in.Fnver), 10))
	out.Set("fnval", strconv.FormatInt(int64(in.Fnval), 10))
	out.Set("device", in.Device)
	out.Set("mobi_app", in.MobiApp)
	out.Set("platform", in.Platform)
	out.Set("force_host", strconv.FormatInt(int64(in.ForceHost), 10))
	out.Set("build", strconv.FormatInt(int64(in.Build), 10))
	out.Set("device_type", strconv.FormatInt(int64(in.DeviceType), 10))
	out.Set("request_type", strconv.FormatInt(int64(in.RequestType), 10))
	out.Set("access_key", in.AccessKey)
	out.Set("actionKey", _appKey)
	out.Set("qn_category", strconv.FormatInt(int64(qnCategory), 10))
	out.Set("dolby", strconv.FormatInt(in.Dolby, 10))
	out.Set("teenagers_mode", strconv.FormatInt(in.TeenagersMode, 10))
	out.Set("lessons_mode", strconv.FormatInt(in.LessonsMode, 10))
	return sign(out, key, secret)
}

// Durl is
type Durl struct {
	Order     int      `json:"order"`
	Length    int64    `json:"length"`
	Size      int64    `json:"size"`
	AHead     string   `json:"ahead,omitempty"`
	VHead     string   `json:"vhead,omitempty"`
	URL       string   `json:"url"`
	BackupURL []string `json:"backup_url,omitempty"`
}

// Param is
type Param struct {
	AID           int64  `form:"aid" validate:"min=1"`
	CID           int64  `form:"cid" validate:"min=1"`
	Qn            int64  `form:"qn" validate:"min=0"`
	Npcybs        int32  `form:"npcybs"`
	Otype         string `form:"otype"`
	MobiApp       string `form:"mobi_app"`
	Fnver         int32  `form:"fnver"`
	Fnval         int32  `form:"fnval"`
	Session       string `form:"session"`
	Build         int32  `form:"build"`
	Device        string `form:"device"`
	ForceHost     int32  `form:"force_host"`
	Fourk         int32  `form:"fourk"`
	Buvid         string `form:"buvid"`
	Platform      string `form:"platform"`
	Dl            int32  `form:"dl"`
	Download      uint32 `form:"-"`
	FourkBool     bool   `form:"-"`
	Protocol      int32  `form:"-"`
	DeviceType    int32  `form:"-"`
	TeenagersMode int32  `form:"-"`
	PreferCodecID uint32 `form:"-"`
	VoiceBalance  int64  `form:"voice_balance"`
	NetType       int32
	TfType        int32
	LessonsMode   int32
	IP            string
	Business      api.Business
}

// Dash is
type Dash struct {
	Video []*DashItem `json:"video"`
	Audio []*DashItem `json:"audio"`
}

// DashItem is
type DashItem struct {
	ID           int64    `json:"id"`
	BaseURL      string   `json:"base_url"`
	BackupURL    []string `json:"backup_url,omitempty"`
	BaseURLRes   string   `json:"baseUrl,omitempty"`
	BackupURLRes []string `json:"backupUrl,omitempty"`
	Bandwidth    int64    `json:"bandwidth"`
	Codecid      int64    `json:"codecid"`
}

type PlayurlReply struct {
	*playurlApi.PlayURLReply
	UpgradeLimit *UpgradeLimit `json:"upgrade_limit,omitempty"`
}

type UpgradeLimit struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Image   string         `json:"image"`
	Button  *UpgradeButton `json:"button"`
}

type UpgradeButton struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

type HlsMasterReply struct {
	M3u8Data []byte `json:"m3u8_data"`
}

type PlayurlHlsReply struct {
	//表示返回的type是hls,flv还是mp4
	Type int32 `json:"type,omitempty"`
	//如果type是hls,则返回原生播放器下一步请求的地址
	URL string `json:"url,omitempty"`
	//返回视频的清晰度
	Quality uint32 `json:"quality,omitempty"`
	//返回视频的格式
	Format string `json:"format,omitempty"`
	//返回视频的总时长, 单位为ms
	Timelength uint64 `json:"timelength,omitempty"`
	//返回视频的编码号
	VideoCodecid uint32 `json:"video_codecid,omitempty"`
	//返回视频的是否支持投影
	VideoProject bool `json:"video_project,omitempty"`
	//返回视频播放url的列表，type为hls时，没有这个字段
	Durl []*api.ResponseUrl `json:"durl,omitempty"`
	//返回视频拥有的格式列表
	SupportFormats []*api.FormatDescription `json:"support_formats,omitempty"`
}

type PlayurlV2Reply struct {
	Quality           uint32        `json:"quality"`
	Format            string        `json:"format"`
	Timelength        uint64        `json:"timelength"`
	NoRexcode         int32         `json:"no_rexcode,omitempty"`
	AcceptFormat      string        `json:"accept_format"`
	AcceptDescription []string      `json:"accept_description,omitempty"`
	AcceptQuality     []uint32      `json:"accept_quality,omitempty"`
	VideoCodecid      uint32        `json:"video_codecid,omitempty"`
	Fnver             uint32        `json:"fnver"`
	Fnval             uint32        `json:"fnval"`
	VideoProject      bool          `json:"video_project"`
	Durl              []*DurlV2     `json:"durl,omitempty"`
	Dash              *DashV2       `json:"dash,omitempty"`
	UpgradeLimit      *UpgradeLimit `json:"upgrade_limit,omitempty"`
}

// DurlV2 is
type DurlV2 struct {
	Order     uint32   `json:"order"`
	Length    uint64   `json:"length"`
	Size      uint64   `json:"size"`
	AHead     string   `json:"ahead,omitempty"`
	VHead     string   `json:"vhead,omitempty"`
	URL       string   `json:"url"`
	BackupURL []string `json:"backup_url,omitempty"`
	Md5       string   `json:"md5,omitempty"`
}

// DashV2 is
type DashV2 struct {
	Video []*DashItemV2 `json:"video"`
	Audio []*DashItemV2 `json:"audio"`
}

// DashItemV2 is
type DashItemV2 struct {
	ID        uint32   `json:"id"`
	BaseURL   string   `json:"base_url"`
	BackupURL []string `json:"backup_url,omitempty"`
	Bandwidth uint32   `json:"bandwidth"`
	Codecid   uint32   `json:"codecid"`
	Md5       string   `json:"md5,omitempty"`
	Size      uint64   `json:"size,omitempty"`
}

func (pl *PlayurlV2Reply) FormatPlayURL(plVod *vod.ResponseMsg) {
	pl.Quality = plVod.Quality
	pl.Format = plVod.Format
	pl.Timelength = plVod.Timelength
	pl.NoRexcode = plVod.NoRexcode
	pl.AcceptFormat = plVod.AcceptFormat
	pl.AcceptDescription = plVod.AcceptDescription
	pl.AcceptQuality = plVod.AcceptQuality
	pl.VideoCodecid = plVod.VideoCodecid
	pl.Fnver = plVod.Fnver
	pl.Fnval = plVod.Fnval
	pl.VideoProject = plVod.VideoProject
	if len(plVod.Durl) > 0 {
		var tmpDurl []*DurlV2
		for _, v := range plVod.Durl {
			backupURL := v.BackupUrl
			promInfo.Incr(fmt.Sprintf("durl_num_%d", len(v.BackupUrl)))
			if len(v.BackupUrl) > _maxBkNum {
				backupURL = v.BackupUrl[:_maxBkNum]
			}
			tmpDurl = append(tmpDurl, &DurlV2{
				Order:     v.Order,
				Length:    v.Length,
				Size:      v.Size_,
				AHead:     v.Ahead,
				VHead:     v.Vhead,
				URL:       v.Url,
				BackupURL: backupURL,
				Md5:       v.Md5,
			})
		}
		pl.Durl = tmpDurl
	}
	if plVod.Dash != nil {
		pl.Dash = new(DashV2)
		if len(plVod.Dash.Video) > 0 {
			var tmpVideo []*DashItemV2
			for _, v := range plVod.Dash.Video {
				backupURL := v.BackupUrl
				promInfo.Incr(fmt.Sprintf("dash_num_%d", len(v.BackupUrl)))
				if len(v.BackupUrl) > _maxBkNum {
					backupURL = v.BackupUrl[:_maxBkNum]
				}
				tmpVideo = append(tmpVideo, &DashItemV2{
					ID:        v.Id,
					BaseURL:   v.BaseUrl,
					BackupURL: backupURL,
					Bandwidth: v.Bandwidth,
					Codecid:   v.Codecid,
					Md5:       v.Md5,
					Size:      v.Size_,
				})
			}
			pl.Dash.Video = tmpVideo
		}
		if len(plVod.Dash.Audio) > 0 {
			var tmpAudio []*DashItemV2
			for _, v := range plVod.Dash.Audio {
				backupURL := v.BackupUrl
				promInfo.Incr(fmt.Sprintf("dash_num_%d", len(v.BackupUrl)))
				if len(v.BackupUrl) > _maxBkNum {
					backupURL = v.BackupUrl[:_maxBkNum]
				}
				tmpAudio = append(tmpAudio, &DashItemV2{
					ID:        v.Id,
					BaseURL:   v.BaseUrl,
					BackupURL: backupURL,
					Bandwidth: v.Bandwidth,
					Codecid:   v.Codecid,
					Md5:       v.Md5,
					Size:      v.Size_,
				})
			}
			pl.Dash.Audio = tmpAudio
		}
	}
}

func FormatPlayURLGRPC(plVod *vod.ResponseMsg) (res *api.PlayURLReply) {
	res = new(api.PlayURLReply)
	res.Quality = plVod.Quality
	res.Format = plVod.Format
	res.Timelength = plVod.Timelength
	res.NoRexcode = plVod.NoRexcode
	res.VideoCodecid = plVod.VideoCodecid
	res.Fnver = plVod.Fnver
	res.Fnval = plVod.Fnval
	res.VideoProject = plVod.VideoProject
	res.Type = api.VideoType(plVod.Type)
	if len(plVod.Durl) > 0 {
		var tmpDurl []*api.ResponseUrl
		for _, v := range plVod.Durl {
			backupURL := v.BackupUrl
			promInfo.Incr(fmt.Sprintf("durl_num_%d", len(v.BackupUrl)))
			if len(v.BackupUrl) > _maxBkNum {
				backupURL = v.BackupUrl[:_maxBkNum]
			}
			tmpDurl = append(tmpDurl, &api.ResponseUrl{
				Order:     v.Order,
				Length:    v.Length,
				Size_:     v.Size_,
				Url:       v.Url,
				BackupUrl: backupURL,
				Md5:       v.Md5,
			})
		}
		res.Durl = tmpDurl
	}
	if plVod.Dash != nil {
		res.Dash = new(api.ResponseDash)
		if len(plVod.Dash.Video) > 0 {
			var tmpVideo []*api.DashItem
			for _, v := range plVod.Dash.Video {
				backupURL := v.BackupUrl
				promInfo.Incr(fmt.Sprintf("dash_num_%d", len(v.BackupUrl)))
				if len(v.BackupUrl) > _maxBkNum {
					backupURL = v.BackupUrl[:_maxBkNum]
				}
				tmpVideo = append(tmpVideo, &api.DashItem{
					Id:        v.Id,
					BaseUrl:   v.BaseUrl,
					BackupUrl: backupURL,
					Bandwidth: v.Bandwidth,
					Codecid:   v.Codecid,
					Md5:       v.Md5,
					Size_:     v.Size_,
					FrameRate: v.FrameRate,
				})
			}
			res.Dash.Video = tmpVideo
		}
		if len(plVod.Dash.Audio) > 0 {
			var tmpAudio []*api.DashItem
			for _, v := range plVod.Dash.Audio {
				backupURL := v.BackupUrl
				promInfo.Incr(fmt.Sprintf("dash_num_%d", len(v.BackupUrl)))
				if len(v.BackupUrl) > _maxBkNum {
					backupURL = v.BackupUrl[:_maxBkNum]
				}
				tmpAudio = append(tmpAudio, &api.DashItem{
					Id:        v.Id,
					BaseUrl:   v.BaseUrl,
					BackupUrl: backupURL,
					Bandwidth: v.Bandwidth,
					Codecid:   v.Codecid,
					Md5:       v.Md5,
					Size_:     v.Size_,
					FrameRate: v.FrameRate,
				})
			}
			res.Dash.Audio = tmpAudio
		}
	}
	if len(plVod.SupportFormats) > 0 {
		var tmpSupFormats []*api.FormatDescription
		for _, v := range plVod.SupportFormats {
			tmpSupFormats = append(tmpSupFormats, &api.FormatDescription{
				Quality:        v.Quality,
				Format:         v.Format,
				Description:    v.Description,
				NewDescription: v.NewDescription,
				DisplayDesc:    v.DisplayDesc,
				Superscript:    v.Superscript,
			})
		}
		res.SupportFormats = tmpSupFormats
	}
	return
}

// DlNumParam is
type DlNumParam struct {
	MobiApp string `form:"mobi_app"`
	Build   int32  `form:"build"`
	Device  string `form:"device"`
	Buvid   string `form:"buvid"`
	Num     int32  `form:"num"`
}

// FormatPlayDurl durl列表拼接 .
func FormatPlayDurl(plVod *vod.ResponseMsg, vipInfo *vod.ExtInfo, vipFree bool, qnSubtitle string) (rly []*api.Stream) {
	tmpDurl := &api.Stream_SegmentVideo{SegmentVideo: &api.SegmentVideo{}}
	for _, v := range plVod.Durl {
		backupURL := v.BackupUrl
		promInfo.Incr(fmt.Sprintf("durl_num_%d", len(v.BackupUrl)))
		if len(v.BackupUrl) > _maxBkNum {
			backupURL = v.BackupUrl[:_maxBkNum]
		}
		tmpDurl.SegmentVideo.Segment = append(tmpDurl.SegmentVideo.Segment, &api.ResponseUrl{
			Order:     v.Order,
			Length:    v.Length,
			Size_:     v.Size_,
			Url:       v.Url,
			BackupUrl: backupURL,
			Md5:       v.Md5,
		})
	}
	// 清晰度列表
	ti := 0
	for _, tVal := range plVod.SupportFormats {
		if tVal == nil {
			continue
		}
		var attr int64
		tmpStresm := &api.Stream{StreamInfo: &api.StreamInfo{
			Quality:        tVal.Quality,
			Format:         tVal.Format,
			Description:    tVal.Description,
			Attribute:      setQnAttr(attr, tVal.Quality),
			NewDescription: tVal.NewDescription,
			DisplayDesc:    tVal.DisplayDesc,
			Superscript:    tVal.Superscript,
		}}
		if IsLoginQuality(tVal.Quality) {
			tmpStresm.StreamInfo.NeedLogin = true
		}
		if IsVipQuality(tVal.Quality) {
			if !vipFree {
				tmpStresm.StreamInfo.NeedVip = true
			}
			tmpStresm.StreamInfo.VipFree = vipFree
		}
		if tVal.Quality == QnHDR || tVal.Quality == QnDolbyHDR {
			tmpStresm.StreamInfo.Subtitle = "绚丽色彩，沉浸体验"
		}
		//最高清晰度
		if IsSubtitleQuality(tVal.Quality) && ti == 0 && qnSubtitle != "" {
			tmpStresm.StreamInfo.Subtitle = qnSubtitle
			ti++
		}
		// 当前可播放清晰度
		if tVal.Quality == plVod.Quality {
			tmpStresm.StreamInfo.Intact = true
			tmpStresm.Content = tmpDurl
		} else {
			if IsVipQuality(tVal.Quality) && vipInfo != nil && vipInfo.VipControl != nil && vipInfo.VipControl.Control { // 大会员清晰度管控
				// vip 管控中
				tmpStresm.StreamInfo.ErrCode = api.PlayErr_WithMultiDeviceLoginErr
				tmpStresm.StreamInfo.Limit = &api.StreamLimit{Title: "账户存在风险，该清晰度禁用", Msg: "验证观看", Uri: "https://big.bilibili.com/mobile/windControl"}
			}
		}
		rly = append(rly, tmpStresm)
	}
	return
}

// nolint:gocognit
func FormatPlayDash(plVod *vod.ResponseMsg, vipInfo *vod.ExtInfo, arg *Param, vipFree bool, cdnScore map[string]map[string]string, qnSubtitle string) (rly []*api.Stream, dashAudio []*api.DashItem, dolby *api.DolbyItem, lossLess *api.LossLessItem) {
	//如果视频有无损
	if plVod.GetDash().GetLossLessItem() != nil {
		lossLess = &api.LossLessItem{IsLosslessAudio: plVod.GetDash().GetLossLessItem().GetIsLosslessAudio(), NeedVip: !vipFree}
		lv := plVod.GetDash().GetLossLessItem().GetAudio()
		if lv != nil {
			lossLess.Audio = &api.DashItem{
				Id:        lv.Id,
				BaseUrl:   lv.BaseUrl,
				BackupUrl: lv.BackupUrl,
				Bandwidth: lv.Bandwidth,
				Codecid:   lv.Codecid,
				Md5:       lv.Md5,
				Size_:     lv.Size_,
				FrameRate: lv.FrameRate,
			}
		}
	}
	// 如果视频有杜比
	if plVod.Dash.Dolby != nil {
		dolby = &api.DolbyItem{Type: api.DolbyItem_Type(plVod.Dash.Dolby.Type)}
		if len(plVod.Dash.Dolby.Audio) > 0 { //如用户非大会员会过滤音频流 仅保留type
			var tmpAd []*api.DashItem
			for _, v := range plVod.Dash.Dolby.Audio {
				tmpAd = append(tmpAd, &api.DashItem{
					Id:        v.Id,
					BaseUrl:   v.BaseUrl,
					BackupUrl: v.BackupUrl,
					Bandwidth: v.Bandwidth,
					Codecid:   v.Codecid,
					Md5:       v.Md5,
					Size_:     v.Size_,
					FrameRate: v.FrameRate,
				})
			}
			dolby.Audio = tmpAd
		}
	}
	// 音频信息默认给第一个
	var defaultAudioId uint32
	if len(plVod.Dash.Audio) > 0 {
		for _, aVal := range plVod.Dash.Audio {
			backupURL := aVal.BackupUrl
			promInfo.Incr(fmt.Sprintf("dash_num_%d", len(aVal.BackupUrl)))
			if len(aVal.BackupUrl) > _maxBkNum {
				backupURL = aVal.BackupUrl[:_maxBkNum]
			}
			tmpAudio := &api.DashItem{
				Id:        aVal.Id,
				BaseUrl:   aVal.BaseUrl,
				BackupUrl: backupURL,
				Bandwidth: aVal.Bandwidth,
				Codecid:   aVal.Codecid,
				Md5:       aVal.Md5,
				Size_:     aVal.Size_,
				FrameRate: aVal.FrameRate,
			}
			tmpAudio.BaseUrl = ChooseCdnUrl(tmpAudio.BaseUrl, cdnScore)
			for k, v := range tmpAudio.BackupUrl {
				tmpAudio.BackupUrl[k] = ChooseCdnUrl(v, cdnScore)
			}
			dashAudio = append(dashAudio, tmpAudio)
		}
		defaultAudioId = plVod.Dash.Audio[0].Id
	}
	// 每路清晰度对应的播放地址和音频信息
	var fnVideo map[uint32]*api.DashVideo
	if len(plVod.Dash.Video) > 0 {
		//每一路清晰度对应多种编码格式的播放地址
		var tmpVideo = make(map[uint32][]*api.DashVideo)
		for _, v := range plVod.Dash.Video {
			backupURL := v.BackupUrl
			promInfo.Incr(fmt.Sprintf("dash_num_%d", len(v.BackupUrl)))
			if len(v.BackupUrl) > _maxBkNum {
				backupURL = v.BackupUrl[:_maxBkNum]
			}
			tmpDash := &api.DashVideo{
				BaseUrl:   v.BaseUrl,
				BackupUrl: backupURL,
				Bandwidth: v.Bandwidth,
				Codecid:   v.Codecid,
				Md5:       v.Md5,
				Size_:     v.Size_,
				AudioId:   defaultAudioId, // 音频信息
				NoRexcode: v.NoRexcode == 1,
				FrameRate: v.FrameRate,
				Width:     v.Width,
				Height:    v.Height,
			}
			tmpDash.BaseUrl = ChooseCdnUrl(tmpDash.BaseUrl, cdnScore)
			for k, v := range tmpDash.BackupUrl {
				tmpDash.BackupUrl[k] = ChooseCdnUrl(v, cdnScore)
			}
			tmpVideo[v.Id] = append(tmpVideo[v.Id], tmpDash)
		}
		fnVideo = chooseFnVideo(arg.PreferCodecID, tmpVideo)
	}
	// 清晰度列表 拼接信息
	ti := 0
	for _, tVal := range plVod.SupportFormats {
		if tVal == nil {
			continue
		}
		var attr int64
		tmpStresm := &api.Stream{
			StreamInfo: &api.StreamInfo{
				Quality:        tVal.Quality,
				Format:         tVal.Format,
				Description:    tVal.Description,
				Attribute:      setQnAttr(attr, tVal.Quality),
				NewDescription: tVal.NewDescription,
				DisplayDesc:    tVal.DisplayDesc,
				Superscript:    tVal.Superscript,
			},
		}
		//未登录最高清晰度 480
		if IsLoginQuality(tVal.Quality) {
			tmpStresm.StreamInfo.NeedLogin = true
		}
		if IsVipQuality(tVal.Quality) {
			if !vipFree {
				tmpStresm.StreamInfo.NeedVip = true
			}
			tmpStresm.StreamInfo.VipFree = vipFree
		}
		if tVal.Quality == QnHDR || tVal.Quality == QnDolbyHDR {
			tmpStresm.StreamInfo.Subtitle = "绚丽色彩，沉浸体验"
		}
		if IsSubtitleQuality(tVal.Quality) && ti == 0 && qnSubtitle != "" {
			tmpStresm.StreamInfo.Subtitle = qnSubtitle
			ti++
		}
		// 视频云返回了对应的播放地址
		if _, ok := fnVideo[tVal.Quality]; ok {
			tmpStresm.StreamInfo.Intact = true
			tmpStresm.StreamInfo.NoRexcode = fnVideo[tVal.Quality].NoRexcode
			// 同一清晰度只返回一路
			tmpStresm.Content = &api.Stream_DashVideo{DashVideo: fnVideo[tVal.Quality]}
		} else { // 没有返回播放地址
			if IsVipQuality(tVal.Quality) && vipInfo != nil && vipInfo.VipControl != nil && vipInfo.VipControl.Control { // 大会员清晰度管控
				// vip 管控中
				tmpStresm.StreamInfo.ErrCode = api.PlayErr_WithMultiDeviceLoginErr
				tmpStresm.StreamInfo.Limit = &api.StreamLimit{Title: "账户存在风险，该清晰度禁用", Msg: "验证观看", Uri: "https://big.bilibili.com/mobile/windControl"}
			}
		}
		rly = append(rly, tmpStresm)
	}
	return
}

// 不同的编码格式对应的播放地址
type codecidDashVideos struct {
	// key:清晰度
	av1DashVideos  map[uint32]*api.DashVideo
	h265DashVideos map[uint32]*api.DashVideo
	h264DashVideos map[uint32]*api.DashVideo
}

// 分离编不同码格式的播放地址
func spliteDashVideos(in map[uint32][]*api.DashVideo) codecidDashVideos {
	out := codecidDashVideos{
		av1DashVideos:  make(map[uint32]*api.DashVideo, len(in)),
		h265DashVideos: make(map[uint32]*api.DashVideo, len(in)),
		h264DashVideos: make(map[uint32]*api.DashVideo, len(in)),
	}
	//qn 清晰度
	for qn, dashVideos := range in {
		for _, dashVideo := range dashVideos {
			switch dashVideo.Codecid {
			case CodeH264:
				out.h264DashVideos[qn] = dashVideo
			case CodeH265:
				out.h265DashVideos[qn] = dashVideo
			case CodeAV1:
				out.av1DashVideos[qn] = dashVideo
			default:
				log.Error("unexpected codecid(%d)", dashVideo.Codecid)
			}
		}
	}
	return out
}

// 选择合适编码方式的播放地址
// 优先匹配客户端需要的编码方式对应的播放地址
// 当客户端要av1的时候：有av1返回av1,否则就拿h265,以h264兜底
// 当客户端要h265的时候：有h265就拿h265,否则拿h264
func chooseFnVideo(codecid uint32, in map[uint32][]*api.DashVideo) map[uint32]*api.DashVideo {
	codecidDashVideos := spliteDashVideos(in)
	fnVideo := make(map[uint32]*api.DashVideo)
	for qn := range in {
		dashVideo := func() *api.DashVideo {
			switch codecid {
			case CodeAV1:
				if dashVideo, ok := codecidDashVideos.av1DashVideos[qn]; ok {
					return dashVideo
				}
				if dashVideo, ok := codecidDashVideos.h265DashVideos[qn]; ok {
					return dashVideo
				}
				return codecidDashVideos.h264DashVideos[qn]
			case CodeH265:
				if dashVideo, ok := codecidDashVideos.h265DashVideos[qn]; ok {
					return dashVideo
				}
				return codecidDashVideos.h264DashVideos[qn]
			default:
				return codecidDashVideos.h264DashVideos[qn]
			}
		}()
		if dashVideo == nil { //该清晰度下没有对应编码格式的视频播放地址
			continue
		}
		fnVideo[qn] = dashVideo
	}
	return fnVideo
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

func FormatPlayInfoGRPC(plVod *vod.ResponseMsg, vipControl *vod.ExtInfo, arg *Param, vipFree bool, cdnScore map[string]map[string]string, qnSubtitle string) (res *api.VideoInfo) {
	res = new(api.VideoInfo)
	res.Quality = plVod.Quality
	res.Format = plVod.Format
	res.Timelength = plVod.Timelength
	// 返回当前的
	res.VideoCodecid = plVod.VideoCodecid
	// 视频下所有清晰度列表
	if len(plVod.SupportFormats) == 0 {
		return
	}
	//durl 返回的当前请求清晰度,整个数组才是完整的播放地址
	if len(plVod.Durl) > 0 {
		res.StreamList = FormatPlayDurl(plVod, vipControl, vipFree, qnSubtitle)
	} else if plVod.Dash != nil { // dash 返回多路清晰度,每一路都是完整的播放地址
		res.StreamList, res.DashAudio, res.Dolby, res.LossLessItem = FormatPlayDash(plVod, vipControl, arg, vipFree, cdnScore, qnSubtitle)
	}
	return
}

// FormatPlayArcConf
func FormatPlayArcConf(conf *vod.PlayArcConf) (rly *api.PlayArcConf) {
	rly = &api.PlayArcConf{}
	rly.BackgroundPlayConf = FormatArc(conf.BackgroundPlayConf)
	rly.FlipConf = FormatArc(conf.FlipConf)
	rly.CastConf = FormatArc(conf.CastConf)
	rly.FeedbackConf = FormatArc(conf.FeedbackConf)
	rly.SubtitleConf = FormatArc(conf.SubtitleConf)
	rly.PlaybackRateConf = FormatArc(conf.PlaybackRateConf)
	rly.TimeUpConf = FormatArc(conf.TimeUpConf)
	rly.PlaybackModeConf = FormatArc(conf.PlaybackModeConf)
	rly.ScaleModeConf = FormatArc(conf.ScaleModeConf)
	rly.LikeConf = FormatArc(conf.LikeConf)
	rly.DislikeConf = FormatArc(conf.DislikeConf)
	rly.CoinConf = FormatArc(conf.CoinConf)
	rly.ElecConf = FormatArc(conf.ElecConf)
	rly.ShareConf = FormatArc(conf.ShareConf)
	rly.ScreenShotConf = FormatArc(conf.ScreenShotConf)
	rly.LockScreenConf = FormatArc(conf.LockScreenConf)
	rly.RecommendConf = FormatArc(conf.RecommendConf)
	rly.PlaybackSpeedConf = FormatArc(conf.PlaybackSpeedConf)
	rly.DefinitionConf = FormatArc(conf.DefinitionConf)
	rly.SelectionsConf = FormatArc(conf.SelectionsConf)
	rly.NextConf = FormatArc(conf.NextConf)
	rly.EditDmConf = FormatArc(conf.EditDmConf)
	rly.InnerDmConf = FormatArc(conf.InnerDmConf)
	rly.OuterDmConf = FormatArc(conf.OuterDmConf)
	rly.ShakeConf = FormatArc(conf.ShakeConf)
	rly.SmallWindowConf = FormatArc(conf.SmallWindowConf)
	rly.PanoramaConf = FormatArc(conf.PanoramaConf)
	rly.DolbyConf = FormatArc(conf.DolbyConf)
	rly.ScreenRecordingConf = FormatArc(conf.ScreenRecordingConf)
	rly.ColorFilterConf = FormatArc(conf.ColorFilterConf)
	rly.LossLessConf = FormatArc(conf.LossLessConf)
	return
}

func confTypeConvert(in api.ConfType) (vod.ConfType, bool) {
	var out vod.ConfType
	switch in {
	case api.ConfType_BACKGROUNDPLAY:
		out = vod.ConfType_BACKGROUNDPLAY
	case api.ConfType_SCALEMODE:
		out = vod.ConfType_SCALEMODE
	case api.ConfType_PLAYBACKMODE:
		out = vod.ConfType_PLAYBACKMODE
	case api.ConfType_TIMEUP:
		out = vod.ConfType_TIMEUP
	case api.ConfType_PLAYBACKRATE:
		out = vod.ConfType_PLAYBACKRATE
	case api.ConfType_SUBTITLE:
		out = vod.ConfType_SUBTITLE
	case api.ConfType_FEEDBACK:
		out = vod.ConfType_FEEDBACK
	case api.ConfType_CASTCONF:
		out = vod.ConfType_CASTCONF
	case api.ConfType_FLIPCONF:
		out = vod.ConfType_FLIPCONF
	case api.ConfType_DISLIKE:
		out = vod.ConfType_DISLIKE
	case api.ConfType_COIN:
		out = vod.ConfType_COIN
	case api.ConfType_ELEC:
		out = vod.ConfType_ELEC
	case api.ConfType_SCREENSHOT:
		out = vod.ConfType_SCREENSHOT
	case api.ConfType_LOCKSCREEN:
		out = vod.ConfType_LOCKSCREEN
	case api.ConfType_RECOMMEND:
		out = vod.ConfType_RECOMMEND
	case api.ConfType_PLAYBACKSPEED:
		out = vod.ConfType_PLAYBACKSPEED
	case api.ConfType_DEFINITION:
		out = vod.ConfType_DEFINITION
	case api.ConfType_SELECTIONS:
		out = vod.ConfType_SELECTIONS
	case api.ConfType_SHAKE:
		out = vod.ConfType_SHAKE
	case api.ConfType_SMALLWINDOW:
		out = vod.ConfType_SMALLWINDOW
	case api.ConfType_INNERDM:
		out = vod.ConfType_INNERDM
	case api.ConfType_PANORAMA:
		out = vod.ConfType_PANORAMA
	case api.ConfType_DOLBY:
		out = vod.ConfType_DOLBY
	case api.ConfType_COLORFILTER:
		out = vod.ConfType_COLORFILTER
	case api.ConfType_LOSSLESS:
		out = vod.ConfType_LOSSLESS
	//以下配置产品要求默认不可编辑
	//case api.ConfType_SHARE:
	//	out = vod.ConfType_SHARE
	//case api.ConfType_LIKE:
	//	out = vod.ConfType_LIKE
	//case api.ConfType_NEXT:
	//	out = vod.ConfType_NEXT
	//case api.ConfType_EDITDM:
	//	out = vod.ConfType_EDITDM
	//case api.ConfType_OUTERDM:
	//	out = vod.ConfType_OUTERDM
	default:
		log.Warn("Failed to match conf type")
		return out, false
	}
	return out, true
}

func confValueToVodConf(in *api.ConfValue) *vod.ConfValue {
	var out *vod.ConfValue
	switch v := in.Value.(type) {
	case *api.ConfValue_SwitchVal:
		out = &vod.ConfValue{Value: &vod.ConfValue_SwitchVal{SwitchVal: v.SwitchVal}}
	case *api.ConfValue_SelectedVal:
		out = &vod.ConfValue{Value: &vod.ConfValue_SelectedVal{SelectedVal: v.SelectedVal}}
	default:
		log.Warn("Failed to convert confValue")
	}
	return out
}

func vodConfToConfValue(in *vod.ConfValue) *api.ConfValue {
	var out *api.ConfValue
	switch cval := in.Value.(type) {
	case *vod.ConfValue_SwitchVal:
		out = &api.ConfValue{Value: &api.ConfValue_SwitchVal{SwitchVal: cval.SwitchVal}}
	case *vod.ConfValue_SelectedVal:
		out = &api.ConfValue{Value: &api.ConfValue_SelectedVal{SelectedVal: cval.SelectedVal}}
	default:
		log.Warn("Failed to convert vodConfValue")
	}
	return out
}

func FormatCloud(in *vod.CloudConf, confType api.ConfType) *api.CloudConf {
	if in == nil {
		return nil
	}
	out := &api.CloudConf{ConfType: confType, Show: in.Show}
	if in.FieldValue != nil {
		switch fVal := in.FieldValue.Value.(type) {
		case *vod.FieldValue_Switch:
			out.FieldValue = &api.FieldValue{Value: &api.FieldValue_Switch{Switch: fVal.Switch}}
		}
	}
	if in.ConfValue != nil {
		out.ConfValue = vodConfToConfValue(in.ConfValue)
	}

	return out
}

func FormatArc(in *vod.ArcConf) *api.ArcConf {
	if in == nil {
		return nil
	}
	reply := &api.ArcConf{
		IsSupport:      in.IsSupport,
		Disabled:       in.Disabled,
		UnsupportScene: in.UnsupportScene,
	}
	if in.ExtraContent != nil {
		reply.ExtraContent = &api.ExtraContent{
			DisabledReason: in.ExtraContent.DisabledReason,
			DisabledCode:   in.ExtraContent.DisabledCode,
		}
	}
	return reply
}

func FormatPlayConf(conf *vod.PlayAbilityConf) (rly *api.PlayAbilityConf) {
	rly = &api.PlayAbilityConf{}
	rly.BackgroundPlayConf = FormatCloud(conf.BackgroundPlayConf, api.ConfType_BACKGROUNDPLAY)
	rly.FlipConf = FormatCloud(conf.FlipConf, api.ConfType_FLIPCONF)
	rly.CastConf = FormatCloud(conf.CastConf, api.ConfType_CASTCONF)
	rly.FeedbackConf = FormatCloud(conf.FeedbackConf, api.ConfType_FEEDBACK)
	rly.SubtitleConf = FormatCloud(conf.SubtitleConf, api.ConfType_SUBTITLE)
	rly.PlaybackRateConf = FormatCloud(conf.PlaybackRateConf, api.ConfType_PLAYBACKRATE)
	rly.TimeUpConf = FormatCloud(conf.TimeUpConf, api.ConfType_TIMEUP)
	rly.PlaybackModeConf = FormatCloud(conf.PlaybackModeConf, api.ConfType_PLAYBACKMODE)
	rly.ScaleModeConf = FormatCloud(conf.ScaleModeConf, api.ConfType_SCALEMODE)
	rly.LikeConf = FormatCloud(conf.LikeConf, api.ConfType_LIKE)
	rly.DislikeConf = FormatCloud(conf.DislikeConf, api.ConfType_DISLIKE)
	rly.CoinConf = FormatCloud(conf.CoinConf, api.ConfType_COIN)
	rly.ElecConf = FormatCloud(conf.ElecConf, api.ConfType_ELEC)
	rly.ShareConf = FormatCloud(conf.ShareConf, api.ConfType_SHARE)
	rly.ScreenShotConf = FormatCloud(conf.ScreenShotConf, api.ConfType_SCREENSHOT)
	rly.LockScreenConf = FormatCloud(conf.LockScreenConf, api.ConfType_LOCKSCREEN)
	rly.RecommendConf = FormatCloud(conf.RecommendConf, api.ConfType_RECOMMEND)
	rly.PlaybackSpeedConf = FormatCloud(conf.PlaybackSpeedConf, api.ConfType_PLAYBACKSPEED)
	rly.DefinitionConf = FormatCloud(conf.DefinitionConf, api.ConfType_DEFINITION)
	rly.SelectionsConf = FormatCloud(conf.SelectionsConf, api.ConfType_SELECTIONS)
	rly.NextConf = FormatCloud(conf.NextConf, api.ConfType_NEXT)
	rly.EditDmConf = FormatCloud(conf.EditDmConf, api.ConfType_EDITDM)
	rly.InnerDmConf = FormatCloud(conf.InnerDmConf, api.ConfType_INNERDM)
	rly.OuterDmConf = FormatCloud(conf.OuterDmConf, api.ConfType_OUTERDM)
	rly.ShakeConf = FormatCloud(conf.ShakeConf, api.ConfType_SHAKE)
	rly.SmallWindowConf = FormatCloud(conf.SmallWindowConf, api.ConfType_SMALLWINDOW)
	rly.PanoramaConf = FormatCloud(conf.PanoramaConf, api.ConfType_PANORAMA)
	rly.DolbyConf = FormatCloud(conf.DolbyConf, api.ConfType_DOLBY)
	rly.ColorFilterConf = FormatCloud(conf.ColorFilterConf, api.ConfType_COLORFILTER)
	rly.LossLessConf = FormatCloud(conf.LossLessConf, api.ConfType_LOSSLESS)
	return
}

func ConfConvert(conf []*api.PlayConfState) (rly []*vod.PlayConfState) {
	for _, v := range conf {
		tmp := &vod.PlayConfState{}
		tmp.Show = v.Show
		tmpConfType, matched := confTypeConvert(v.ConfType)
		if !matched {
			continue
		}
		tmp.ConfType = tmpConfType
		if v.FieldValue != nil {
			switch fVal := v.FieldValue.Value.(type) {
			case *api.FieldValue_Switch:
				tmp.FieldValue = &vod.FieldValue{Value: &vod.FieldValue_Switch{Switch: fVal.Switch}}
			}
		}
		if v.ConfValue != nil {
			tmp.ConfValue = confValueToVodConf(v.ConfValue)
		}
		rly = append(rly, tmp)
	}
	return
}

// CloudEditParam .
type CloudEditParam struct {
	Buvid    string `json:"buvid"`
	Platform string `json:"platform"`
	Brand    string `json:"brand"`
	Model    string `json:"model"`
	Build    int64  `json:"build"`
}

func setQnAttr(attr int64, qn uint32) int64 {
	var res int64
	if qn == QnHDR {
		res = attr | 1<<AttrIsHDR
	}
	if qn == QnDolbyHDR {
		res = attr | 1<<AttrIsDolbyHDR
	}
	return res
}

func (pl *PlayurlHlsReply) FormatPlayHls(plVod *vod.HlsResponseMsg, arg *ParamHls, key, secret string) {
	pl.Type = int32(plVod.Type)
	pl.Quality = plVod.Quality
	pl.Format = plVod.Format
	pl.Timelength = plVod.Timelength
	pl.VideoCodecid = plVod.VideoCodecid
	pl.VideoProject = plVod.VideoProject
	// 清晰度列表
	for _, tVal := range plVod.SupportFormats {
		if tVal == nil {
			continue
		}
		tmpStresm := &api.FormatDescription{
			Quality:        tVal.Quality,
			Format:         tVal.Format,
			NewDescription: tVal.NewDescription,
			DisplayDesc:    tVal.DisplayDesc,
			Superscript:    tVal.Superscript,
		}
		pl.SupportFormats = append(pl.SupportFormats, tmpStresm)
	}
	// 拼接url
	if plVod.Type == vod.ResponseType_HLS {
		pl.URL = _baseUri + "/hls/master.m3u8?" + formatNextRequest(arg, int64(plVod.Quality), key, secret, vod.QnCategory_MixType)
	}
	for _, v := range plVod.Durl {
		if v == nil {
			continue
		}
		backupURL := v.BackupUrl
		if len(v.BackupUrl) > _maxBkNum {
			backupURL = v.BackupUrl[:_maxBkNum]
		}
		//不支持下载所以没有md5参数
		pl.Durl = append(pl.Durl, &api.ResponseUrl{
			Order:     v.Order,
			Length:    v.Length,
			Size_:     v.Size_,
			Url:       v.Url,
			BackupUrl: backupURL,
		})
	}
}

func (pl *HlsMasterReply) FormatPlayMaster(plVod *vod.MasterScheduler, arg *ParamHls, key, secret string) {
	audioUrl := _baseUri + "/hls/stream.m3u8?" + formatNextRequest(arg, int64(plVod.Audio.Qn), key, secret, vod.QnCategory_Audio)
	videoUrl := _baseUri + "/hls/stream.m3u8?" + formatNextRequest(arg, int64(plVod.Video.Qn), key, secret, vod.QnCategory_Video)
	str := `#EXTM3U
#EXT-X-VERSION:6
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="%d",NAME="%s",DEFAULT=YES,URI="%s"
#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s,CODECS="%s,%s",AUDIO="%d"
%s`
	lastStr := fmt.Sprintf(str, plVod.Audio.GroupId, plVod.Video.Name, audioUrl, plVod.Video.Bandwidth, plVod.Video.Resolution, plVod.Video.Codecs, plVod.Audio.Codecs, plVod.Audio.GroupId, videoUrl)
	pl.M3u8Data = []byte(lastStr)
}

// 没有音频的时候使用新模板
func (pl *HlsMasterReply) FormatPlayNoAudioMaster(plVod *vod.MasterScheduler, arg *ParamHls, key, secret string) {
	videoUrl := _baseUri + "/hls/stream.m3u8?" + formatNextRequest(arg, int64(plVod.Video.Qn), key, secret, vod.QnCategory_Video)
	str := `#EXTM3U
#EXT-X-VERSION:6
#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s,CODECS="%s"
%s`
	lastStr := fmt.Sprintf(str, plVod.Video.Bandwidth, plVod.Video.Resolution, plVod.Video.Codecs, videoUrl)
	pl.M3u8Data = []byte(lastStr)
}

func (pl *HlsMasterReply) FormatMultPlayMaster(plVod *vod.MasterScheduler, arg *ParamHls, key, secret string) {
	var lastStr []string
	lastStr = append(lastStr, "#EXTM3U")
	lastStr = append(lastStr, "#EXT-X-VERSION:6")
	videoUrl := _baseUri + "/hls/stream.m3u8?%s"
	//audio为空，使用新模板
	if plVod.Audio == nil {
		extStream := `#EXT-X-STREAM-INF:BANDWIDTH=%d,AVERAGE-BANDWIDTH=%d,RESOLUTION=%s,CODECS="%s",FRAME-RATE=%s`
		for _, v := range plVod.Videos {
			if v == nil {
				continue
			}
			//拼接ext stream
			lastStr = append(lastStr, fmt.Sprintf(extStream, v.Bandwidth, v.AverageBandwidth, v.Resolution, v.Codecs, v.FrameRate))
			//拼接videourl
			lastStr = append(lastStr, fmt.Sprintf(videoUrl, formatNextRequest(arg, int64(v.Qn), key, secret, vod.QnCategory_Video)))
		}
	} else {
		audioUrl := _baseUri + "/hls/stream.m3u8?" + formatNextRequest(arg, int64(plVod.Audio.Qn), key, secret, vod.QnCategory_Audio)
		//拼接音屏地址
		audioExt := `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="%d",NAME="%d",DEFAULT=YES,URI="%s"`
		lastStr = append(lastStr, fmt.Sprintf(audioExt, plVod.Audio.Qn, plVod.Audio.Qn, audioUrl))
		extStream := `#EXT-X-STREAM-INF:BANDWIDTH=%d,AVERAGE-BANDWIDTH=%d,RESOLUTION=%s,CODECS="%s,%s",FRAME-RATE=%s,AUDIO="%d"`
		for _, v := range plVod.Videos {
			if v == nil {
				continue
			}
			//拼接ext stream
			lastStr = append(lastStr, fmt.Sprintf(extStream, v.Bandwidth, v.AverageBandwidth, v.Resolution, v.Codecs, plVod.Audio.Codecs, v.FrameRate, plVod.Audio.Qn))
			//拼接videourl
			lastStr = append(lastStr, fmt.Sprintf(videoUrl, formatNextRequest(arg, int64(v.Qn), key, secret, vod.QnCategory_Video)))
		}
	}
	pl.M3u8Data = []byte(strings.Join(lastStr, "\n"))
}

type BubbleParam struct {
	Aid      int64  `form:"aid"`
	Cid      int64  `form:"cid"`
	SeasonId int64  `form:"season_id"`
	EpId     int64  `form:"ep_id"`
	MobiApp  string `form:"mobi_app"`
	Build    int32  `form:"build"`
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

type ProjPageParam struct {
	PlayurlType int64  `form:"playurl_type"` // 视频类型 1:ugc  2:pgc  3:pugv
	Aid         int64  `form:"aid"`
	Cid         int64  `form:"cid"`
	EpId        int64  `form:"ep_id"`
	SeasonId    int64  `form:"season_id"`
	Mid         int64  `form:"-"`
	Channel     string `form:"channel"`
	Platform    string `form:"platform"`
	MobiApp     string `form:"mobi_app"`
	Build       int64  `form:"build"`
}

type ProjActAllParam struct {
	ActTypeBits int64  `form:"act_type"`     // 活动类型bit位描述 0位：投屏红点 1位：控制页icon 2位：设备列表banner
	PlayurlType int64  `form:"playurl_type"` // 视频类型 1:ugc  2:pgc  3:pugv 4:live
	Aid         int64  `form:"aid"`
	Cid         int64  `form:"cid"`
	SeasonId    int64  `form:"season_id"`
	EpId        int64  `form:"ep_id"`
	RoomId      int64  `form:"room_id"`
	PartitionId int64  `form:"partition_id"` // 稿件分区id
	Mid         int64  `form:"-"`
	NewUser     int64  `form:"new_user"` // 是否为投屏新用户
	Build       int64  `form:"build"`
	MobiApp     string `form:"mobi_app"`
	Channel     string `form:"channel"`
	Platform    string `form:"platform"`
}

type AuthArcParam struct {
	PlayurlType int64 // 视频类型 1:ugc  2:pgc  3:pugv 4:live
	Aid         int64
	Cid         int64
	SeasonId    int64
	EpId        int64
	RoomId      int64
}
