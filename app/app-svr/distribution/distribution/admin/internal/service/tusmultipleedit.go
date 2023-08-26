package service

import (
	"context"
	"sync"

	"go-gateway/app/app-svr/distribution/distribution/admin/internal/model"
	"go-gateway/app/app-svr/distribution/distribution/admin/internal/model/rename"
	tusm "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/tus"
	tme "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/tusmultipleedit"
	"go-gateway/app/app-svr/distribution/distribution/admin/tool"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	"github.com/pkg/errors"
)

func (s *Service) Overview(ctx context.Context) ([]*tme.Overview, error) {
	meta, ok := preferenceproto.TryGetPreference(multipleTusPreference)
	if !ok {
		return nil, errors.Wrapf(ecode.NothingFound, "Failed to fetch proto meta from %s", multipleTusPreference)
	}
	var (
		overviews  []*tme.Overview
		fieldNames []string
	)
	for _, v := range meta.ProtoDesc.GetFields() {
		fieldNames = append(fieldNames, v.GetFullyQualifiedName())
		overview := &tme.Overview{
			FieldName: v.GetFullyQualifiedName(),
			Name:      tool.RemoveCRLF(v.GetSourceInfo().GetLeadingComments()),
		}
		tusValues, _, err := parseTusValuesAndDmByFieldName(v.GetFullyQualifiedName())
		if err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		tusInfos, err := s.dao.BatchFetchTusInfos(ctx, tusValues)
		if err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		performances := wrapPerformance(tusInfos)
		overview.Performances = performances
		overviews = append(overviews, overview)
	}
	s.rename(ctx, fieldNames, &multipleTusEditRenamer{overviews: overviews})
	return overviews, nil
}

func wrapPerformance(tusInfos []*tusm.Info) []*tme.TusPerformance {
	performances := []*tme.TusPerformance{
		{
			TusValue: model.DefaultTusValue,
			Text:     model.DefaultTusValueName,
		},
	}
	for _, v := range tusInfos {
		performance := &tme.TusPerformance{
			TusValue: v.TusValue,
			Text:     v.Name,
		}
		performances = append(performances, performance)
	}
	return performances
}

