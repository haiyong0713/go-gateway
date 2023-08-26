package model

import (
	arcmdl "go-gateway/app/app-svr/archive/service/model"
	v2 "go-gateway/app/app-svr/playurl/service/api/v2"
)

// FromPlayurlV2 .
func (p *PlayurlRes) FromPlayurlV2(reply *v2.PlayURLReply, isLogin bool) {
	p.From = "local"
	p.Result = "suee"
	p.Quality = reply.Playurl.Quality
	p.Format = reply.Playurl.Format
	p.Timelength = reply.Playurl.Timelength
	p.AcceptFormat = reply.Playurl.AcceptFormat
	p.AcceptDescription = reply.Playurl.AcceptDescription
	p.AcceptQuality = reply.Playurl.AcceptQuality
	p.VideoCodeCid = reply.Playurl.VideoCodecid
	p.SeekParam = "start"
	if reply.Playurl.Format == "mp4" {
		p.SeekType = "second"
	} else {
		p.SeekType = "offset"
	}
	for _, v := range reply.Playurl.Durl {
		if v == nil {
			continue
		}
		durlItem := new(durl)
		durlItem.fromDurlV2(v)
		p.Durl = append(p.Durl, durlItem)
	}
	if reply.Playurl.Dash != nil {
		pDash := new(dash)
		pDash.fromDashV2(reply.Playurl.Dash, isLogin)
		p.Dash = pDash
	}
	// 杜比HDR 与 HDR10 互斥，优先输出杜比HDR
	var hasDolby, hasHDR bool
	for _, v := range reply.Playurl.SupportFormats {
		if v == nil {
			continue
		}
		switch v.Quality {
		case arcmdl.QnHDR:
			hasHDR = true
		case arcmdl.QnDolbyHDR:
			hasDolby = true
		default:
		}
	}
	for _, v := range reply.Playurl.SupportFormats {
		if v == nil {
			continue
		}
		if v.Quality == arcmdl.QnHDR && hasDolby && hasHDR {
			continue
		}
		formatDescItem := new(formatDescription)
		formatDescItem.FromFormatDescription(v)
		p.SupportFormats = append(p.SupportFormats, formatDescItem)
	}
	if reply.Playurl.HighFormat != nil {
		formatDescItem := new(formatDescription)
		formatDescItem.FromFormatDescription(reply.Playurl.HighFormat)
		p.HighFormat = formatDescItem
	}
	if reply.Volume != nil {
		volumeInfo := new(VolumeInfo)
		volumeInfo.FromVolumeInfo(reply.Volume)
		p.Volume = volumeInfo
	}
}

// FromDurl from durl data.
func (d *durl) fromDurlV2(item *v2.ResponseUrl) {
	d.Order = item.Order
	d.Length = item.Length
	d.Size = item.Size_
	d.Ahead = item.Ahead
	d.Vhead = item.Vhead
	d.URL = item.Url
	d.BackupURL = item.BackupUrl
}

func (d *dash) fromDashV2(item *v2.ResponseDash, isLogin bool) {
	d.Duration = item.Duration
	d.MinBufferTime = item.MinBufferTime
	d.MinBufferTime2 = item.MinBufferTime
	for _, v := range item.Video {
		if v == nil {
			continue
		}
		videoItem := new(dashItem)
		videoItem.fromDashItemV2(v)
		d.Video = append(d.Video, videoItem)
	}
	for _, v := range item.Audio {
		if v == nil {
			continue
		}
		audioItem := new(dashItem)
		audioItem.fromDashItemV2(v)
		d.Audio = append(d.Audio, audioItem)
	}
	d.Dolby = item.Dolby
	if item.LossLessItem != nil {
		// 无损音频
		d.FLAC = &flac{
			Display: item.LossLessItem.IsLosslessAudio,
		}
		if item.LossLessItem.Audio == nil || !isLogin {
			// 1、未登陆不下发url 2、大会员由上游限制不下发
			return
		}
		d.FLAC.Audio = &dashItem{}
		d.FLAC.Audio.fromDashItemV2(item.LossLessItem.Audio)
	}
}

// FromDashItem from dash item.
func (d *dashItem) fromDashItemV2(item *v2.DashItem) {
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
	d.StartWithSAP = item.StartWithSap
	d.StartWithSAP2 = item.StartWithSap
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

func (d *formatDescription) FromFormatDescription(item *v2.FormatDescription) {
	d.Quality = item.Quality
	d.Format = item.Format
	d.NewDescription = item.NewDescription
	d.DisplayDesc = item.DisplayDesc
	d.Superscript = item.Superscript
	d.Codecs = item.Codecs
}

func (d *VolumeInfo) FromVolumeInfo(item *v2.VolumeInfo) {
	d.MeasuredI = item.MeasuredI
	d.MeasuredLra = item.MeasuredLra
	d.MeasuredTp = item.MeasuredTp
	d.MeasuredThreshold = item.MeasuredThreshold
	d.TargetOffset = item.TargetOffset
	d.TargetI = item.TargetI
	d.TargetTp = item.TargetTp
}
