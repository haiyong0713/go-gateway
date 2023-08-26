package model

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	playurlapi "go-gateway/app/app-svr/app-player/interface/api/playurl"
	"go-gateway/app/app-svr/playurl/service/api"
	vod "go-gateway/app/app-svr/playurl/service/api/v2"
)

const _baseURI = "https://api.bilibili.com/x/player"

// PlayurlArg playurl arg.
type PlayurlArg struct {
	Cid           int64  `form:"cid" validate:"min=1"`
	Aid           int64  `form:"avid"`
	Bvid          string `form:"bvid"`
	Qn            int64  `form:"qn"`
	Type          string `form:"type"`
	MaxBackup     int    `form:"max_backup"`
	Npcybs        int32  `form:"npcybs"`
	Platform      string `form:"platform"`
	Buvid         string `form:"buvid"`
	Resolution    string `form:"resolution"`
	Model         string `form:"model"`
	Build         int32  `form:"build"`
	Fnver         int32  `form:"fnver"`
	Fnval         int32  `form:"fnval"`
	Session       string `form:"session"`
	HTML5         int    `form:"html5"`
	Fourk         int    `form:"fourk"`
	H5GoodQuality int    `form:"h5_good_quality"`
	HighQuality   int    `form:"high_quality"`
	VoiceBalance  int64  `form:"voice_balance"`
}

// PlayurlRes playurl res.
type PlayurlRes struct {
	From              string               `json:"from"`
	Result            string               `json:"result"`
	Message           string               `json:"message"`
	Quality           uint32               `json:"quality"`
	Format            string               `json:"format"`
	Timelength        uint64               `json:"timelength"`
	AcceptFormat      string               `json:"accept_format"`
	AcceptDescription []string             `json:"accept_description"`
	AcceptQuality     []uint32             `json:"accept_quality"`
	VideoCodeCid      uint32               `json:"video_codecid"`
	SeekParam         string               `json:"seek_param"`
	SeekType          string               `json:"seek_type"`
	Abtid             int32                `json:"abtid,omitempty"`
	Durl              []*durl              `json:"durl,omitempty"`
	Dash              *dash                `json:"dash,omitempty"`
	SupportFormats    []*formatDescription `json:"support_formats"`
	HighFormat        *formatDescription   `json:"high_format"`
	Volume            *VolumeInfo          `json:"volume,omitempty"`
	LastPlayTime      int64                `json:"last_play_time"` // 上次观看进度
	LastPlayCid       int64                `json:"last_play_cid"`
}

type formatDescription struct {
	Quality        uint32   `json:"quality"`
	Format         string   `json:"format"`
	NewDescription string   `json:"new_description"`
	DisplayDesc    string   `json:"display_desc"`
	Superscript    string   `json:"superscript"`
	Codecs         []string `json:"codecs"`
}

type durl struct {
	Order     uint32   `json:"order"`
	Length    uint64   `json:"length"`
	Size      uint64   `json:"size"`
	Ahead     string   `json:"ahead"`
	Vhead     string   `json:"vhead"`
	URL       string   `json:"url"`
	BackupURL []string `json:"backup_url"`
}

type dash struct {
	Duration       uint32         `json:"duration"`
	MinBufferTime  float32        `json:"minBufferTime"`
	MinBufferTime2 float32        `json:"min_buffer_time"`
	Video          []*dashItem    `json:"video"`
	Audio          []*dashItem    `json:"audio"`
	Dolby          *vod.DolbyItem `json:"dolby"`
	FLAC           *flac          `json:"flac"`
}

type flac struct {
	Display bool      `json:"display"`
	Audio   *dashItem `json:"audio"`
}

type dashItem struct {
	ID            uint32        `json:"id"`
	BaseURL       string        `json:"baseUrl"`
	BaseURL2      string        `json:"base_url"`
	BackupURL     []string      `json:"backupUrl"`
	BackupURL2    []string      `json:"backup_url"`
	Bandwidth     uint32        `json:"bandwidth"`
	MimeType      string        `json:"mimeType"`
	MimeType2     string        `json:"mime_type"`
	Codecs        string        `json:"codecs"`
	Width         uint32        `json:"width"`
	Height        uint32        `json:"height"`
	FrameRate     string        `json:"frameRate"`
	FrameRate2    string        `json:"frame_rate"`
	Sar           string        `json:"sar"`
	StartWithSAP  uint32        `json:"startWithSap"`
	StartWithSAP2 uint32        `json:"start_with_sap"`
	SegmentBase   *segmentBase  `json:"SegmentBase"`
	SegmentBase2  *segmentBase2 `json:"segment_base"`
	Codecid       uint32        `json:"codecid"`
}