func (s *Service) Performance(ctx context.Context, fieldName string, mid int64) (*tme.TusPerformance, error) {
	performance := &tme.TusPerformance{}
	tusValues, _, err := parseTusValuesAndDmByFieldName(fieldName)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	targetTusValue, err := s.tusEditDao.FetchTargetTusValue(ctx, tusValues, mid)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	var (
		eg         = errgroup.WithContext(ctx)
		tusInfo    = &tusm.Info{TusValue: model.DefaultTusValue, Name: model.DefaultTusValueName}
		renameInfo *rename.Rename
	)
	if targetTusValue != model.DefaultTusValue {
		eg.Go(func(ctx context.Context) error {
			reply, err := s.dao.BatchFetchTusInfos(ctx, []string{targetTusValue})
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			tusInfo = reply[0]
			return nil
		})
	}
	eg.Go(func(ctx context.Context) error {
		renameInfo, err = s.renameDao.FetchRenameInfo(ctx, fieldName)
		if err != nil {
			log.Error("%+v", err)
			return nil
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	performanceWithDefault := wrapPerformance([]*tusm.Info{tusInfo})
	func() {
		if targetTusValue == model.DefaultTusValue {
			performance = performanceWithDefault[0]
			return
		}
		performance = performanceWithDefault[1]
	}()
	if renameInfo != nil && len(renameInfo.TabNames) != 0 && renameInfo.TabNames[targetTusValue] != "" {
		performance.Text = renameInfo.TabNames[targetTusValue]
	}
	return performance, nil
}

func (s *Service) PerformanceSave(ctx context.Context, fieldName, targetTusValue string, mids []int64) error {
	tusValues, _, err := parseTusValuesAndDmByFieldName(fieldName)
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	migrators := s.classifyMidToMigrator(ctx, mids, tusValues, targetTusValue)
	for _, m := range migrators {
		if err := m.Migrate(ctx); err != nil {
			return err
		}
	}
	s.AsyncLog(ctx)
	return nil
}

type migrateTusValue struct {
	tusValue               string
	midsWithOriginTusValue map[int64]string
	migrate                func(ctx context.Context, tusValue string, mids map[int64]string) error
}

func (m migrateTusValue) Migrate(ctx context.Context) error {
	return m.migrate(ctx, m.tusValue, m.midsWithOriginTusValue)
}

type toDefaultTusValue struct {
	midsWithOriginTusValue map[int64]string
	toDefault              func(ctx context.Context, tusValue string, mids map[int64]string) error
}

func (d toDefaultTusValue) Migrate(ctx context.Context) error {
	return d.toDefault(ctx, model.DefaultTusValue, d.midsWithOriginTusValue)
}

type puInTusValue struct {
	tusValue               string
	midsWithOriginTusValue map[int64]string
	putIn                  func(ctx context.Context, tusValue string, mids map[int64]string) error
}

func (p puInTusValue) Migrate(ctx context.Context) error {
	return p.putIn(ctx, p.tusValue, p.midsWithOriginTusValue)
}

type migrator interface {
	Migrate(ctx context.Context) error
}

func (s *Service) classifyMidToMigrator(ctx context.Context, mids []int64, tusValues []string, targetTusValue string) []migrator {
	var (
		midsToMigrate          = make(map[int64]string)
		midsPutInTus           = make(map[int64]string)
		midsWithOriginTusValue = midsWithTusValues(make(map[int64]string))
		lock                   sync.Mutex
		eg                     = errgroup.WithContext(ctx)
	)
	//todo:后续数平提供批量查询接口
	for _, mid := range mids {
		tmpMid := mid
		eg.Go(func(ctx context.Context) error {
			originTusValue, err := s.tusEditDao.FetchTargetTusValue(ctx, tusValues, tmpMid)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			lock.Lock()
			midsWithOriginTusValue[tmpMid] = originTusValue
			lock.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("%+v", err)
		return nil
	}
	midsWithOriginTusValue.iter(func(mid int64, tusValue string) {
		if tusValue == targetTusValue {
			delete(midsWithOriginTusValue, mid)
		}
	})
	//如果targetTusValue是default，说明对这批mid都是删除操作，否则对于单个mid有可能是迁移，有可能新增到人群包中
	if targetTusValue == model.DefaultTusValue {
		return []migrator{
			toDefaultTusValue{
				midsWithOriginTusValue: midsWithOriginTusValue,
				toDefault:              s.tusEditDao.MigrateTusValueToDefaultWithMids,
			},
		}
	}
	midsWithOriginTusValue.iter(func(mid int64, tusValue string) {
		switch tusValue {
		case model.DefaultTusValue: //之前不在人群包中，现在要挪到人群包，对应数平新增操作
			midsPutInTus[mid] = tusValue
		default: //迁移操作
			midsToMigrate[mid] = tusValue
		}
	})
	var migrators []migrator
	if len(midsPutInTus) != 0 {
		migrators = append(migrators, puInTusValue{
			midsWithOriginTusValue: midsPutInTus,
			putIn:                  s.tusEditDao.PutinTusValueWithMids,
			tusValue:               targetTusValue,
		})
	}
	if len(midsToMigrate) != 0 {
		migrators = append(migrators, migrateTusValue{
			midsWithOriginTusValue: midsToMigrate,
			migrate:                s.tusEditDao.MigrateTusValueWithMids,
			tusValue:               targetTusValue,
		})
	}
	return migrators
}

type midsWithTusValues map[int64]string

func (m midsWithTusValues) iter(fn func(mid int64, tusValue string)) {
	for mid, tusValue := range m {
		fn(mid, tusValue)
	}
}
