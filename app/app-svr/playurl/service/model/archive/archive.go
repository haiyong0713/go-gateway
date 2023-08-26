package archive

import (
	arcApi "go-gateway/app/app-svr/archive/service/api"
)

// CloudInfo.
type CloudInfo struct {
	Ctime     string `json:"ctime"`
	Buvid     string `json:"buvid"`
	Platform  string `json:"platform"`
	FMode     int32  `json:"f_mode"`
	Ver       string `json:"ver"`
	Function  string `json:"function"`
	Brand     string `json:"brand"`
	Model     string `json:"model"`
	EditSouce string `json:"edit_souce"`
	FpLocal   string `json:"fp_local"`
}

// IsNormal check archive is normal
func (info *Info) IsNormal() bool {
	return info.State >= arcApi.StateOpen
}

// IsPGC is
func (info *Info) IsPGC() bool {
	return info.AttrVal(arcApi.AttrBitIsPGC) == arcApi.AttrYes
}

// IsSteinsGate is
func (info *Info) IsSteinsGate() bool {
	return info.AttrVal(arcApi.AttrBitSteinsGate) == arcApi.AttrYes
}

// Is360 .
func (info *Info) Is360() bool {
	return info.AttrValV2(arcApi.AttrBitV2Is360) == arcApi.AttrYes
}

// AttrVal get attr val by bit.
func (info *Info) AttrVal(bit uint) int32 {
	return (info.Attribute >> bit) & int32(1)
}

// IsNoBackground is
func (info *Info) IsNoBackground() bool {
	return info.AttrValV2(arcApi.AttrBitV2NoBackground) == arcApi.AttrYes
}

// AttrValV2 get attr v2 val by bit.
func (info *Info) AttrValV2(bit uint) int32 {
	return int32((info.AttributeV2 >> bit) & int64(1))
}

// HasCid check cid is in info.Cids
func (info *Info) HasCid(cid int64) (ok bool) {
	for _, id := range info.Cids {
		if cid == id {
			ok = true
			break
		}
	}
	return
}

// SteinsCanPreview check Steins Can Preview
func (info *Info) SteinsCanPreview() bool {
	return info.State >= arcApi.StateOpen || info.State == arcApi.StateForbidSteins
}

type Info struct {
	Aid         int64
	Cids        []int64
	State       int32
	Mid         int64
	Attribute   int32
	AttributeV2 int64
	SeasonID    int64
	TypeID      int32
	Copyright   int32
	Duration    int64
	Premiere    *arcApi.Premiere
	Pay         *arcApi.PayInfo
}

type VipFree struct {
	//是否限免 0-否 1-是
	LimitFree int32
	//副标题
	Subtitle string
}