type segmentBase struct {
	Initialization string `json:"Initialization"`
	IndexRange     string `json:"indexRange"`
}

type segmentBase2 struct {
	Initialization string `json:"initialization"`
	IndexRange     string `json:"index_range"`
}

type VolumeInfo struct {
	MeasuredI         float64 `json:"measured_i"`
	MeasuredLra       float64 `json:"measured_lra"`
	MeasuredTp        float64 `json:"measured_tp"`
	MeasuredThreshold float64 `json:"measured_threshold"`
	TargetOffset      float64 `json:"target_offset"`
	TargetI           float64 `json:"target_i"`
	TargetTp          float64 `json:"target_tp"`
}

// FromPlayurl from playurl data.
func (p *PlayurlRes) FromPlayurl(reply *api.PlayURLReply) {
	p.From = reply.From
	p.Result = reply.Result
	p.Quality = reply.Quality
	p.Format = reply.Format
	p.Timelength = reply.Timelength
	p.AcceptFormat = reply.AcceptFormat
	p.AcceptDescription = reply.AcceptDescription
	p.AcceptQuality = reply.AcceptQuality
	p.VideoCodeCid = reply.VideoCodecid
	p.SeekParam = reply.SeekParam
	p.SeekType = reply.SeekType
	p.Abtid = reply.Abtid
	for _, v := range reply.Durl {
		if v == nil {
			continue
		}
		durlItem := new(durl)
		durlItem.fromDurl(v)
		p.Durl = append(p.Durl, durlItem)
	}
	if reply.Dash != nil {
		pDash := new(dash)
		pDash.fromDash(reply.Dash)
		p.Dash = pDash
	}
}

// FromDurl from durl data.
func (d *durl) fromDurl(item *api.Durl) {
	d.Order = item.Order
	d.Length = item.Length
	d.Size = item.Size_
	d.Ahead = item.Ahead
	d.Vhead = item.Vhead
	d.URL = item.Url
	d.BackupURL = item.BackupUrl
}

// FromDash from dash data.
func (d *dash) fromDash(item *api.Dash) {
	d.Duration = item.Duration
	d.MinBufferTime = item.MinBufferTime
	d.MinBufferTime2 = item.MinBufferTime
	for _, v := range item.Video {
		if v == nil {
			continue
		}
		videoItem := new(dashItem)
		videoItem.fromDashItem(v)
		d.Video = append(d.Video, videoItem)
	}
	for _, v := range item.Audio {
		if v == nil {
			continue
		}
		audioItem := new(dashItem)
		audioItem.fromDashItem(v)
		d.Audio = append(d.Audio, audioItem)
	}
}

// FromDashItem from dash item.
func (d *dashItem) fromDashItem(item *api.DashItem) {
	d.ID = item.Id
	d.BaseURL = item.BaseUrl
	d.BaseURL2 = item.BaseUrl
	d.BackupURL = item.BackupUrl
	d.BackupURL2 = item.BackupUrl
	d.Bandwidth = item.Bandwidth
	d.MimeType = item.MimeType
	d.MimeType2 = item.MimeType
	d.Codecs = item.Codecs
	d.Width = item.Width
	d.Height = item.Height
	d.FrameRate = item.FrameRate
	d.FrameRate2 = item.FrameRate
	d.Sar = item.Sar
	d.StartWithSAP = item.StartWithSAP
	d.StartWithSAP2 = item.StartWithSAP
	if item.SegmentBase != nil {
		d.SegmentBase = &segmentBase{
			Initialization: item.SegmentBase.GetInitialization(),
			IndexRange:     item.SegmentBase.GetIndexRange(),
		}
		d.SegmentBase2 = &segmentBase2{
			Initialization: item.SegmentBase.GetInitialization(),
			IndexRange:     item.SegmentBase.GetIndexRange(),
		}
	}
	d.Codecid = item.Codecid
}

type ParamHls struct {
	AID int64 `form:"aid" validate:"min=1"`
	CID int64 `form:"cid" validate:"min=1"`
	// qn清晰度
	Qn int64 `form:"qn" validate:"min=0"`
	// fnver和fnval标识视频格式
	Fnver int32 `form:"fnver"`
	Fnval int32 `form:"fnval"`
	// 返回url是否强制使用域名(非ip地址), 1-http域名 2-https域名
	ForceHost int32  `form:"force_host"`
	Buvid     string `form:"-"` //从header获取
	// 投屏设备 默认其他=0，OTT设备=1
	DeviceType int32 `form:"device_type"`
	// 0:投屏 1:画中画
	RequestType int32 `form:"request_type"`
	// 0:mix 1:video 2:audio
	QnCategory int64  `form:"qn_category"`
	Dolby      int64  `form:"dolby"`
	Platform   string `form:"platform"`
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
	Durl []*playurlapi.ResponseUrl `json:"durl,omitempty"`
	//返回视频拥有的格式列表
	SupportFormats []*playurlapi.FormatDescription `json:"support_formats,omitempty"`
}

