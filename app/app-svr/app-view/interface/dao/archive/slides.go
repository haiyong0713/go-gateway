package archive

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/net/metadata"
)

type SlidesRequest struct {
	FromAV      int64
	SessionID   string
	DisplayID   int64
	FromTrackID string
	Mid         int64
	Buvid       string
	Timeout     int64
	Build       int64
	MobiApp     string
	Plat        int8
	RequestCnt  int64
	ZoneID      int64
	IP          string
	Network     string
}

type SlidesReply struct {
	Code      int64         `json:"code"`
	Data      []*SlidesItem `json:"data"`
	PVFeature string        `json:"pv_feature"`

	status struct {
		isBackupReply bool
		originCode    *int64
	}
}

type SlidesItem struct {
	TrackID   string `json:"trackid"`
	ID        int64  `json:"id"`
	Goto      string `json:"goto"`
	Source    string `json:"source"`
	AVFeature string `json:"av_feature"`
}

func (r *SlidesReply) IsBackupReply() bool {
	return r.status.isBackupReply
}

func (r *SlidesReply) MarkAsBackupReply() {
	r.status.isBackupReply = true
}

func (r *SlidesReply) ReturnCode() int64 {
	if r.status.originCode != nil {
		return *r.status.originCode
	}
	return r.Code
}

func (r *SlidesReply) StoreOriginCode(in int64) {
	r.status.originCode = &in
}

func (s *SlidesRequest) AsURLValue() url.Values {
	params := url.Values{}
	params.Set("cmd", "slides")
	params.Set("from_av", strconv.FormatInt(s.FromAV, 10))
	params.Set("session_id", s.SessionID)
	params.Set("display_id", strconv.FormatInt(s.DisplayID, 10))
	params.Set("from_trackid", s.FromTrackID)
	params.Set("mid", strconv.FormatInt(s.Mid, 10))
	params.Set("buvid", s.Buvid)
	params.Set("timeout", "200")
	params.Set("mobi_app", s.MobiApp)
	params.Set("build", strconv.FormatInt(s.Build, 10))
	params.Set("plat", strconv.Itoa(int(s.Plat)))
	params.Set("request_cnt", strconv.FormatInt(s.RequestCnt, 10))
	params.Set("zone_id", strconv.FormatInt(s.ZoneID, 10))
	params.Set("ip", s.IP)
	params.Set("network", s.Network)
	return params
}

func (d *Dao) SlidesRecommend(ctx context.Context, req *SlidesRequest) (*SlidesReply, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := req.AsURLValue()
	reply := &SlidesReply{}
	if err := d.client.Get(ctx, d.relateRecURL, ip, params, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}
