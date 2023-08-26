package model

import (
	"encoding/json"
	"hash/crc64"

	pb "go-gateway/app/app-svr/kvo/interface/api"
)

// Action action message
type Action struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

// UserConf user configruation
type UserConf struct {
	ModuleKey int   `json:"module_key,omitempty"`
	Mid       int64 `json:"mid,omitempty"`
	CheckSum  int64 `json:"check_sum"`
	Timestamp int64 `json:"timestamp"`
}

// Document data store
type Document struct {
	CheckSum int64  `json:"check_sum"`
	Doc      string `json:"doc"`
}

type Buvid string

func (b Buvid) Crc63() int {
	if b == "" {
		return 0
	}
	return int(crc64.Checksum([]byte(b), crcTable) >> 1)
}

type CfgMessage struct {
	Mid      int64
	Body     pb.ConfigModify
	Action   string
	Platform string
	Buvid    Buvid
}

type MergeCfgMessage struct {
	Mid      int64
	Bodys    map[int]pb.ConfigModify
	Platform string
	Buvid    Buvid
}

func (m *MergeCfgMessage) Merge(moduleId int, dst pb.ConfigModify) {
	if body, ok := m.Bodys[moduleId]; ok {
		body.Merge(dst)
	} else {
		m.Bodys[moduleId] = dst
	}
	return
}
