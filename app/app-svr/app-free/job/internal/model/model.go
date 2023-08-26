package model

import (
	"fmt"
	"math/big"
	"net"

	xtime "go-common/library/time"
)

type RecordState int

const (
	StateRecording RecordState = 0
	StateSucess    RecordState = 1
	StateCancel    RecordState = 2
)

type ISP string

const (
	ISPUnicom ISP = "cu"
	ISPTelcom ISP = "ct"
	ISPMobile ISP = "cm"
)

type RecordBusiness string

// Kratos hello kratos.
type Kratos struct {
	Hello string
}

type TFRecord struct {
	ISP        string `json:"isp,omitempty"`
	RemoteIP   string `json:"remote_ip,omitempty"`
	RemoteHost string `json:"remote_host,omitempty"`
	Count      int    `json:"count,omitempty"`
	Size       int    `json:"size,omitempty"`
	FullURI    string `json:"full_uri,omitempty"`
	Info       string `json:"info,omitempty"`
}

// Comment: 免流局数据备案表
type FreeRecord struct {
	// Comment: 主键ID
	ID int `json:"id"`
	// Comment: IP段起始
	IPStart string `json:"ip_start"`
	// Comment: IP段起始int型
	// Default: 0
	IPStartInt int64 `json:"ip_start_int"`
	// Comment: IP段结束
	IPEnd string `json:"ip_end"`
	// Comment: IP段结束int型
	// Default: 0
	IPEndInt int64 `json:"ip_end_int"`
	// Comment: 局数据备案所属运营商
	ISP ISP `json:"isp"`
	// Comment: 局数据是否是BGP，0-否,1-是
	// Default: 0
	IsBGP bool `json:"is_bgp"`
	// Comment: 业务
	Business RecordBusiness `json:"business"`
	// Comment: IP备案状态，0-备案中,1-备案成功,2-已下线
	// Default: 0
	State RecordState `json:"state"`
	// Comment: 备案成功时间
	// Default: 0000-00-00 00:00:00
	SuccessTime xtime.Time `json:"success_time"`
	// Comment: 备案下线时间
	// Default: 0000-00-00 00:00:00
	CancelTime xtime.Time `json:"cancel_time"`
	// Comment: 创建时间
	// Default: CURRENT_TIMESTAMP
	Ctime xtime.Time `json:"ctime"`
	// Comment: 最后修改时间
	// Default: CURRENT_TIMESTAMP
	Mtime xtime.Time `json:"mtime"`
}

func InetNtoA(ip int64) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func InetAtoN(ip string) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return ret.Int64()
}

func IsIPv4(ip string) bool {
	ipv := net.ParseIP(ip)
	if ip := ipv.To4(); ip != nil {
		return true
	} else {
		return false
	}
}

func CheckIP(ipStart, ipEnd string) bool {
	if !IsIPv4(ipStart) || !IsIPv4(ipEnd) {
		return false
	}
	if InetAtoN(ipStart) > InetAtoN(ipEnd) {
		return false
	}
	return true
}

// nolint:gomnd
func IsPublicIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		switch {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}

