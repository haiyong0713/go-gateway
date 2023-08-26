package v2

import (
	playurlgrpc "git.bilibili.co/bapis/bapis-go/video/vod/playurlugc"
)

func (out *ResponseMsg) FromPlayurlV2(in *playurlgrpc.ResponseMsg, isView, isSp, isFreeSp bool) {
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
		durlItem.fromDurl(v)
		out.Durl = append(out.Durl, durlItem)
	}
	if in.Dash != nil {
		pDash := new(ResponseDash)
		pDash.fromDash(in.Dash, isView, isSp, isFreeSp)
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
			Codecs:         v.Codecs,
		}
		out.SupportFormats = append(out.SupportFormats, tmpFormat)
	}
}

func (out *ResponseUrl) fromDurl(in *playurlgrpc.ResponseUrl) {
	out.Order = in.Order
	out.Length = in.Length
	out.Size_ = in.Size_
	out.Ahead = in.Ahead
	out.Vhead = in.Vhead
	out.Url = in.Url
	out.BackupUrl = in.BackupUrl
	out.Md5 = in.Md5
}

func (out *ResponseDash) fromDash(in *playurlgrpc.ResponseDash, isView, isSp, isFreeSp bool) {
	out.Duration = in.Duration
	out.MinBufferTime = in.MinBufferTime
	for _, v := range in.Video {
		if v == nil {
			continue
		}
		videoItem := new(DashItem)
		videoItem.fromDashItem(v, isView)
		out.Video = append(out.Video, videoItem)
	}
	for _, v := range in.Audio {
		if v == nil {
			continue
		}
		audioItem := new(DashItem)
		audioItem.fromDashItem(v, false)
		out.Audio = append(out.Audio, audioItem)
	}
	if in.Dolby != nil && in.Dolby.Type != playurlgrpc.DolbyItem_NONE {
		out.Dolby = &DolbyItem{
			Type: DolbyItem_Type(in.Dolby.Type),
		}
		if isSp && len(in.Dolby.Audio) > 0 { //非大会员且非up本人仅返回type用以区分杜比音效和杜比全景声
			for _, v := range in.Dolby.Audio {
				if v == nil {
					continue
				}
				dolbyAudioItem := new(DashItem)
				dolbyAudioItem.fromDashItem(v, false)
				out.Dolby.Audio = append(out.Dolby.Audio, dolbyAudioItem)
			}
		}
	}
	//无损音频
	if in.GetLosslessAudio().GetAudio() != nil {
		out.LossLessItem = &LossLessItem{IsLosslessAudio: in.GetLosslessAudio().GetIsLosslessAudio()}
		if isFreeSp || isSp { //非大会员且非up本人仅返回type用以区分杜比音效和杜比全景声或者限时免费
			lessAudioItem := new(DashItem)
			lessAudioItem.fromDashItem(in.GetLosslessAudio().GetAudio(), false)
			out.LossLessItem.Audio = lessAudioItem
		}
	}
}

func (out *DashItem) fromDashItem(in *playurlgrpc.DashItem, isView bool) {
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
	// playView需要每路的清晰度是否全二压
	if isView {
		out.NoRexcode = in.NoRexcode
	}
}
