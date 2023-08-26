package player

import (
	vod "go-gateway/app/app-svr/playurl/service/api/v2"
)

// Param is
type Param struct {
	AID       int64  `form:"aid"`
	CID       int64  `form:"cid"`
	Qn        int64  `form:"qn"`
	Npcybs    int32  `form:"npcybs"`
	Otype     string `form:"otype"`
	MobiApp   string `form:"mobi_app"`
	Fnver     int32  `form:"fnver"`
	Fnval     int32  `form:"fnval"`
	Session   string `form:"session"`
	Build     int32  `form:"build"`
	Device    string `form:"device"`
	ForceHost int32  `form:"force_host"`
	Fourk     int32  `form:"fourk"`
	Buvid     string `form:"buvid"`
	Platform  string `form:"platform"`
	Dl        int32  `form:"dl"`
}

type PlayurlV2Reply struct {
	Quality           uint32    `json:"quality"`
	Format            string    `json:"format"`
	Timelength        uint64    `json:"timelength"`
	NoRexcode         int32     `json:"no_rexcode,omitempty"`
	AcceptFormat      string    `json:"accept_format"`
	AcceptDescription []string  `json:"accept_description,omitempty"`
	AcceptQuality     []uint32  `json:"accept_quality,omitempty"`
	VideoCodecid      uint32    `json:"video_codecid,omitempty"`
	Fnver             uint32    `json:"fnver"`
	Fnval             uint32    `json:"fnval"`
	VideoProject      bool      `json:"video_project"`
	Durl              []*DurlV2 `json:"durl,omitempty"`
	Dash              *DashV2   `json:"dash,omitempty"`
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
			tmpDurl = append(tmpDurl, &DurlV2{
				Order:     v.Order,
				Length:    v.Length,
				Size:      v.Size_,
				AHead:     v.Ahead,
				VHead:     v.Vhead,
				URL:       v.Url,
				BackupURL: v.BackupUrl,
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
				tmpVideo = append(tmpVideo, &DashItemV2{
					ID:        v.Id,
					BaseURL:   v.BaseUrl,
					BackupURL: v.BackupUrl,
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
				tmpAudio = append(tmpAudio, &DashItemV2{
					ID:        v.Id,
					BaseURL:   v.BaseUrl,
					BackupURL: v.BackupUrl,
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
