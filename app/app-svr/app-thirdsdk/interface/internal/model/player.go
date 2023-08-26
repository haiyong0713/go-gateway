package model

type PlayURLParam struct {
	Bvid            string   `form:"bvid" validate:"required"`
	Qn              uint32   `form:"qn" validate:"min=0"`
	MobiApp         string   `form:"mobi_app"`
	Fnver           uint32   `form:"fnver"`
	Fnval           uint32   `form:"fnval"`
	PreferCodecType CodeType `form:"prefer_codec_type"`
	ForceHost       uint32   `form:"force_host"`
	SDKVersion      string   `form:"sdk_version"`
	SDKIdentifier   string   `form:"sdk_identifier"`
	Device          string   `form:"device"`
	Platform        string   `form:"platform"`
	PreferCodecID   uint32
}

type DmSegParam struct {
	Pid           int64  `form:"pid"`
	Oid           int64  `form:"oid"`
	Type          int32  `form:"type"`
	SegmentIndex  int64  `form:"segment_index"`
	SDKIdentifier string `form:"sdk_identifier"`
	Platform      string `form:"platform"`
}

type StreamInfo struct {
	Quality        uint32 `json:"quality"`
	Format         string `json:"format"`
	Intact         bool   `json:"intact"`
	NoRexcode      bool   `json:"no_rexcode"`
	Attribute      int64  `json:"attribute"`
	NewDescription string `json:"new_description"`
	DisplayDesc    string `json:"display_desc"`
}

type DashVideo struct {
	BaseURL   string   `json:"base_url"`
	BackupURL []string `json:"backup_url"`
	Bandwidth uint32   `json:"bandwidth"`
	Codecid   uint32   `json:"codecid"`
	Md5       string   `json:"md5"`
	Size      uint64   `json:"size"`
	AudioID   uint32   `json:"audio_id"`
	NoRexcode bool     `json:"no_rexcode"`
}

type Segment struct {
	Order     uint32   `json:"order"`
	Length    uint64   `json:"length"`
	Size      uint64   `json:"size"`
	URL       string   `json:"url"`
	BackupURL []string `json:"backup_url"`
	Md5       string   `json:"md5"`
}

type SegmentVideo struct {
	Segment []*Segment `json:"segment"`
}

type Stream struct {
	StreamInfo   *StreamInfo   `json:"stream_info"`
	DashVideo    *DashVideo    `json:"dash_video"`
	SegmentVideo *SegmentVideo `json:"segment_video"`
}

type DashItem struct {
	ID        uint32   `json:"id"`
	BaseURL   string   `json:"base_url"`
	BackupURL []string `json:"backup_url"`
	Bandwidth uint32   `json:"bandwidth"`
	Codecid   uint32   `json:"codecid"`
	Md5       string   `json:"md5"`
	Size      uint64   `json:"size"`
}

type PlayURLMsg struct {
	Aid          int64       `json:"aid"`
	Cid          int64       `json:"cid"`
	Quality      uint32      `json:"quality"`
	Format       string      `json:"format"`
	Timelength   uint64      `json:"timelength"`
	VideoCodecid uint32      `json:"video_codecid"`
	StreamList   []*Stream   `json:"stream_list"`
	DashAudio    []*DashItem `json:"dash_audio"`
}
