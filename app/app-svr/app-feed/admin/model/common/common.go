package common

import (
	"fmt"
	"strconv"

	"go-gateway/pkg/idsafe/bvid"
)

const (
	//NotDeleted db not deleted
	NotDeleted = 0
	//Deleted db deleted
	Deleted = 1
	//Verify 待审核
	Verify = 1
	//Pass 已通过
	Pass = 2
	//Rejecte 已拒绝
	Rejecte = 3
	//Valid 已生效
	Valid = 4
	//InValid 已失效
	InValid = 5
	//StatusOnline status online
	StatusOnline = 1
	//StatusDownline status downline
	StatusDownline = 0
	//OptionOnline option online
	OptionOnline = "online"
	//OptionHidden option downline
	OptionHidden = "hidden"
	//OptionPass option pass
	OptionPass = "pass"
	//OptionReject option reject
	OptionReject      = "reject"
	OptionBatchPass   = "batch_pass"
	OptionBatchReject = "batch_reject"
	OptionBatchHidden = "batch_hidden"
	//Notify send message to up
	Notify = 1
	//NotifyDone notify done
	NotifyDone = 1
	//NotifyNotDone not notify
	NotifyNotDone = 0
	//DySearFilLevel dynamic search sensitive word level
	DySearFilLevel = 20
	//DySearFilArea dynamic search sensitive word area
	DySearFilArea = "bplus_dongtai"
	StateOK       = 0
	StateBlock    = 1
)

// notify
const (
	//NotifyBusnessTianma 天马业务
	NotifyBusnessTianma = 1
	//NotifyBusnessPopular 热门业务
	NotifyBusnessPopular = 2
	//NotifyTypArchive 视频
	NotifyTypArchive = 1
	//NotifyTypLive 直播
	NotifyTypLive = 2
	//NotifyTypArticle 专栏
	NotifyTypArticle   = 3
	NotifyTitleArchive = "稿件"
	NotifyTitleLive    = "直播"
	NotifyTitleArticle = "专栏"
)

// Page pager
type Page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// CardPreview card preview
type CardPreview struct {
	Title string      `json:"title"`
	Raw   interface{} `json:"raw"`
}

func GetAvIDStr(input string) (aid string, err error) {
	var aidInt int64
	if aidInt, err = GetAvID(input); err != nil {
		return "", err
	}
	aid = strconv.FormatInt(aidInt, 10)
	return
}

func GetAvID(input string) (aid int64, err error) {
	if aid, err = strconv.ParseInt(input, 10, 64); err != nil {
		err = nil
		if aid, err = bvid.BvToAv(input); err != nil {
			return 0, fmt.Errorf("视频ID非法！")
		}
	}
	return
}

func GetBvID(input int64) (bid string, err error) {
	if bid, err = bvid.AvToBv(input); err != nil {
		return "", fmt.Errorf("视频ID非法！")
	}
	return
}
