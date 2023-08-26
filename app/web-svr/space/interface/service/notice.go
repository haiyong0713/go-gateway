package service

import (
	"context"
	"encoding/json"
	"html/template"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/space/interface/model"

	"golang.org/x/sync/errgroup"
)

const (
	_noticeTable   = "member_up_notice"
	_official      = "space_official"
	_topPhotoTable = "member_topphoto"
	//_uploadPhotoTable = "member_upload_topphoto"
)

// Notice get notice.
func (s *Service) Notice(c context.Context, mid int64) (res string, err error) {
	if _, ok := s.noNoticeMids[mid]; ok {
		return
	}
	var notice *model.Notice
	if notice, err = s.dao.Notice(c, mid); err != nil {
		return
	}
	if notice.IsForbid == _noticeForbid {
		notice.Notice = ""
	}
	res = template.HTMLEscapeString(notice.Notice)
	return
}

// SetNotice set notice.
func (s *Service) SetNotice(c context.Context, mid int64, notice string) (err error) {
	eg := errgroup.Group{}
	eg.Go(func() error {
		info, e := s.realName(c, mid)
		if e != nil {
			return e
		}
		if info.Silence == _silenceForbid {
			e = ecode.UserDisabled
			return e
		}
		return nil
	})
	eg.Go(func() error {
		e := s.Filter(c, []string{notice})
		if e != nil {
			return e
		}
		return nil
	})
	eg.Go(func() error {
		preData, e := s.dao.Notice(c, mid)
		if e != nil {
			return e
		}
		if notice == preData.Notice {
			e = ecode.NotModified
			return e
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		return
	}
	if err = s.dao.SetNotice(c, mid, notice); err != nil {
		log.Error("s.dao.SetNotice(%d,%s) error(%v)", mid, notice, err)
		return
	}
	s.cache.Do(c, func(c context.Context) {
		_ = s.dao.AddCacheNotice(c, mid, &model.Notice{Notice: notice})
	})
	return
}

// ClearNotice .
func (s *Service) ClearNotice(c context.Context, msg string) (err error) {
	var m struct {
		Table string `json:"table"`
		Old   struct {
			Mid      int64  `json:"mid"`
			Notice   string `json:"notice"`
			IsForbid int    `json:"is_forbid"`
		} `json:"old,omitempty"`
		New struct {
			Mid      int64  `json:"mid"`
			Notice   string `json:"notice"`
			IsForbid int    `json:"is_forbid"`
		} `json:"new,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil || m.Table == "" {
		log.Error("ClearCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("ClearCache json.Unmarshal msg(%s)", msg)
	if strings.HasPrefix(m.Table, _noticeTable) && (m.Old.IsForbid != m.New.IsForbid || m.Old.Notice != m.New.Notice) {
		if err = s.dao.DelCacheNotice(c, m.New.Mid); err != nil {
			log.Error("s.dao.DelCacheNotice mid(%d) error(%v)", m.New.Mid, err)
		}
	}
	return
}

// ClearOfficial .
func (s *Service) ClearOfficial(c context.Context, msg string) (err error) {
	var m struct {
		New struct {
			Uid int64 `json:"uid"`
		} `json:"new,omitempty"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("ClearOfficial json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	if m.New.Uid == 0 {
		log.Error("ClearOfficial msg(%s) error(m.New.Uid == 0)", msg)
		return nil
	}
	//clear cache
	if err = s.dao.DelOfficialCache(c, m.New.Uid); err != nil {
		log.Error("ClearOfficial DelOfficialCache msg(%s) error(%v)", msg, err)
		err = nil
	}
	log.Info("ClearOfficial uid(%d) json.Unmarshal msg(%s) success", m.New.Uid, msg)
	return
}

// ClearCache del match and object cache
func (s *Service) ClearCache(c context.Context, msg string) (err error) {
	var m struct {
		Table string `json:"table"`
	}
	log.Info("ClearCache msg(%s)", msg)
	if err = json.Unmarshal([]byte(msg), &m); err != nil || m.Table == "" {
		log.Error("ClearCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	if strings.HasPrefix(m.Table, _noticeTable) {
		return s.ClearNotice(c, msg)
	} else if strings.HasPrefix(m.Table, _topPhotoTable) {
		return s.ClearTopPhotoArc(c, msg)
	} else if m.Table == _official {
		return s.ClearOfficial(c, msg)
	}
	//else if m.Table == _uploadPhotoTable {
	//	return s.clearMemberUploadTopPhoto(c, msg)
	//}
	return
}
