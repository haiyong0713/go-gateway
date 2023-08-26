package v2

import (
	playurlgrpc "git.bilibili.co/bapis/bapis-go/video/vod/playurlhtml5"
)

func (out *ResponseMsg) FromPlayurlH5(in *playurlgrpc.ResponseMsg) {
	out.Code = in.Code
	out.Message = in.Message
	out.Type = int32(in.Type)
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
		durlItem := new(ResponseUrl)
		durlItem.fromDurlH5(v)
		out.Durl = append(out.Durl, durlItem)
	}
	if in.Dash != nil {
		pDash := new(ResponseDash)
		pDash.fromDashH5(in.Dash)
		out.Dash = pDash
	}
	out.NoRexcode = in.NoRexcode
	for _, v := range in.SupportFormats {
		if v == nil {
			continue
		}
		tmpFormat := &FormatDescription{
			Quality:        v.Quality,
			Format:         v.Format,
			Description:    v.Description,
			NewDescription: v.NewDescription,
			DisplayDesc:    v.DisplayDesc,
			Superscript:    v.Superscript,
		}
		out.SupportFormats = append(out.SupportFormats, tmpFormat)
	}
	if in.HighFormat == nil {
		return
	}
	out.HighFormat = &FormatDescription{
		Quality:        in.HighFormat.Quality,
		Format:         in.HighFormat.Format,
		Description:    in.HighFormat.Description,
		NewDescription: in.HighFormat.NewDescription,
		DisplayDesc:    in.HighFormat.DisplayDesc,
		Superscript:    in.HighFormat.Superscript,
	}
}

func (out *ResponseUrl) fromDurlH5(in *playurlgrpc.ResponseUrl) {
	out.Order = in.Order
	out.Length = in.Length
	out.Size_ = in.Size_
	out.Ahead = in.Ahead
	out.Vhead = in.Vhead
	out.Url = in.Url
	out.BackupUrl = in.BackupUrl
	out.Md5 = in.Md5
}

func (out *ResponseDash) fromDashH5(in *playurlgrpc.ResponseDash) {
	out.Duration = in.Duration
	out.MinBufferTime = in.MinBufferTime
	for _, v := range in.Video {
		if v == nil {
			continue
		}
		videoItem := new(DashItem)
		videoItem.fromDashItemH5(v)
		out.Video = append(out.Video, videoItem)
	}
	for _, v := range in.Audio {
		if v == nil {
			continue
		}
		audioItem := new(DashItem)
		audioItem.fromDashItemH5(v)
		out.Audio = append(out.Audio, audioItem)
	}
}

func (out *DashItem) fromDashItemH5(in *playurlgrpc.DashItem) {
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
	out.StartWithSap = in.StartWithSap
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
