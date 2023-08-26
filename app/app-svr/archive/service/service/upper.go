package service

import (
	"context"

	"go-common/library/log"
	"go-common/library/time"

	"go-gateway/app/app-svr/archive/service/model/archive"

	"go-common/library/sync/errgroup"
)

// UpperCount upper count.
func (s *Service) UpperCount(c context.Context, mid int64) (count int, err error) {
	if count, err = s.arc.UpperCountCache(c, mid); err != nil {
		log.Error("s.arc.UpperCountCache(%d) error(%v)", mid, err)
	}
	if count >= 0 {
		return
	}
	if count, err = s.arc.ZCARDUpperCnt(c, mid); err != nil {
		log.Error("s.arc.UpperCount(%d) error(%v)", mid, err)
		return
	}
	if count == 0 {
		if count, err = s.arc.RawUpperCount(c, mid); err != nil {
			log.Error("s.arc.RawUpperCount(%d) error(%v)", mid, err)
			return
		}
	}
	_ = s.arc.AddUpperCountCache(c, mid, count)
	return
}

// UppersAidPubTime get aid and pime by mids
func (s *Service) UppersAidPubTime(c context.Context, mids []int64, pn, ps int) (mas map[int64][]*archive.AidPubTime, err error) {
	if pn < 1 {
		pn = 1
	}
	if ps < 1 {
		ps = 20
	}
	var (
		cachedUp []int64
		missed   []int64
		start    = (pn - 1) * ps
		end      = start + ps - 1
	)
	if cachedUp, missed, err = s.arc.ExpireUppersCountCache(c, mids); err != nil {
		return
	}
	var (
		eg        errgroup.Group
		cacheAidM map[int64][]*archive.AidPubTime
		missAidM  = make(map[int64][]*archive.AidPubTime)
	)
	eg.Go(func() (err error) {
		if len(cachedUp) == 0 {
			return
		}
		_, err = s.arc.ExpireUppersPassedCache(c, cachedUp)
		return
	})
	eg.Go(func() (err error) {
		if len(cachedUp) == 0 {
			return
		}
		if cacheAidM, err = s.arc.UppersPassedCacheWithScore(c, cachedUp, start, end); err != nil {
			log.Error("s.arc.UppersPassedCache(%v) error(%v)", cachedUp, err)
			return
		}
		return
	})
	eg.Go(func() (err error) {
		if len(missed) == 0 {
			return
		}
		var (
			ptimem     map[int64][]time.Time
			missM      map[int64][]int64
			copyrightM map[int64][]int8
		)
		if missM, ptimem, copyrightM, err = s.arc.RawUppersPassed(c, missed); err != nil {
			log.Error("s.arc.UppersPassed(%v) error(%v)", missed, err)
			return
		}
		for mid, aids := range missM {
			length := len(aids)
			if length == 0 || length < start {
				continue
			}
			var (
				cmid       = mid
				clength    = length
				cptime     = ptimem[mid]
				ccopyright = copyrightM[mid]
				caids      = aids
			)
			s.addCache(func() {
				_ = s.arc.AddUpperCountCache(context.TODO(), cmid, clength)
				_ = s.arc.AddUpperPassedCache(context.TODO(), cmid, caids, cptime, ccopyright)
			})
			if len(aids) > end+1 {
				missM[mid] = aids[start : end+1]
			} else {
				missM[mid] = aids[start:]
			}
			for i := start; i < len(missM[mid]); i++ {
				missAidM[mid] = append(missAidM[mid], &archive.AidPubTime{Aid: aids[i], PubDate: time.Time(cptime[i]), Copyright: ccopyright[i]})
			}
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return
	}
	mas = make(map[int64][]*archive.AidPubTime, len(mids))
	for mid, v := range cacheAidM {
		mas[mid] = v
	}
	for mid, v := range missAidM {
		mas[mid] = v
	}
	return
}
