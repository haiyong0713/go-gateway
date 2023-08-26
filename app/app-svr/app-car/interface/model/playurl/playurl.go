package playurl

import (
	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/bangumi"
	v2 "go-gateway/app/app-svr/playurl/service/api/v2"
)

const (
	_flv  = 1
	_dash = 2
	_mp4  = 3
)

type Param struct {
	model.DeviceInfo
	AccessKey string `form:"access_key"`
	MobiApp   string `form:"mobi_app"`
	Platform  string `form:"platform"`
	Device    string `form:"device"`
	Build     int    `form:"build"`
	Oid       int64  `form:"oid"`
	Cid       int64  `form:"cid"`
	Aid       int64  `form:"aid"`
	Otype     string `form:"otype"`
	VideoType string `form:"type"`
	//清晰度
	Qn    int64 `form:"qn"`
	Fnver int   `form:"fnver"`
	//每位(为1)标识一个功能,256是否需要杜比音频，此位为0，代表不需要杜比音频，为1，代表需要杜比音频
	Fnval        int   `form:"fnval"`
	Fourk        int   `form:"fourk"`
	ForceHost    int   `form:"force_host"`
	IsPreview    int   `form:"is_preview"`
	SeasonID     int64 `form:"season_id"`
	EpID         int64 `form:"ep_id"`
	IsDazhongcar bool  `form:"is_dazhongcar" default:"false"`
	NetType      int32
	TfType       int32
	//用于传递 从ctx中获取
	Mid   int64
	Buvid string
}

type Info struct {
	Quality        uint32               `json:"quality,omitempty"`
	Format         string               `json:"format,omitempty"`
	Timelength     uint64               `json:"timelength,omitempty"`
	VideoCodecid   uint32               `json:"video_codecid,omitempty"`
	IsPreview      uint32               `json:"is_preview,omitempty"`
	Fnval          uint32               `json:"fnval,omitempty"`
	Fnver          uint32               `json:"fnver,omitempty"`
	VideoProject   bool                 `json:"video_project,omitempty"`
	Durl           []*ResponseUrl       `json:"durl,omitempty"`
	Dash           *ResponseDash        `json:"dash,omitempty"`
	NoRexcode      int32                `json:"no_rexcode,omitempty"`
	AcceptQuality  []uint32             `json:"accept_quality,omitempty"`
	SupportFormats []*FormatDescription `json:"support_formats,omitempty"`
	VideoType      string               `json:"type,omitempty"`
	//是否支持杜比 true/false
	Dolby    bool `json:"dolby,omitempty"`
	DolbyLog bool
}

type ResponseUrl struct {
	Order     uint32   `json:"order,omitempty"`
	Length    uint64   `json:"length,omitempty"`
	Size      uint64   `json:"size,omitempty"`
	URL       string   `json:"url,omitempty"`
	BackupURL []string `json:"backup_url,omitempty"`
	MD5       string   `json:"md5,omitempty"`
}

type ResponseDash struct {
	Duration      uint32      `json:"duration,omitempty"`
	MinBufferTime float32     `json:"min_buffer_time,omitempty"`
	Video         []*DashItem `json:"video,omitempty"`
	Audio         []*DashItem `json:"audio,omitempty"`
	Dolby         *DolbyItem  `json:"dolby,omitempty"`
}

type DolbyItem struct {
	Type  int32       `json:"type,omitempty"`
	Audio []*DashItem `json:"audio,omitempty"`
}

type DashItem struct {
	ID           uint32           `json:"id,omitempty"`
	BaseURL      string           `json:"base_url,omitempty"`
	BackupURL    []string         `json:"backup_url,omitempty"`
	BandWidth    uint32           `json:"bandwidth,omitempty"`
	MimeType     string           `json:"mime_type,omitempty"`
	Codecs       string           `json:"codecs,omitempty"`
	Width        uint32           `json:"width,omitempty"`
	Height       uint32           `json:"height,omitempty"`
	FrameRate    string           `json:"frame_rate,omitempty"`
	Sar          string           `json:"sar,omitempty"`
	StartWithSap uint32           `json:"start_with_sap,omitempty"`
	SegmentBase  *DashSegmentBase `json:"segment_base,omitempty"`
	CodecID      uint32           `json:"codecid,omitempty"`
	MD5          string           `json:"md5,omitempty"`
	Size         uint64           `json:"size,omitempty"`
}

