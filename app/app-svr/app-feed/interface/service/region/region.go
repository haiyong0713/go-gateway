package region

import (
	"context"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/interface/model"
	"go-gateway/app/app-svr/app-feed/interface/model/tag"
)

var (
	_emptyHotTags = []*tag.Hot{}
	_emptyTags    = []*tag.Tag{}
	_banRegion    = map[int16]struct{}{
		// bangumi
		33:  {},
		32:  {},
		153: {},
		51:  {},
		152: {},
		// music
		29:  {},
		54:  {},
		130: {},
		// tech
		37: {},
		96: {},
		// movie
		145: {},
		146: {},
		147: {},
		// TV series
		15: {},
		34: {},
		86: {},
		// entertainment
		71:  {},
		137: {},
		131: {},
	}
)

// HotTags get hot tags of region id.
func (s *Service) HotTags(c context.Context, mid int64, rid int64, ver string, plat int8, now time.Time) (hs []*tag.Hot, version string, err error) {
	if hs, err = s.tg.Hots(c, mid, rid, now); err != nil {
		log.Error("tg.HotTags(%d) error(%v)", rid, err)
		return
	}
	if model.IsOverseas(plat) {
		for _, hot := range hs {
			if _, ok := _banRegion[hot.Rid]; ok {
				hot.Tags = _emptyTags
			}
		}
	}
	if len(hs) == 0 {
		hs = _emptyHotTags
		return
	}
	version = s.md5(hs)
	if ver == version {
		err = ecode.NotModified
	}
	return
}
