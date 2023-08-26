package archive

import (
	"context"

	videoUpOpen "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/time"

	"go-gateway/app/app-svr/ugc-season/job/model/archive"
)

// Season get season by season_id
func (d *Dao) Season(c context.Context, sid int64) (s *archive.Season, err error) {
	info, err := d.videoUpOpenClient.GetSeasonInfo(c, &videoUpOpen.GetSeasonInfoReq{Id: sid})
	if err != nil {
		log.Error("日志告警 db to grpc GetSeasonInfo sid(%d) error(%+v)", sid, err)
		if ecode.EqualError(ecode.NothingFound, err) {
			return nil, nil
		}
		return
	}

	res := &archive.Season{
		SeasonID:  sid,
		Title:     info.Title,
		Desc:      info.Desc,
		Cover:     info.Cover,
		Mid:       info.Mid,
		Attribute: info.Attribute,
		SignState: int32(info.SignState),
		Show:      int32(info.Show),
		State:     int32(info.State),
		EpNum:     info.EpNum,
	}
	return res, nil
}

// Sections get sections by season_id
func (d *Dao) Sections(c context.Context, sid int64) (ss []*archive.SeasonSection, err error) {
	secs, err := d.videoUpOpenClient.GetSeasonSection(c, &videoUpOpen.GetSeasonSectionReq{SeasonId: sid})
	if err != nil {
		log.Error("日志告警 db to grpc GetSeasonSection sid(%d) error(%+v)", sid, err)
		if ecode.EqualError(ecode.NothingFound, err) {
			return nil, nil
		}
		return
	}

	res := make([]*archive.SeasonSection, 0, len(secs.SeasonSection))
	for _, sec := range secs.SeasonSection {
		info := &archive.SeasonSection{
			SectionID: sec.Id,
			SeasonID:  sec.SeasonId,
			Title:     sec.Title,
			Type:      int32(sec.Type),
			Order:     sec.Order,
			Show:      int32(sec.Show),
			State:     int32(sec.State),
		}
		res = append(res, info)
	}
	return res, nil
}

// Episodes get episodes by season_id
func (d *Dao) Episodes(c context.Context, sid int64) (se []*archive.SeasonEp, err error) {
	eps, err := d.videoUpOpenClient.GetSeasonEpisode(c, &videoUpOpen.GetSeasonEpisodeReq{SeasonId: sid})
	if err != nil {
		log.Error("日志告警 db to grpc GetSeasonEpisode sid(%d) error(%+v)", sid, err)
		if ecode.EqualError(ecode.NothingFound, err) {
			return nil, nil
		}
		return
	}

	res := make([]*archive.SeasonEp, 0, len(eps.SeasonEpisode))
	for _, episode := range eps.SeasonEpisode {
		info := &archive.SeasonEp{
			EpID:      episode.Id,
			SeasonID:  episode.SeasonId,
			SectionID: episode.SectionId,
			Title:     episode.Title,
			AID:       episode.Aid,
			CID:       episode.Cid,
			Order:     episode.Order,
			Attribute: episode.Attribute,
			Show:      int32(episode.Show),
			State:     int32(episode.State),
		}
		res = append(res, info)
	}
	return res, nil
}

// SeasonMaxPtime get season max ptime by aids
func (d *Dao) SeasonMaxPtime(c context.Context, sid int64, aids []int64) (maxPtime time.Time, err error) {
	info, err := d.videoUpOpenClient.GetArchiveMaxPubtime(c, &videoUpOpen.GetArchiveMaxPubtimeReq{Aids: aids})
	if err != nil {
		log.Error("日志告警 db to grpc GetArchiveMaxPubtime sid(%d) error(%+v)", sid, err)
		if ecode.EqualError(ecode.NothingFound, err) {
			return 0, nil
		}
		return
	}
	return time.Time(info.Pubtime), nil
}