type DashSegmentBase struct {
	Initialization string `json:"initialization,omitempty"`
	IndexRange     string `json:"index_range,omitempty"`
}

type FormatDescription struct {
	Quality     uint32 `json:"quality,omitempty"`
	Format      string `json:"format,omitempty"`
	Description string `json:"description,omitempty"`
	DisplayDesc string `json:"display_desc,omitempty"`
	Superscript string `json:"superscript,omitempty"`
}

func (i *Info) FromPGC(p *bangumi.PlayInfo, isFilter bool) {
	if p == nil {
		return
	}
	dashIDm := map[uint32]struct{}{}
	i.Quality = p.Quality
	i.Format = p.Format
	i.Timelength = p.Timelength
	i.VideoCodecid = p.VideoCodecid
	i.IsPreview = p.IsPreview
	i.Fnval = p.Fnval
	i.Fnver = p.Fnver
	i.VideoProject = p.VideoProject
	log.Info("dolby log :%+v", p)
	for _, v := range p.Durl {
		durl := &ResponseUrl{
			Order:     v.Order,
			Length:    v.Length,
			Size:      v.Size,
			URL:       v.Url,
			BackupURL: v.BackupUrl,
			MD5:       v.MD5,
		}
		i.Durl = append(i.Durl, durl)
	}
	if p.Dash != nil {
		var video, audio []*DashItem
		var dolby *DolbyItem
		for _, v := range p.Dash.Video {
			// linux 干掉1080 30帧以上的
			if isFilter && v.ID > 80 {
				continue
			}
			item := &DashItem{}
			item.fromPGCDash(v)
			video = append(video, item)
			dashIDm[v.ID] = struct{}{}
		}
		for _, v := range p.Dash.Audio {
			item := &DashItem{}
			item.fromPGCDash(v)
			audio = append(audio, item)
		}
		log.Info("dolby log :%+v", p.Dash.Dolby)
		i.Dolby = false
		if p.Dash.Dolby != nil {
			dolby = &DolbyItem{Type: p.Dash.Dolby.GetTypeToInt()}
			log.Info("dolby log :%+v", p.Dash.Dolby)
			var dolbyAudio []*DashItem
			for _, v := range p.Dash.Dolby.Audio {
				log.Info("dolby log :%+v", v)
				item := &DashItem{}
				item.fromPGCDash(v)
				dolbyAudio = append(dolbyAudio, item)
			}
			if dolby.Type > 0 {
				i.Dolby = true
			}
			dolby.Audio = dolbyAudio
			if len(dolbyAudio) > 0 {
				i.DolbyLog = true
			}
		}

		if len(video) > 0 || len(audio) > 0 {
			i.Dash = &ResponseDash{
				Video:         video,
				Audio:         audio,
				Dolby:         dolby,
				Duration:      p.Dash.Duration,
				MinBufferTime: p.Dash.MinBufferTime,
			}
		}
	}
	for _, v := range p.SupportFormats {
		if _, ok := dashIDm[v.Quality]; !ok && isFilter {
			continue
		}
		sf := &FormatDescription{
			Quality:     v.Quality,
			Format:      v.Format,
			Description: v.NewDescription,
			DisplayDesc: v.DisplayDesc,
			Superscript: v.Superscript,
		}
		i.SupportFormats = append(i.SupportFormats, sf)
	}
	for _, v := range p.AcceptQuality {
		if _, ok := dashIDm[v]; !ok && isFilter {
			continue
		}
		i.AcceptQuality = append(i.AcceptQuality, v)
	}
	i.NoRexcode = p.NoRexcode
	i.VideoType = p.VideoType
}

