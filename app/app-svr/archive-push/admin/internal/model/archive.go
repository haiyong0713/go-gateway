package model

import (
	"fmt"
	archiveGRPC "git.bilibili.co/bapis/bapis-go/archive/service"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/archive-push/ecode"
)

// Dimension 视频分辨率
type Dimension struct {
	// Width 宽 如 1920
	Width int32 `json:"width"`
	// Height 高 如 1080
	Height int32 `json:"height"`
	// Rotate 是否竖屏 0=否 1=是
	Rotate int `json:"rotate"`
}

// Author 稿件作者信息
type Author struct {
	// Up主mid
	Mid int64 `json:"mid"`
	// Up主名称
	Name string `json:"name"`
	// Up主头像地址 绝对地址
	Face string `json:"face"`
}

// ArchivePushDetailByBVID 根据bvid查询稿件详情
type ArchivePushDetailByBVID struct {
	BVID          string     `json:"bvid"`
	VendorID      int64      `json:"vendorId"`
	ArchiveStatus string     `json:"archiveStatus"`
	PushStatus    string     `json:"pushStatus"`
	PushType      string     `json:"pushType"`
	BatchIDs      []int64    `json:"batchIds"`
	CUser         string     `json:"cuser"`
	CTime         xtime.Time `json:"ctime"`
}

type ArchivePushDetailByBVIDSlice []*ArchivePushDetailByBVID

func (s ArchivePushDetailByBVIDSlice) Len() int {
	return len(s)
}
func (s ArchivePushDetailByBVIDSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ArchivePushDetailByBVIDSlice) Less(i, j int) bool {
	strA := fmt.Sprintf("%010d%015d%s", s[i].VendorID, s[i].CTime, s[i].BVID)
	strB := fmt.Sprintf("%010d%015d%s", s[j].VendorID, s[j].CTime, s[j].BVID)

	return strA > strB
}

type ArchiveMetadataAll struct {
	*archiveGRPC.Arc
	Tags   map[int64]string `json:"tags"`
	DocID  string           `json:"docid"`
	OpenID string           `json:"openId"`
}

// SyncArchiveStatusReq 稿件审核状态变更消息Req
type SyncArchiveStatusReq struct {
	VendorID   int64  `json:"vendorId" form:"vendorId"`
	BVID       string `json:"bvid" form:"bvid"`
	OVID       string `json:"ovid" form:"ovid"`
	Status     string `json:"status" form:"status"`
	StatusTime string `json:"statusTime" form:"statusTime"`
	StatusMsg  string `json:"statusMsg" form:"statusMsg"`
}

// GetArchiveWhiteListKeyForVendor 根据vendor获取对应稿件白名单key
func GetArchiveWhiteListKeyForVendor(vendorID int64) (string, error) {
	switch vendorID {
	case 0, DefaultVendors[0].ID, DefaultVendors[1].ID:
		return RedisWhiteListKey, nil
	case DefaultVendors[2].ID:
		return fmt.Sprintf("%d_"+RedisWhiteListKey, vendorID), nil
	default:
		return "", ecode.VendorNotFound
	}
}
