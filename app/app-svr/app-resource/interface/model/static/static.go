package static

import (
	"strings"

	xtime "go-common/library/time"
)

// Static
type Static struct {
	Sid       int             `json:"sid"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	URL       string          `json:"url"`
	Hash      string          `json:"hash"`
	ImageHash string          `json:"imageHash"`
	Size      int             `json:"size"`
	Plat      int8            `json:"-"`
	Build     int             `json:"-"`
	Condition string          `json:"-"`
	Start     xtime.Time      `json:"-"`
	End       xtime.Time      `json:"-"`
	PreTime   xtime.Time      `json:"-"`
	Mids      string          `json:"-"`
	Mtime     xtime.Time      `json:"-"`
	Whitelist map[int64]int64 `json:"-"`
}

func (s *Static) StaticChange(eggType string) {
	var (
		urls    = strings.Split(s.URL, "/")
		urlsLen = len(urls)
	)
	if urlsLen == 0 {
		return
	}
	s.Name = urls[urlsLen-1]
	s.ImageHash = s.Hash
	s.Type = eggType
}
