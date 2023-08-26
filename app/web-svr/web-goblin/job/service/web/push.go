package web

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-common/library/log"
	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web-goblin/job/model/web"
)

const (
	_aidBulkSize = 50
)

func (s *Service) LoadChangeOutArc() {
	s.cache.Do(context.Background(), func(ctx context.Context) {
		s.loadChangeOurArc()
	})
}

// nolint:gocognit
func (s *Service) loadChangeOurArc() {
	// 每天扫昨天修改的数据
	ctx := context.Background()
	nowTime := time.Now()
	lastDay := nowTime.AddDate(0, 0, -1).Format("2006-01-02")
	from, err := time.ParseInLocation("2006-01-02-15:04:05", lastDay+"-00:00:00", time.Local)
	if err != nil {
		log.Error("loadChangeOurArc time.Parse from lastDay:%s error:%v", lastDay, err)
		return
	}
	to, err := time.ParseInLocation("2006-01-02-15:04:05", lastDay+"-23:59:59", time.Local)
	if err != nil {
		log.Error("loadChangeOurArc time.Parse to lastDay:%s error:%v", lastDay, err)
		return
	}
	outArcs, err := s.dao.OutArcByMtime(ctx, from, to)
	if err != nil {
		log.Error("loadChangeOurArc s.dao.OutArcByMtime from:%v to:%v error:%v", from, to, err)
		return
	}
	if len(outArcs) == 0 {
		log.Error("loadChangeOurArc lastDay:%s from:%v to:%v no outArcs", lastDay, from, to)
		return
	}
	var (
		changeAids []int64
		delAids    []int64
	)
	for _, v := range outArcs {
		if v == nil || v.Aid <= 0 {
			continue
		}
		if v.IsDeleted == 1 {
			delAids = append(delAids, v.Aid)
			continue
		}
		changeAids = append(changeAids, v.Aid)
	}
	changeAidsLen := len(changeAids)
	delAidsLen := len(delAids)
	if changeAidsLen == 0 && delAidsLen == 0 {
		log.Error("loadChangeOurArc lastDay:%s from:%v to:%v no changed arcs", lastDay, from, to)
		return
	}
	log.Warn("loadChangeOurArc lastDay:%s from:%v to:%v len(changeAids):%d len(delAids):%d", lastDay, from, to, changeAidsLen, delAidsLen)
	archives := make(map[int64]*arcapi.Arc, changeAidsLen)
	for i := 0; i < changeAidsLen; i += _aidBulkSize {
		var partAids []int64
		if i+_aidBulkSize > changeAidsLen {
			partAids = changeAids[i:]
		} else {
			partAids = changeAids[i : i+_aidBulkSize]
		}
		arcReply, err := s.archiveClient.Arcs(ctx, &arcapi.ArcsRequest{Aids: partAids})
		time.Sleep(100 * time.Millisecond)
		if err != nil {
			log.Error("loadChangeOurArc s.arcGRPC.Arcs(%v) error(%+v)", partAids, err)
			continue
		}
		for _, v := range arcReply.GetArcs() {
			if v != nil && v.Aid > 0 && v.IsNormal() {
				archives[v.Aid] = v
			}
		}
	}
	log.Warn("loadChangeOurArc lastDay:%s from:%v to:%v len(archives):%d", lastDay, from, to, len(archives))
	var pushArcs []*web.PushArc
	var forbidCount int64
	for _, arc := range archives {
		tmp := new(web.PushArc)
		tmp.CopyFromArc(arc, s.arcTypes)
		if tmp.ForbidArc() {
			forbidCount++
			continue
		}
		pushArcs = append(pushArcs, tmp)
	}
	var delArcs []*web.PushDelArc
	for _, aid := range delAids {
		tmp := new(web.PushDelArc)
		tmp.FmtFromAid(aid)
		delArcs = append(delArcs, tmp)
	}
	if len(pushArcs) > 0 {
		func() {
			pushByte, err := json.Marshal(pushArcs)
			if err != nil {
				log.Error("loadChangeOurArc pushArcs json.Marshal error:%v", err)
				return
			}
			pushFileName := fmt.Sprintf("push_arc_%s.json", lastDay)
			pushPath, err := s.dao.UploadBFS(ctx, pushFileName, pushByte)
			if err != nil {
				log.Error("loadChangeOurArc pushArcs UploadBFS error:%v", err)
				return
			}
			// 下载原文文件
			preContent, err := s.dao.ReadURLContent(ctx, s.c.Rule.PushArcBfsURL)
			if err != nil {
				log.Error("loadChangeOurArc pushArcs ReadURLContent url:%s error:%v", s.c.Rule.PushArcBfsURL, err)
				return
			}
			// 解析源文件内容
			preData := new(web.BaiduSitemap)
			if err = json.Unmarshal(preContent, &preData); err != nil {
				log.Error("loadChangeOurArc pushArcs json.Unmarshal preContent:%s error:%v", string(preContent), err)
				return
			}
			for _, v := range preData.Sitemapindex {
				if v != nil && v.Sitemap != nil {
					if v.Sitemap.Lastmod == lastDay {
						log.Warn("loadChangeOurArc pushArcs lastDay:%s exist", lastDay)
						return
					}
				}
			}
			// 加入新文件地址
			newSiteMap := &web.BaiduSitemapItem{Sitemap: &web.BaiduSiteMapDetail{Loc: pushPath, Lastmod: lastDay}}
			preData.Sitemapindex = append(preData.Sitemapindex, newSiteMap)
			// 覆盖上传
			newSiteMapByte, err := json.Marshal(preData)
			if err != nil {
				log.Error("loadChangeOurArc pushArcs json.Marshal newSiteMap:%+v error:%v", newSiteMap, err)
				return
			}
			if _, err := s.dao.UploadBFS(ctx, s.c.Rule.PushArcFileName, newSiteMapByte); err != nil {
				log.Error("loadChangeOurArc pushArcs UploadBFS newSiteMap error:%v", err)
				return
			}
		}()
	}
	if len(delArcs) > 0 {
		func() {
			delByte, err := json.Marshal(delArcs)
			if err != nil {
				log.Error("loadChangeOurArc delArcs json.Marshal error:%v", err)
				return
			}
			delFileName := fmt.Sprintf("del_arc_%s.json", lastDay)
			delPath, err := s.dao.UploadBFS(ctx, delFileName, delByte)
			if err != nil {
				log.Error("loadChangeOurArc delArcs UploadBFS error:%v", err)
				return
			}
			// 下载原文文件
			preContent, err := s.dao.ReadURLContent(ctx, s.c.Rule.DelArcBfsURL)
			if err != nil {
				log.Error("loadChangeOurArc delArcs ReadURLContent url:%s error:%v", s.c.Rule.DelArcBfsURL, err)
				return
			}
			// 解析源文件内容
			preData := new(web.BaiduSitemap)
			if err = json.Unmarshal(preContent, &preData); err != nil {
				log.Error("loadChangeOurArc delArcs json.Unmarshal preContent:%s error:%v", string(preContent), err)
				return
			}
			for _, v := range preData.Sitemapindex {
				if v != nil && v.Sitemap != nil {
					if v.Sitemap.Lastmod == lastDay {
						log.Warn("loadChangeOurArc delArcs lastDay:%s exist", lastDay)
						return
					}
				}
			}
			// 加入新文件地址
			newSiteMap := &web.BaiduSitemapItem{Sitemap: &web.BaiduSiteMapDetail{Loc: delPath, Lastmod: lastDay}}
			preData.Sitemapindex = append(preData.Sitemapindex, newSiteMap)
			// 覆盖上传
			newSiteMapByte, err := json.Marshal(preData)
			if err != nil {
				log.Error("loadChangeOurArc delArcs json.Marshal newSiteMap:%+v error:%v", newSiteMap, err)
				return
			}
			if _, err := s.dao.UploadBFS(ctx, s.c.Rule.DelArcFileName, newSiteMapByte); err != nil {
				log.Error("loadChangeOurArc delArcs UploadBFS newSiteMap error:%v", err)
				return
			}
		}()
	}
	log.Warn("loadChangeOurArc success lastDay:%s from:%v to:%v len(pushArcs):%d forbidCount:%d len(delArcs):%d", lastDay, from, to, len(pushArcs), forbidCount, len(delArcs))
}

func (s *Service) loadArcTypes() {
	typesReply, err := s.archiveClient.Types(context.Background(), &arcapi.NoArgRequest{})
	if err != nil {
		log.Error("loadArcType s.archiveClient.Types error:%+v", err)
		return
	}
	if typesReply == nil || len(typesReply.Types) == 0 {
		log.Warn("loadArcType typesReply data(%+v) wrong", typesReply)
		return
	}
	s.arcTypes = typesReply.Types
}
