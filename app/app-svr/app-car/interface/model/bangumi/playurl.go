package bangumi

// 参考：https://git.bilibili.co/bapis/bapis/blob/master/video/vod/playurlpgc/service.proto
type PlayInfo struct {
	//返回视频的清晰度
	Quality uint32 `json:"quality"`
	//返回视频的格式
	Format string `json:"format"`
	//返回视频的总时长, 单位为ms
	Timelength uint64 `json:"timelength"`
	//返回视频的编码号
	VideoCodecid uint32 `json:"video_codecid"`
	IsPreview    uint32 `json:"is_preview"`
	Fnval        uint32 `json:"fnval"`
	//透传返回请求的fnval
	Fnver uint32 `json:"fnver"`
	//返回视频的是否支持投影
	VideoProject bool `json:"video_project"`
	//返回视频播放url的列表，有durl则没dash字段
	Durl []*ResponseUrl `json:"durl"`
	//返回DASH视频的MPD格式文件,有dash则没durl字段
	Dash *ResponseDash `json:"dash"`
	//返回视频的拥有的清晰度列表
	AcceptQuality []uint32 `json:"accept_quality"`
	//返回视频拥有的格式列表
	SupportFormats []*FormatDescription `json:"support_formats"`
	NoRexcode      int32                `json:"no_rexcode"`
	//表示返回flv url 或是 dash url
	VideoType string `json:"type"`
}

type ResponseUrl struct {
	Order     uint32   `json:"order"`
	Length    uint64   `json:"length"`
	Size      uint64   `json:"size"`
	Url       string   `json:"url"`
	BackupUrl []string `json:"backup_url"`
	MD5       string   `json:"md5"`
}

type ResponseDash struct {
	Duration      uint32  `json:"duration"`
	MinBufferTime float32 `json:"min_buffer_time"`
	//dash视频信息
	Video []*DashItem `json:"video"`
	Audio []*DashItem `json:"audio"`
	//杜比音频
	Dolby *DolbyItem `json:"dolby"`
}

type DolbyItem struct {
	//// NONE
	//NONE = 0;
	//// 普通杜比音效
	//COMMON = 1;
	//// 全景杜比音效
	//ATMOS = 2;
	Type  string      `json:"type"`
	Audio []*DashItem `json:"audio"`
}

const (
	Dolby_NO     = 0
	Dolby_COMMON = 1
	Dolby_ATMOS  = 2
)

func (di *DolbyItem) GetTypeToInt() int32 {
	if di.Type == "ATMOS" {
		return Dolby_ATMOS
	}
	if di.Type == "COMMON" {
		return Dolby_COMMON
	}
	return Dolby_NO
}

type DashItem struct {
	ID           uint32           `json:"id"`
	BaseURL      string           `json:"base_url"`
	BackupURL    []string         `json:"backup_url"`
	BandWidth    uint32           `json:"bandwidth"`
	MimeType     string           `json:"mime_type"`
	Codecs       string           `json:"codecs"`
	Width        uint32           `json:"width"`
	Height       uint32           `json:"height"`
	FrameRate    string           `json:"frame_rate"`
	Sar          string           `json:"sar"`
	StartWithSap uint32           `json:"start_with_sap"`
	SegmentBase  *DashSegmentBase `json:"segment_base"`
	CodecID      uint32           `json:"codecid"`
	MD5          string           `json:"md5"`
	Size         uint64           `json:"size"`
}

type DashSegmentBase struct {
	Initialization string `json:"initialization"`
	IndexRange     string `json:"index_range"`
}

type FormatDescription struct {
	Quality        uint32 `json:"quality"`
	Format         string `json:"format"`
	Description    string `json:"description"`
	NewDescription string `json:"new_description"`
	DisplayDesc    string `json:"display_desc"`
	Superscript    string `json:"superscript"`
}
