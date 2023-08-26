package abtest

import (
	"encoding/json"
	"fmt"
	"strings"

	"go-gateway/app/app-svr/distribution/distribution/admin/internal/model"

	"github.com/pkg/errors"
)

// abtest信息
type Infos struct {
	//实验id
	ID string `json:"id"`
	//实验名称
	Name string `json:"name"`
	//实验变量
	FlagValue string `json:"flag_value"`
	//创建人
	Creator string `json:"creator"`
	//实验状态
	Status string `json:"status"`
}

// taishan中的key
// 实验id+分组id来唯一确认一份配置
type TaishanKeyInfos struct {
	//实验id
	ID string
	//分组id
	GroupIDs []string
}

func (k TaishanKeyInfos) BuildKeys() []string {
	var keys []string
	for _, v := range k.GroupIDs {
		keys = append(keys, fmt.Sprintf("abtest_%s_%s", k.ID, v))
	}
	return keys
}

func TaishanKeyStringToDetail(key string) (*Detail, error) {
	keySplits := strings.Split(key, "_")
	if len(keySplits) != model.ValidateKeyLen {
		return nil, errors.New("Wrong key format")
	}
	return &Detail{
		ID:      keySplits[1],
		GroupID: keySplits[2],
	}, nil
}

type Detail struct {
	ID        string          `json:"id"`
	GroupID   string          `json:"group_id"`
	GroupName string          `json:"group_name"`
	FlagValue string          `json:"flag_value"`
	Config    json.RawMessage `json:"config"`
}

type DetailReq struct {
	ExpID     string `json:"exp_id" form:"exp_id" validate:"required"`
	FlagValue string `json:"flag_value" form:"flag_value" validate:"required"`
}

type DetailReply struct {
	Details        []*Detail               `json:"details"`
	FieldBasicInfo []*model.FieldBasicInfo `json:"field_basic_info"`
}