type HlsMasterReply struct {
	M3u8Data []byte `json:"m3u8_data"`
}

func formatNextRequest(in *ParamHls, qn int64, qnCategory vod.QnCategory) string {
	out := url.Values{}
	out.Set("aid", strconv.FormatInt(in.AID, 10))
	out.Set("cid", strconv.FormatInt(in.CID, 10))
	out.Set("qn", strconv.FormatInt(qn, 10))
	out.Set("fnver", strconv.FormatInt(int64(in.Fnver), 10))
	out.Set("fnval", strconv.FormatInt(int64(in.Fnval), 10))
	out.Set("platform", in.Platform)
	out.Set("force_host", strconv.FormatInt(int64(in.ForceHost), 10))
	out.Set("device_type", strconv.FormatInt(int64(in.DeviceType), 10))
	out.Set("request_type", strconv.FormatInt(int64(in.RequestType), 10))
	out.Set("qn_category", strconv.FormatInt(int64(qnCategory), 10))
	out.Set("dolby", strconv.FormatInt(in.Dolby, 10))
	return query(out)
}

func query(params url.Values) (query string) {
	if params == nil {
		params = url.Values{}
	}
	tmp := params.Encode()
	if strings.IndexByte(tmp, '+') > -1 {
		tmp = strings.Replace(tmp, "+", "%20", -1)
	}
	// query
	var qb bytes.Buffer
	qb.WriteString(tmp)
	query = qb.String()
	return
}

// nolint:gomnd
func (pl *PlayurlHlsReply) FormatPlayHls(plVod *vod.HlsResponseMsg, arg *ParamHls) {
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
		tmpStresm := &playurlapi.FormatDescription{
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
		pl.URL = _baseURI + "/hls/master.m3u8?" + formatNextRequest(arg, int64(plVod.Quality), vod.QnCategory_MixType)
	}
	for _, v := range plVod.Durl {
		if v == nil {
			continue
		}
		backupURL := v.BackupUrl
		if len(v.BackupUrl) > 2 {
			backupURL = v.BackupUrl[:2]
		}
		//不支持下载所以没有md5参数
		pl.Durl = append(pl.Durl, &playurlapi.ResponseUrl{
			Order:     v.Order,
			Length:    v.Length,
			Size_:     v.Size_,
			Url:       v.Url,
			BackupUrl: backupURL,
		})
	}
}

func (pl *HlsMasterReply) FormatPlayMaster(plVod *vod.MasterScheduler, arg *ParamHls) {
	audioURL := _baseURI + "/hls/stream.m3u8?" + formatNextRequest(arg, int64(plVod.Audio.Qn), vod.QnCategory_Audio)
	videoURL := _baseURI + "/hls/stream.m3u8?" + formatNextRequest(arg, int64(plVod.Video.Qn), vod.QnCategory_Video)
	str := `#EXTM3U
#EXT-X-VERSION:6
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="%d",NAME="%s",DEFAULT=YES,URI="%s"
#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s,CODECS="%s,%s",AUDIO="%d"
%s`
	lastStr := fmt.Sprintf(str, plVod.Audio.GroupId, plVod.Video.Name, audioURL, plVod.Video.Bandwidth, plVod.Video.Resolution, plVod.Video.Codecs, plVod.Audio.Codecs, plVod.Audio.GroupId, videoURL)
	pl.M3u8Data = []byte(lastStr)
}

// 没有音频的时候使用新模板
func (pl *HlsMasterReply) FormatPlayNoAudioMaster(plVod *vod.MasterScheduler, arg *ParamHls) {
	videoURL := _baseURI + "/hls/stream.m3u8?" + formatNextRequest(arg, int64(plVod.Video.Qn), vod.QnCategory_Video)
	str := `#EXTM3U
#EXT-X-VERSION:6
#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s,CODECS="%s"
%s`
	lastStr := fmt.Sprintf(str, plVod.Video.Bandwidth, plVod.Video.Resolution, plVod.Video.Codecs, videoURL)
	pl.M3u8Data = []byte(lastStr)
}
