package v2

import (
	hlsgrpc "git.bilibili.co/bapis/bapis-go/video/vod/playurlhls"
)

func (out *HlsResponseMsg) FromPlayurlHls(in *hlsgrpc.SchedulerResponseMsg) {
	out.Code = in.Code
	out.Message = in.Message
	out.Type = ResponseType(in.Type)
	out.Quality = in.Quality
	out.Format = in.Format
	out.Timelength = in.Timelength
	out.VideoCodecid = in.VideoCodecid
	out.VideoProject = in.VideoProject
	for _, v := range in.Durl {
		if v == nil {
			continue
		}
		durlItem := new(ResponseUrl)
		durlItem.fromDurlHls(v)
		out.Durl = append(out.Durl, durlItem)
	}
	for _, v := range in.SupportFormats {
		if v == nil {
			continue
		}
		tmpFormat := &FormatDescription{
			Quality:        v.Quality,
			Format:         v.Format,
			NewDescription: v.Description,
			DisplayDesc:    v.DisplayDesc,
			Superscript:    v.Superscript,
		}
		out.SupportFormats = append(out.SupportFormats, tmpFormat)
	}
}

func (out *ResponseUrl) fromDurlHls(in *hlsgrpc.ResponseUrl) {
	out.Order = in.Order
	out.Length = in.Length
	out.Size_ = in.Size_
	out.Ahead = in.Ahead
	out.Vhead = in.Vhead
	out.Url = in.Url
	out.BackupUrl = in.BackupUrl
}

func fromM3U8Video(in *hlsgrpc.M3U8Video) *M3U8Video {
	return &M3U8Video{
		Qn:               in.Qn,
		Bandwidth:        in.Bandwidth,
		Resolution:       in.Resolution,
		Codecs:           in.Codecs,
		Name:             in.Name,
		FrameRate:        in.FrameRate,
		AverageBandwidth: in.AverageBandwidth,
	}
}

func (out *MasterScheduler) FromPlayurlMaster(in *hlsgrpc.MasterResponseMsg, supportDolby bool) {
	out.Code = in.Code
	out.Message = in.Message
	if in.Video != nil {
		out.Video = fromM3U8Video(in.Video)
	}
	if in.Audio != nil {
		out.Audio = &M3U8Audio{
			Qn:      in.Audio.Qn,
			Codecs:  in.Audio.Codecs,
			GroupId: in.Audio.GroupId,
		}
	}
	for _, v := range in.Videos {
		if v == nil {
			continue
		}
		out.Videos = append(out.Videos, fromM3U8Video(v))
	}
	if !supportDolby {
		return
	}
	if in.DolbyAudio != nil {
		out.Audio = &M3U8Audio{
			Qn:      in.DolbyAudio.Qn,
			Codecs:  in.DolbyAudio.Codecs,
			GroupId: in.DolbyAudio.GroupId,
		}
	}
}

// FromPlayurlM3U8 .
func (out *M3U8ResponseMsg) FromPlayurlM3U8(in *hlsgrpc.M3U8ResponseMsg) {
	out.Code = in.Code
	out.Message = in.Message
	out.M3U8Data = in.M3U8Data
}
