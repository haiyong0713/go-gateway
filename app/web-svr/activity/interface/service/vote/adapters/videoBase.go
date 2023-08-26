package adapters

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/dao/vote"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup"
	archive "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/interface/client"
)

const (
	_aidBulkSize = 50
)

// video: 视频
type video struct {
	Id         int64          `json:"id"`
	Author     archive.Author `json:"author"`
	Stat       archive.Stat   `json:"stat"`
	VideoCover string         `json:"video_cover"`
	VideoTitle string         `json:"video_title"`
	VideoUrl   string         `json:"video_url"`
	Duration   int64          `json:"duration"`
}

type operConfig struct {
	VideoIdsStr string `json:"video-ids"`
}

func (i *video) GetName() string {
	return i.VideoTitle
}

func (i *video) GetId() int64 {
	return i.Id
}

func (i *video) GetSearchField1() string {
	return i.VideoTitle
}

func (i *video) GetSearchField2() string {
	return i.Author.Name
}

func (i *video) GetSearchField3() string {
	return ""
}

func archives(c context.Context, aids []int64) (archives map[int64]*archive.Arc, err error) {
	var (
		mutex         = sync.Mutex{}
		aidsLen       = len(aids)
		group, errCtx = errgroup.WithContext(c)
	)
	archives = make(map[int64]*archive.Arc, aidsLen)
	for i := 0; i < aidsLen; i += _aidBulkSize {
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		group.Go(func() (err error) {
			var arcs *archive.ArcsReply
			arg := &archive.ArcsRequest{Aids: partAids}
			if arcs, err = client.ArchiveClient.Arcs(errCtx, arg); err != nil || arcs == nil {
				log.Error("Vote.Archives (%v) error(%v)", partAids, err)
				return
			}
			mutex.Lock()
			for _, v := range arcs.Arcs {
				archives[v.Aid] = v
			}
			mutex.Unlock()
			return
		})
	}
	err = group.Wait()
	return
}

func getVoteVideoInfoByAids(c context.Context, aids []int64) (res []vote.DataSourceItem, err error) {
	res = make([]vote.DataSourceItem, 0, len(aids))
	archiveInfos, err := archives(c, aids)
	if err != nil {
		return
	}
	for _, aid := range aids {
		tmp, ok := archiveInfos[aid]
		if !ok {
			continue
		}
		if !tmp.IsNormal() {
			log.Errorc(c, "vote.adapters getVoteVideoInfoByAids skip aid %v because state=%v", tmp.Aid, tmp.State)
			continue
		}
		res = append(res, &video{
			Id:         tmp.Aid,
			Author:     tmp.Author,
			VideoCover: tmp.Pic,
			VideoTitle: tmp.Title,
			VideoUrl:   tmp.ShortLinkV2,
			Duration:   tmp.Duration,
			Stat:       tmp.Stat,
		})
	}
	return
}
