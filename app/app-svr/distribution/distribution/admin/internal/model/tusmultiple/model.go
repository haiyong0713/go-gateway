package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"go-gateway/app/app-svr/distribution/distribution/admin/internal/model"
)

type FieldInfo struct {
	Name       string `json:"field_name"`
	Descriptor string `json:"descriptor"`
}

type Detail struct {
	TusValue string          `json:"tus_value"`
	TusName  string          `json:"tus_name"`
	Config   json.RawMessage `json:"config"`
}

type DetailReply struct {
	Details        []*Detail               `json:"details"`
	FieldBasicInfo []*model.FieldBasicInfo `json:"field_basic_info"`
}

type TaishanKeyInfo struct {
	ConfigVersion string
	Filed         string
	TusValues     []string
}

func (t *TaishanKeyInfo) BuildKeys() []string {
	var keys []string
	for _, v := range t.TusValues {
		keys = append(keys, KeyFormat(t.Filed, t.ConfigVersion, v))
	}
	return keys
}

func TaishanKeyStringToDetail(key string) (*Detail, error) {
	keySplits := strings.Split(key, "_")
	if len(keySplits) < model.ValidateKeyLen {
		return nil, errors.New("Wrong key format")
	}
	return &Detail{
		TusValue: keySplits[2],
	}, nil
}

func KeyFormat(field, configVersion, tusValue string) string {
	if configVersion == "v1.0" { //1.0版本还是使用老key，不用进行配置迁移
		return fmt.Sprintf("tusmultiple_%s_%s", field, tusValue)
	}
	return fmt.Sprintf("tusmultiple_%s_%s_%s", field, tusValue, configVersion)
}
