package api

import (
	v2 "go-gateway/app/app-svr/playurl/service/api/v2"
)

const (
	ProjectPlatform = "tvproj"
	FnvalNeedDolby  = uint32(8)
	// qn 1080
	Qn1080 = uint32(80)
	// qn 720
	QnFlv720 = uint32(64)
)

func (out *PlayURLInfo) FromPlayurlV2(in *v2.ResponseMsg) {
	out.From = "local"
	out.Result = "suee"
	out.SeekParam = "start"
	if in.Format == "mp4" {
		out.SeekType = "second"
	} else {
		out.SeekType = "offset"
	}
	out.Quality = in.Quality
	out.Format = in.Format
	out.AcceptFormat = in.AcceptFormat
	out.AcceptDescription = in.AcceptDescription
	out.AcceptQuality = in.AcceptQuality
	out.Timelength = in.Timelength
	out.VideoCodecid = in.VideoCodecid
	out.Fnver = in.Fnver
	out.Fnval = in.Fnval
	out.VideoProject = in.VideoProject
	for _, v := range in.Durl {
		if v == nil {
			continue
		}
		durlItem := new(Durl)
		durlItem.fromDurl(v)
		out.Durl = append(out.Durl, durlItem)
	}
	if in.Dash != nil {
		pDash := new(Dash)
		pDash.fromDash(in.Dash)
		out.Dash = pDash
	}
	out.NoRexcode = in.NoRexcode
}

func (out *Durl) fromDurl(in *v2.ResponseUrl) {
	out.Order = in.Order
	out.Length = in.Length
	out.Size_ = in.Size_
	out.Ahead = in.Ahead
	out.Vhead = in.Vhead
	out.Url = in.Url
	out.BackupUrl = in.BackupUrl
	out.Md5 = in.Md5
}

func (out *Dash) fromDash(in *v2.ResponseDash) {
	out.Duration = in.Duration
	out.MinBufferTime = in.MinBufferTime
	for _, v := range in.Video {
		if v == nil {
			continue
		}
		videoItem := new(DashItem)
		videoItem.fromDashItem(v)
		out.Video = append(out.Video, videoItem)
	}
	for _, v := range in.Audio {
		if v == nil {
			continue
		}
		audioItem := new(DashItem)
		audioItem.fromDashItem(v)
		out.Audio = append(out.Audio, audioItem)
	}
}

func (out *DashItem) fromDashItem(in *v2.DashItem) {
	out.Id = in.Id
	out.BaseUrl = in.BaseUrl
	out.BackupUrl = in.BackupUrl
	out.Bandwidth = in.Bandwidth
	out.MimeType = in.MimeType
	out.Codecs = in.Codecs
	out.Width = in.Width
	out.Height = in.Height
	out.FrameRate = in.FrameRate
	out.Sar = in.Sar
	out.StartWithSAP = in.StartWithSap
	if in.SegmentBase != nil {
		out.SegmentBase = &DashSegmentBase{
			Initialization: in.SegmentBase.Initialization,
			IndexRange:     in.SegmentBase.IndexRange,
		}
	}
	out.Codecid = in.Codecid
	out.Md5 = in.Md5
	out.Size_ = in.Size_
}