func (i *Info) FromUGC(p *v2.ResponseMsg, isFilter bool) {
	if p == nil {
		return
	}
	dashIDm := map[uint32]struct{}{}
	i.Quality = p.Quality
	i.Format = p.Format
	i.Timelength = p.Timelength
	i.VideoCodecid = p.VideoCodecid
	i.Fnval = p.Fnval
	i.Fnver = p.Fnver
	i.VideoProject = p.VideoProject
	for _, v := range p.Durl {
		durl := &ResponseUrl{
			Order:     v.Order,
			Length:    v.Length,
			Size:      v.Size_,
			URL:       v.Url,
			BackupURL: v.BackupUrl,
			MD5:       v.Md5,
		}
		i.Durl = append(i.Durl, durl)
	}
	if p.Dash != nil {
		var video, audio []*DashItem
		var dolby *DolbyItem
		for _, v := range p.Dash.Video {
			// linux 干掉1080 30帧以上的
			if isFilter && v.Id > 80 {
				continue
			}
			item := &DashItem{}
			item.fromUGCDash(v)
			video = append(video, item)
			dashIDm[v.Id] = struct{}{}
		}
		for _, v := range p.Dash.Audio {
			item := &DashItem{}
			item.fromUGCDash(v)
			audio = append(audio, item)
		}
		log.Info("dolby log :%+v", p.Dash)
		log.Info("dolby log :%+v", p.Dash.Dolby)
		if p.Dash.Dolby != nil {
			dolby = &DolbyItem{Type: int32(p.Dash.Dolby.Type)}
			log.Info("dolby log :%+v", p.Dash.Dolby)
			var dolbyAudio []*DashItem
			for _, v := range p.Dash.Dolby.Audio {
				log.Info("dolby log v :%+v", v)
				item := &DashItem{}
				item.fromUGCDash(v)
				dolbyAudio = append(dolbyAudio, item)
			}
			if dolby.Type > 0 {
				i.Dolby = true
			}
			dolby.Audio = dolbyAudio
			if len(dolbyAudio) > 0 {
				i.DolbyLog = true
			}

		}
		if len(video) > 0 || len(audio) > 0 {
			i.Dash = &ResponseDash{
				Video:         video,
				Audio:         audio,
				Dolby:         dolby,
				Duration:      p.Dash.Duration,
				MinBufferTime: p.Dash.MinBufferTime,
			}
		}
	}
	for _, v := range p.SupportFormats {
		if _, ok := dashIDm[v.Quality]; !ok && isFilter {
			continue
		}
		sf := &FormatDescription{
			Quality:     v.Quality,
			Format:      v.Format,
			Description: v.NewDescription,
			DisplayDesc: v.DisplayDesc,
			Superscript: v.Superscript,
		}
		i.SupportFormats = append(i.SupportFormats, sf)
	}
	for _, v := range p.AcceptQuality {
		if _, ok := dashIDm[v]; !ok && isFilter {
			continue
		}
		i.AcceptQuality = append(i.AcceptQuality, v)
	}
	i.NoRexcode = p.NoRexcode
	i.VideoType = fromType(p.Type)
}

func (i *DashItem) fromUGCDash(dash *v2.DashItem) {
	i.ID = dash.Id
	i.BaseURL = dash.BaseUrl
	i.BackupURL = dash.BackupUrl
	i.BandWidth = dash.Bandwidth
	i.MimeType = dash.MimeType
	i.Codecs = dash.Codecs
	i.Width = dash.Width
	i.Height = dash.Height
	i.FrameRate = dash.FrameRate
	i.Sar = dash.Sar
	i.StartWithSap = dash.StartWithSap
	if dash.SegmentBase != nil {
		i.SegmentBase = &DashSegmentBase{
			Initialization: dash.SegmentBase.Initialization,
			IndexRange:     dash.SegmentBase.IndexRange,
		}
	}
	i.CodecID = dash.Codecid
	i.MD5 = dash.Md5
	i.Size = dash.Size_
}

func (i *DashItem) fromPGCDash(dash *bangumi.DashItem) {
	i.ID = dash.ID
	i.BaseURL = dash.BaseURL
	i.BackupURL = dash.BackupURL
	i.BandWidth = dash.BandWidth
	i.MimeType = dash.MimeType
	i.Codecs = dash.Codecs
	i.Width = dash.Width
	i.Height = dash.Height
	i.FrameRate = dash.FrameRate
	i.Sar = dash.Sar
	i.StartWithSap = dash.StartWithSap
	if dash.SegmentBase != nil {
		i.SegmentBase = &DashSegmentBase{
			Initialization: dash.SegmentBase.Initialization,
			IndexRange:     dash.SegmentBase.IndexRange,
		}
	}
	i.CodecID = dash.CodecID
	i.MD5 = dash.MD5
	i.Size = dash.Size
}

func fromType(vType int32) string {
	switch vType {
	case _flv:
		return "flv"
	case _dash:
		return "dash"
	case _mp4:
		return "mp4"
	default:
		return ""
	}
}
