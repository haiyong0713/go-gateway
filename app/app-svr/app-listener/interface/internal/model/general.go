package model

import "go-common/library/ecode"

type BmGenericResp struct {
	Code int64       `json:"code"`
	Msg  string      `json:"msg"`
	Data genericData `json:"data"`
}

func (bgp BmGenericResp) IsNormal() error {
	if bgp.Code != 0 {
		return ecode.New(int(bgp.Code))
	}
	return nil
}

type genericData []byte

func (gd *genericData) UnmarshalJSON(b []byte) error {
	*gd = b
	return nil
}

func (gd genericData) MarshalJSON() ([]byte, error) {
	if len(gd) == 0 {
		return []byte("{}"), nil
	}
	return gd, nil
}

func (gd genericData) String() string {
	return string(gd)
}

func (gd genericData) Bytes() []byte {
	if len(gd) == 0 {
		return []byte("{}")
	}
	return gd
}
