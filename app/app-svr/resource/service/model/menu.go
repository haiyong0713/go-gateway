package model

import (
	"encoding/json"
	"strconv"

	"go-common/library/log"
	xtime "go-common/library/time"
)

const (
	InitSkinExtKey  = "skin_ext_%d"
	SkinExtCacheKey = "skin_ext"
)

type Menu struct {
	TabID       int64                    `json:"tab_id,omitempty"`
	Name        string                   `json:"name,omitempty"`
	Img         string                   `json:"img,omitempty"`
	Icon        string                   `json:"icon,omitempty"`
	Color       string                   `json:"color,omitempty"`
	ID          int64                    `json:"id,omitempty"`
	Plat        int                      `json:"-"`
	CType       int                      `json:"-"`
	CValue      string                   `json:"-"`
	PlatVersion json.RawMessage          `json:"-"`
	STime       xtime.Time               `json:"-"`
	ETime       xtime.Time               `json:"-"`
	Status      int                      `json:"-"`
	Badge       string                   `json:"-"`
	Versions    map[int32][]*MenuVersion `json:"-"`
}

type MenuVersion struct {
	PlatStr   string `json:"plat,omitempty"`
	BuildStr  string `json:"build,omitempty"`
	Condition string `json:"conditions,omitempty"`
	Plat      int32  `json:"-"`
	Build     int    `json:"-"`
}

func (m *Menu) Change() bool {
	m.Icon = m.Badge
	var vs []*MenuVersion
	if err := json.Unmarshal(m.PlatVersion, &vs); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", m.PlatVersion, err)
		return false
	}
	vm := make(map[int32][]*MenuVersion, len(vs))
	for _, v := range vs {
		if v.PlatStr == "" || v.BuildStr == "" {
			continue
		}
		//nolint:gosec
		plat, err := strconv.Atoi(v.PlatStr)
		if err != nil {
			log.Error("strconv.Atoi(%s) error(%v)", v.PlatStr, err)
			continue
		}
		build, err := strconv.Atoi(v.BuildStr)
		if err != nil {
			log.Error("strconv.Atoi(%s) error(%v)", v.BuildStr, err)
			continue
		}
		vm[int32(plat)] = append(vm[int32(plat)], &MenuVersion{Plat: int32(plat), Build: build, Condition: v.Condition})
	}
	m.Versions = vm
	if m.CType == 1 {
		var err error
		if m.ID, err = strconv.ParseInt(m.CValue, 10, 64); err != nil {
			log.Error("strconv.ParseInt(%s) error(%v)", m.CValue, err)
			return false
		}
	}
	return true
}
