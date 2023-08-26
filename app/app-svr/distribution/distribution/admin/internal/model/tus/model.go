package tus

import (
	"encoding/json"
	"fmt"
	"strings"

	"go-gateway/app/app-svr/distribution/distribution/admin/internal/model"

	"github.com/pkg/errors"
)

// tus元信息
type Info struct {
	//人群包id
	TusValue string `json:"tus_value"`
	//人群包名称
	Name string `json:"name"`
	//创建人
	Creator string `json:"creator"`
	//人群包状态
	Status int64 `json:"status"`
	//有效期
	ValidDay int64 `json:"valid_day"`
	//人群量级
	CrowdCount int64 `json:"crowed_count"`
	//人群包创建时间
	Ctime int64 `json:"ctime"`
}

// TaishanKeyInfos 泰山中的key
// 实验id+分组id来唯一确认一份配置
type TaishanKeyInfos struct {
	//人群包id
	TusValue string
	//是否命中
	Result string
}

func (k TaishanKeyInfos) BuildKeyByResult() string {
	return fmt.Sprintf("tus_%s_%s", k.TusValue, k.Result)
}

func (k TaishanKeyInfos) BuildKeys() []string {
	keyForMiss := fmt.Sprintf("tus_%s_%s", k.TusValue, "0")
	keyForHit := fmt.Sprintf("tus_%s_%s", k.TusValue, "1")
	return []string{keyForHit, keyForMiss}
}

func TaishanKeyStringToDetail(key string) (*Detail, error) {
	keySplits := strings.Split(key, "_")
	if len(keySplits) != model.ValidateKeyLen {
		return nil, errors.New("Wrong key format")
	}
	return &Detail{
		TusValue: keySplits[1],
		Result:   keySplits[2],
	}, nil
}

type Detail struct {
	TusValue string          `json:"tus_value"`
	Result   string          `json:"result"`
	Config   json.RawMessage `json:"config"`
}

type DetailReply struct {
	Details        []*Detail               `json:"details"`
	FieldBasicInfo []*model.FieldBasicInfo `json:"field_basic_info"`
}