type ExtendedFields struct {
	Cid                         string `json:"cid,omitempty"`
	Mode                        string `json:"mode,omitempty"`
	TimeOfEvent                 string `json:"time_of_event,omitempty"`
	TimeOfVideo                 string `json:"time_of_video,omitempty"`
	SeekDiff                    string `json:"seek_diff,omitempty"`
	Error                       string `json:"error,omitempty"`
	ItemPlay                    string `json:"item_play,omitempty"`
	Iformat                     string `json:"iformat,omitempty"`
	AssetItemTimeOfSession      string `json:"asset_item_time_of_session,omitempty"`
	AssetItemSession            string `json:"asset_item_session,omitempty"`
	PlaybackRate                string `json:"playback_rate,omitempty"`
	BufferingCount              string `json:"buffering_count,omitempty"`
	FirstAudioTime              string `json:"first_audio_time,omitempty"`
	FirstVideoTime              string `json:"first_video_time,omitempty"`
	Width                       string `json:"width,omitempty"`
	Height                      string `json:"height,omitempty"`
	Vcodec                      string `json:"vcodec,omitempty"`
	AudioBitrate                string `json:"audio_bitrate,omitempty"`
	VideoBitrate                string `json:"video_bitrate,omitempty"`
	Vdecoder                    string `json:"vdecoder,omitempty"`
	VdropRate                   string `json:"vdrop_rate,omitempty"`
	DecodeSwitchSoftFrame       string `json:"decode_switch_soft_frame,omitempty"`
	DashTargetQn                string `json:"dash_target_qn,omitempty"`
	DashCurQn                   string `json:"dash_cur_qn,omitempty"`
	DashAuto                    string `json:"dash_auto,omitempty"`
	AssetSession                string `json:"asset_session,omitempty"`
	AssetTimeOfSession          string `json:"asset_time_of_session,omitempty"`
	AudioURL                    string `json:"audio_url,omitempty"`
	VideoURL                    string `json:"video_url,omitempty"`
	StepWaitTime                string `json:"step_wait_time,omitempty"`
	VideoNetError               string `json:"video_net_error,omitempty"`
	AudioNetError               string `json:"audio_net_error,omitempty"`
	DNSType                     string `json:"dns_type,omitempty"`
	AudioTransportNread         string `json:"audio_transport_nread,omitempty"`
	VideoTransportNread         string `json:"video_transport_nread,omitempty"`
	AudioReadBytes              string `json:"audio_read_bytes,omitempty"`
	VideoReadBytes              string `json:"video_read_bytes,omitempty"`
	FirstRenderMode             string `json:"first_render_mode,omitempty"`
	VideoHTTPCode               string `json:"video_http_code,omitempty"`
	AudioHTTPCode               string `json:"audio_http_code,omitempty"`
	HTTPOffset                  string `json:"http_offset,omitempty"`
	VideoIP                     string `json:"video_ip,omitempty"`
	AudioIP                     string `json:"audio_ip,omitempty"`
	VideoNetSpeed               string `json:"video_net_speed,omitempty"`
	VideoHost                   string `json:"video_host,omitempty"`
	AudioHost                   string `json:"audio_host,omitempty"`
	IsAudio                     string `json:"is_audio,omitempty"`
	PlayerGetFirstPkgTime       string `json:"player_get_first_pkg_time,omitempty"`
	RevcVideoFirstPkgTimestamp  string `json:"revc_video_first_pkg_timestamp,omitempty"`
	FirstVideoWillHTTPTimestamp string `json:"first_video_will_http_timestamp,omitempty"`
	IsComplete                  string `json:"is_complete,omitempty"`
	BufferWaterMaker            string `json:"buffer_water_maker,omitempty"`
	MainCPURate                 string `json:"main_cpu_rate,omitempty"`
	MainMem                     string `json:"main_mem,omitempty"`
	IjkCPURate                  string `json:"ijk_cpu_rate,omitempty"`
	IjkMem                      string `json:"ijk_mem,omitempty"`
	AudioDNSTime                string `json:"audio_dns_time,omitempty"`
	VideoDNSTime                string `json:"video_dns_time,omitempty"`
	AudioTCPTime                string `json:"audio_tcp_time,omitempty"`
	VideoTCPTime                string `json:"video_tcp_time,omitempty"`
	AudioPort                   string `json:"audio_port,omitempty"`
	VideoPort                   string `json:"video_port,omitempty"`
	AudioPtsDiff                string `json:"audio_pts_diff,omitempty"`
	LastAudioNetError           string `json:"last_audio_net_error,omitempty"`
	LastAudioNetErrorURL        string `json:"last_audio_net_error_url,omitempty"`
	LastVideoNetError           string `json:"last_video_net_error,omitempty"`
	LastVideoNetErrorURL        string `json:"last_video_net_error_url,omitempty"`
	PlayerStatus                string `json:"player_status,omitempty"`
	HTTPBuildReason             string `json:"http_build_reason,omitempty"`
	AssetUpdateCount            string `json:"asset_update_count,omitempty"`
	URLInfo                     string `json:"url_info,omitempty"`
	SeekFirstPkgTime            string `json:"seek_first_pkg_time,omitempty"`
	SeekBufferingAccTime        string `json:"seek_buffering_acc_time,omitempty"`
	ForceReport                 string `json:"force_report,omitempty"`
	AudioDuration               string `json:"audio_duration,omitempty"`
	VideoDuration               string `json:"video_duration,omitempty"`
	AvgBufferCacheVideoPackets  string `json:"avg_buffer_cache_video_packets,omitempty"`
	AvgBufferCacheVideoTime     string `json:"avg_buffer_cache_video_time,omitempty"`
	LiveDelayTime               string `json:"live_delay_time,omitempty"`
	NetFamily                   string `json:"net_family,omitempty"`
	Ipv6Info                    string `json:"ipv_6_info,omitempty"`
}
