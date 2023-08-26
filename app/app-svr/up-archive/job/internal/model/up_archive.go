package model

import (
	"math/rand"
	"time"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/up-archive/service/api"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"

	"github.com/pkg/errors"
)

const _upArcScoreMaxRand = 512999

type UpArc struct {
	Aid         int64
	Mid         int64
	PubTime     xtime.Time
	CopyRight   int8
	Attribute   int32
	AttributeV2 int32
	RedirectURL string
	Score       int64
	UpFrom      int64
}

type UpArcSub struct {
	Mid   int64 `json:"mid"`
	Ctime int64 `json:"ctime"`
}

func (a *UpArc) IsAllowed(fItem []*cfcgrpc.ForbiddenItem) bool {
	var noSpace bool
	for _, item := range fItem {
		switch item.Key {
		case "no_space":
			noSpace = item.Value == 1
		default:
		}
	}
	return AttrVal(a.Attribute, api.AttrBitIsPUGVPay) == api.AttrNo && !noSpace
}

func (a *UpArc) IsOldAllowed() bool {
	return AttrVal(a.Attribute, api.AttrBitIsPUGVPay) == api.AttrNo && AttrVal(a.AttributeV2, api.AttrBitV2NoPublic) == api.AttrNo && AttrVal(a.AttributeV2, api.AttrBitV2NoSpace) == api.AttrNo
}

func (a *UpArc) IsUpNoSpace(fItem []*cfcgrpc.ForbiddenItem) bool {
	for _, item := range fItem {
		switch item.Key {
		case "up_no_space":
			return item.Value == 1
		default:
		}
	}
	return false
}

func (a *UpArc) IsLivePlayback(upFrom []int64) bool {
	for _, val := range upFrom {
		if a.UpFrom == val {
			return true
		}
	}
	return false
}

func (a *UpArc) IsStory() bool { //过滤掉付费稿件
	return AttrVal(a.Attribute, api.AttrBitIsPGC) == api.AttrNo && AttrVal(a.Attribute, api.AttrBitSteinsGate) == api.AttrNo &&
		a.RedirectURL == "" && AttrVal(a.AttributeV2, api.AttrBitV2Pay) == api.AttrNo //非付费稿件=稿件属性位为非付费
}

func (a *UpArc) RandScoreNumber() {
	rand.Seed(time.Now().UnixNano()) // 随机种子
	a.Score = int64(a.PubTime<<21) | int64(rand.Intn(_upArcScoreMaxRand))<<2 | int64(a.CopyRight)
}

func (a *UpArc) FromArchiveSub(in *ArchiveSub) error {
	pubTime, err := time.ParseInLocation("2006-01-02 15:04:05", in.PubTime, time.Local)
	if err != nil {
		return errors.Wrapf(err, "FromArchiveSub time.ParseInLocation pubtime:%s", in.PubTime)
	}
	a.Aid = in.Aid
	a.Mid = in.Mid
	a.PubTime = xtime.Time(pubTime.Unix())
	a.CopyRight = in.Copyright
	a.Attribute = in.Attribute
	a.AttributeV2 = in.AttributeV2
	a.RedirectURL = in.RedirectURL
	a.RandScoreNumber()
	return nil
}

func AttrVal(attr int32, bit uint) (v int32) {
	v = (attr >> bit) & int32(1)
	return
}
