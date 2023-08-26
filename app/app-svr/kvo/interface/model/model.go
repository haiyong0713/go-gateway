package model

import (
	"encoding/json"

	"go-common/library/ecode"
)

const (
	BILogStreamID = "003353"
)

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

// BILogStream .
type BILogStream struct {
	Mid      int64
	CTime    int64
	Business string
	Buvid    string
	Body     string
	Platform string
}

func (u *UserConf) Diff(dst *UserConf) (err error) {
	if u.CheckSum != dst.CheckSum {
		err = ecode.AccessKeyErr
		return
	}
	if u.ModuleKey != dst.ModuleKey {
		err = ecode.AccessKeyErr
		return
	}
	if u.Mid != dst.Mid {
		err = ecode.AccessKeyErr
		return
	}
	return
}

// Action job msg.
type Action struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}
