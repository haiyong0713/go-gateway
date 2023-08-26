package college

import (
	"context"
	"fmt"
	"go-common/library/ecode"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/admin/model/college"

	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"github.com/pkg/errors"
)

const (
	// concurrency 并发为1
	concurrency = 1
	// tagBatch  批量tag
	tagBatch = 50
)

// collegeGetAndCreateTag 获取及创建tag
func (s *Service) collegeGetAndCreateTag(c context.Context, data []*college.College) (map[string]int64, error) {
	tagName := make([]string, 0)
	for _, v := range data {
		tagName = append(tagName, v.CollegeName)
	}
	getTag, err := s.getTagInfo(c, tagName)
	if err != nil && !xecode.EqualError(xecode.New(16001), err) {
		return nil, ecode.Error(ecode.RequestErr, "tag 获取失败")
	}
	createTag := s.filterCreateTag(c, tagName, getTag)
	newTag, err := s.createTag(c, createTag)
	if err != nil {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("tag 创建失败 err(%v)", err))
	}
	allTag := make(map[string]int64)
	for k, v := range getTag {
		allTag[k] = v
	}
	for k, v := range newTag {
		allTag[k] = v
	}
	return allTag, nil
}

func (s *Service) filterCreateTag(c context.Context, tagName []string, getTag map[string]int64) []string {
	tagList := make([]string, 0)
	for _, v := range tagName {
		if _, ok := getTag[v]; !ok {
			tagList = append(tagList, v)
		}
	}
	return tagList
}

// getTagInfo ...
func (s *Service) getTagInfo(c context.Context, tagName []string) (map[string]int64, error) {
	var times int
	patch := tagBatch
	times = len(tagName) / patch / concurrency
	tagInfo := make(map[string]int64)
	tagAllRes := make([]*tagrpc.TagsReply, 0)
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			b := batch
			i := index
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if start >= len(tagName) {
					return nil
				}
				reqTag := tagName[start:]
				end := start + patch
				if end < len(tagName) {
					reqTag = tagName[start:end]
				}
				if len(reqTag) > 0 {
					tagRes, err := s.tagRPC.TagByNames(c, &tagrpc.TagByNamesReq{Tnames: reqTag})
					if err != nil || tagRes == nil || tagRes.Tags == nil {
						err = errors.Wrapf(err, "s.tagRPC.TagByNames")
						return err
					}
					tagAllRes = append(tagAllRes, tagRes)
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			log.Errorc(c, "eg.Wait error(%v)", err)
			return nil, err
		}
	}
	for _, v := range tagName {
		for _, tagRes := range tagAllRes {
			if tagRes != nil {
				for _, value := range tagRes.Tags {
					if value != nil && value.Name == v {
						tagInfo[v] = value.Id
					}
				}
			}
		}
	}
	return tagInfo, nil
}

// createTag ...
func (s *Service) createTag(c context.Context, tagName []string) (map[string]int64, error) {
	var times int
	patch := tagBatch
	times = len(tagName) / patch / concurrency
	tagInfo := make(map[string]int64)
	tagAllRes := make([]*tagrpc.TagsReply, 0)
	var count int
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			b := batch
			i := index
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if start >= len(tagName) {
					return nil
				}
				reqTag := tagName[start:]
				end := start + patch
				if end < len(tagName) {
					reqTag = tagName[start:end]
				}
				if len(reqTag) > 0 {
					tagRes, err := s.tagRPC.AddTags(c, &tagrpc.AddTagsReq{Names: reqTag})
					if err != nil || tagRes == nil || tagRes.Tags == nil {
						err = errors.Wrapf(err, "s.tagRPC.TagByNames")
						return err
					}
					tagAllRes = append(tagAllRes, tagRes)
					count += len(tagRes.Tags)
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			log.Errorc(c, "eg.Wait error(%v)", err)
			return nil, err
		}
	}

	for _, v := range tagName {
		for _, tagRes := range tagAllRes {
			if tagRes != nil {
				for _, value := range tagRes.Tags {
					if value != nil && value.Name == v {
						tagInfo[v] = value.Id
					}
				}
			}
		}
	}

	for _, v := range tagName {
		if _, ok := tagInfo[v]; !ok {
			log.Errorc(c, "tag 创建失败(%v)", v)
			return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("tag 创建失败 (%v)", v))
		}
	}
	return tagInfo, nil
}

// CollegeBatchInsert ...
func (s *Service) CollegeBatchInsert(c context.Context, data []*college.College) (err error) {
	var times int
	patch := tagBatch
	tx, err := s.college.BeginTran(c)
	times = len(data) / patch / concurrency
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "tx.Rollback()  %v", r)
			err = ecode.Error(ecode.RequestErr, "保存失败")
			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(c, "tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(c, "tx.Commit() error(%v)", err)
		}
	}()
	for index := 0; index <= times; index++ {

		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			b := batch
			i := index
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if start >= len(data) {
					return nil
				}
				reqMids := data[start:]
				end := start + patch
				if end < len(data) {
					reqMids = data[start:end]
				}
				if len(reqMids) > 0 {
					tagInfo, err := s.collegeGetAndCreateTag(c, reqMids)
					if err != nil {
						return err
					}
					for k, v := range reqMids {
						if tagID, ok := tagInfo[v.CollegeName]; ok {
							reqMids[k].TagID = tagID
						} else {
							return ecode.Error(ecode.RequestErr, fmt.Sprintf("tag 创建失败,%v", v.CollegeName))

						}
					}
					err = s.college.BatchAddCollege(c, tx, reqMids)
					if err != nil {
						err = errors.Wrapf(err, "s.college.BatchAddCollege")
						return err
					}
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			log.Errorc(c, "eg.Wait error(%v)", err)
			return err
		}
	}

	return nil
}

// CollegeInsertOrUpdate ...
func (s *Service) CollegeInsertOrUpdate(c context.Context, data *college.College) (err error) {
	if data != nil {
		if data.TagID == 0 {
			tagInfo, err := s.collegeGetAndCreateTag(c, []*college.College{data})
			if err != nil {
				return err
			}
			if tagID, ok := tagInfo[data.CollegeName]; ok {
				data.TagID = tagID
			} else {
				return ecode.Error(ecode.RequestErr, "tag 创建失败")
			}
		}
		err := s.college.BacthInsertOrUpdateCollege(c, data)
		if err != nil {
			log.Errorc(c, "s.college.BacthInsertOrUpdateCollege")
			return err
		}
	}
	return nil
}
