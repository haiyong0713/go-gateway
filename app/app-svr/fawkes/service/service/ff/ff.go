package ff

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"strings"
	"sync"

	"go-common/library/sync/errgroup"
	"go-common/library/xstr"

	xsql "go-common/library/database/sql"

	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	ffmdl "go-gateway/app/app-svr/fawkes/service/model/ff"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// AppFFWhithlist get overrall whithlist.
func (s *Service) AppFFWhithlist(c context.Context, appKey, env string) (res []*ffmdl.Whitch, err error) {
	if res, err = s.fkDao.AppFFWhithlist(c, appKey, env); err != nil {
		log.Error("%v", err)
	}
	return
}

// AppFFWhithlistAdd add overrall whitchlist.
func (s *Service) AppFFWhithlistAdd(c context.Context, appKey, env, userName string, mids []int64) (err error) {
	var tx *xsql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.fkDao.TxAddFFWhithlist(tx, appKey, env, userName, mids); err != nil {
		log.Error("%v", err)
	}
	return
}

// AppFFWhithlistDel del overrall whitchlist.
func (s *Service) AppFFWhithlistDel(c context.Context, appKey, env, userName string, mid int64) (err error) {
	var tx *xsql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.fkDao.TxDelFFWhithlist(tx, appKey, env, mid); err != nil {
		log.Error("%v", err)
	}
	return
}

// AppFFConfigSet set app ff config.
// nolint:gocognit
func (s *Service) AppFFConfigSet(c context.Context, appKey, env, userName, key, desc, status, salt, bucket, version,
	unVersion, romVersion, brand, unBrand, network, isp, channel, whith, blackList, blackMid string, bucketCount int64) (err error) {
	var tx *xsql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	var origin *ffmdl.FF
	if origin, err = s.fkDao.AppFFConfig(c, appKey, env, key); err != nil {
		log.Error("%v", err)
		return
	}
	if origin == nil {
		if _, err = s.fkDao.TxSetFFConfig(tx, appKey, env, userName, key, desc, status, salt, bucket, version, unVersion,
			romVersion, brand, unBrand, network, isp, channel, whith, blackMid, blackList, bucketCount); err != nil {
			log.Error("%v", err)
		}
	} else {
		if origin.Desc == desc && origin.Status == status && origin.Salt == salt && origin.Bucket == bucket &&
			origin.BucketCount == bucketCount && origin.Version == version &&
			origin.UnVersion == unVersion && origin.RomVersion == romVersion && origin.Brand == brand &&
			origin.UnBrand == unBrand && origin.Network == network && origin.ISP == isp && origin.Channel == channel &&
			origin.Whith == whith && origin.BlackMid == blackMid && origin.BlackList == blackList {
			return
		}
		var state int
		if origin.State == ffmdl.FFStatAdd {
			state = ffmdl.FFStatAdd
		} else {
			state = ffmdl.FFStatModify
		}
		if _, err = s.fkDao.TxUpFFConfig(tx, appKey, env, userName, key, desc, status, salt, bucket, version, unVersion,
			romVersion, brand, unBrand, network, isp, channel, whith, blackMid, blackList, bucketCount, state); err != nil {
			log.Error("%v", err)
		}
	}
	return
}

// AppFFList get app ff list.
func (s *Service) AppFFList(c context.Context, appKey, env, userName, filterKey string, pn, ps int) (res *ffmdl.ResultFF, err error) {
	var total int
	if total, err = s.fkDao.FFCount(c, appKey, env, filterKey); err != nil {
		log.Error("%v", err)
		return
	}
	res = &ffmdl.ResultFF{
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    pn,
			Ps:    ps,
		},
	}
	if res.Items, err = s.fkDao.FFList(c, appKey, env, filterKey, pn, ps); err != nil {
		log.Error("%v", err)
		return
	}
	if res.ModifyNum, err = s.fkDao.FFModiyCount(c, appKey, env); err != nil {
		log.Error("%v", err)
	}
	return
}

// AppFFPublish app ff publish.
// nolint:gocognit
func (s *Service) AppFFPublish(c context.Context, appKey, env, userName, desc string) (err error) {
	log.Info("AppFFPublish appkey(%v) env(%v) start", appKey, env)
	var (
		appChannels   []*appmdl.Channel
		appChannelMap = make(map[int64]*appmdl.Channel)
	)
	log.Info("AppFFPublish appkey(%v) env(%v) start get app channle", appKey, env)
	if appChannels, err = s.fkDao.AppChannelList(context.Background(), appKey, "", "", "", -1, -1, -1); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("AppFFPublish appkey(%v) env(%v) success get app channle %+v", appKey, env, appChannels)
	for _, appChannel := range appChannels {
		appChannelMap[appChannel.ID] = appChannel
	}
	log.Info("AppFFPublish appkey(%v) env(%v) form to appChannelMap %+v", appKey, env, appChannelMap)
	var ffconfigs []*ffmdl.FF
	log.Info("AppFFPublish appkey(%v) env(%v) start get ff configs", appKey, env)
	if ffconfigs, err = s.fkDao.AppFFConfigs(context.Background(), appKey, env); err != nil || len(ffconfigs) == 0 {
		log.Error("%v or ffconfigs is nil", err)
		return
	}
	log.Info("AppFFPublish appkey(%v) env(%v) success get ff configs %+v", appKey, env, ffconfigs)
	var tx *xsql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	var fvid int64
	log.Info("AppFFPublish appkey(%v) env(%v) start add ff publish user(%v) desc(%v)", appKey, env, userName, desc)
	if fvid, err = s.fkDao.TxAddFFPublish(tx, appKey, env, userName, desc); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("AppFFPublish appkey(%v) env(%v) success add ff publish fvid(%v)", appKey, env, fvid)
	var (
		sqls         []string
		args         []interface{}
		publishItems []*ffmdl.PublishItem
	)
	for _, ffconfig := range ffconfigs {
		log.Info("AppFFPublish appkey(%v) env(%v) start form ff file content", appKey, env)
		sqls = append(sqls, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
		args = append(args, ffconfig.AppKey, ffconfig.Env, fvid, ffconfig.Key, ffconfig.Desc, ffconfig.Status,
			ffconfig.Salt, ffconfig.Bucket, ffconfig.BucketCount, ffconfig.Version, ffconfig.UnVersion, ffconfig.RomVersion,
			ffconfig.Brand, ffconfig.UnBrand, ffconfig.Network, ffconfig.ISP, ffconfig.Channel, ffconfig.Whith, ffconfig.BlackMid,
			ffconfig.BlackList, ffconfig.Operator, ffconfig.State)
		log.Info("AppFFPublish appkey(%v) env(%v) success form ff file content(%+v)", appKey, env, args)
		log.Info("AppFFPublish appkey(%v) env(%v) ff config state(%v)", appKey, env, ffconfig.State)
		if ffconfig.State == ffmdl.FFStatDel {
			continue
		}
		if ffconfig.Whith == "" && ffconfig.Version == "" && ffconfig.UnVersion == "" && ffconfig.RomVersion == "" &&
			ffconfig.Channel == "" && ffconfig.Brand == "" && ffconfig.UnBrand == "" && ffconfig.Network == "" &&
			ffconfig.ISP == "" && ffconfig.BlackMid == "" && ffconfig.BlackList == "" && ffconfig.Status == "" {
			continue
		}
		log.Info("AppFFPublish appkey(%v) env(%v) start form ff trees", appKey, env)
		trees := s.formTree(appKey, env, appChannelMap, ffconfig)
		var bucket []string
		if ffconfig.Bucket != "" {
			bs := strings.Split(ffconfig.Bucket, ",")
			// nolint:gomnd
			if len(bs) == 2 {
				if bs[0] == bs[1] {
					bucket = append(bucket, bs[0])
				} else {
					bucket = append(bucket, bs[0])
					bucket = append(bucket, bs[1])
				}
			}
		}
		trees = append(trees, &ffmdl.PublishTree{
			Prop:        "bucket",
			Salt:        ffconfig.Salt,
			Bucket:      strings.Join(bucket, "~"),
			BucketCount: ffconfig.BucketCount,
			Logic:       ffconfig.Status,
		})
		pt := &ffmdl.PublishTree{}
		s.formTree2(pt, trees, len(trees))
		log.Info("AppFFPublish appkey(%v) env(%v) success form ff publish trees(%+v)", appKey, env, trees)
		var bts []*ffmdl.PublishTree
		if ffconfig.BlackList != "" {
			log.Info("AppFFPublish appkey(%v) env(%v) start form ff black trees(%+v)", appKey, env, ffconfig.BlackList)
			var (
				btrees []*ffmdl.PublishTree
				fbs    []*ffmdl.FF
			)
			if err = json.Unmarshal([]byte(ffconfig.BlackList), &fbs); err != nil {
				log.Error("%v", err)
				return
			}
			for _, fb := range fbs {
				btrees = s.formTree(appKey, env, appChannelMap, fb)
				bt := &ffmdl.PublishTree{}
				s.formTree2(bt, btrees, len(btrees))
				bts = append(bts, bt)
			}
			log.Info("AppFFPublish appkey(%v) env(%v) success form ff trees", appKey, env)
		}
		publishItem := &ffmdl.PublishItem{
			Name:        ffconfig.Key,
			WhitchList:  ffconfig.Whith,
			BlackList:   ffconfig.BlackMid,
			PublishTree: pt,
			BlackTree:   bts,
		}
		publishItems = append(publishItems, publishItem)
	}
	log.Info("AppFFPublish appkey(%v) env(%v) start insert ff file(%+v)", appKey, env, args)
	if _, err = s.fkDao.TxAddFFConfigFile(tx, sqls, args); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("AppFFPublish appkey(%v) env(%v) success insert ff file", appKey, env)
	log.Info("AppFFPublish appkey(%v) env(%v) start flush config(del state =-1 and modify all to 3)", appKey, env)
	if _, err = s.fkDao.TxFlushFFConfig(tx, appKey, env); err != nil {
		log.Error("%v", err)
		return
	}
	if _, err = s.fkDao.TxUpFFConfigState(tx, appKey, env, ffmdl.FFStatePublish); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("AppFFPublish appkey(%v) env(%v) success flush config", appKey, env)
	var app *appmdl.APP
	log.Info("AppFFPublish appkey(%v) env(%v) start get app info", appKey, env)
	if app, err = s.fkDao.AppPass(context.Background(), appKey); err != nil {
		log.Error("%v", err)
		return
	}
	var (
		wl   []*ffmdl.Whitch
		mids []int64
	)
	log.Info("AppFFPublish appkey(%v) env(%v) start get whith list", appKey, env)
	if wl, err = s.fkDao.AppFFWhithlist(context.Background(), appKey, env); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("AppFFPublish appkey(%v) env(%v) success get whith list %+v", appKey, env, wl)
	for _, w := range wl {
		mids = append(mids, w.MID)
	}
	var fvids = strconv.FormatInt(fvid, 10)
	ffp := &ffmdl.Publish{
		WhitchList:  xstr.JoinInts(mids),
		PlatForm:    app.Platform,
		VID:         fvids,
		PublishTree: publishItems,
	}
	var ffpb []byte
	if ffpb, err = json.Marshal(ffp); err != nil {
		log.Error("%v", err)
		return
	}
	var (
		folder   = path.Join(s.c.LocalPath.LocalDir, "ff", appKey, env, fvids)
		filename = fmt.Sprintf("%v_%v_%v", appKey, "default", fvids)
	)
	log.Info("AppFFPublish appkey(%v) env(%v) start write ff_file folder(%v) filename(%v) content(%v)",
		appKey, env, folder, filename, string(ffpb))
	var zipFilePath string
	if zipFilePath, err = s.fkDao.WriteConfigFile(folder, filename, ffpb); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("AppFFPublish appkey(%v) env(%v) success write ff file %v", appKey, env, zipFilePath)
	log.Info("AppFFPublish appkey(%v) env(%v) start up bfs %v", appKey, env, zipFilePath)
	var cdnURL, md5Str string
	if cdnURL, md5Str, err = s.fkDao.UpBFSV2(path.Join("ff", env), zipFilePath, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("AppFFPublish appkey(%v) env(%v) start diff", appKey, env)
	var ffPublishs []*ffmdl.ConfigPublish
	if ffPublishs, err = s.fkDao.AppFFHistory(c, appKey, env, -1, -1); err != nil {
		log.Error("%v", err)
		return
	}
	var diffs = make(map[string]string)
	for _, ffPublish := range ffPublishs {
		var (
			patchPath string
			filename  = fmt.Sprintf("%v_%v_%v_%v.patch", appKey, "default", ffPublish.ID, fvids)
		)
		log.Info("AppFFPublish appkey(%v) env(%v) start diff file %v %v %v", appKey, env, filename, zipFilePath, ffPublish.LocalPath)
		if patchPath, err = s.fkDao.DiffCmd(folder, filename, zipFilePath, ffPublish.LocalPath); err != nil {
			log.Error("%v", err)
			err = nil
			continue
		}
		log.Info("AppFFPublish appkey(%v) env(%v) success diff file", appKey, env)
		log.Info("AppFFPublish appkey(%v) env(%v) start up bfs diff file %v", appKey, env, patchPath)
		patch := &model.Diff{}
		if patch.URL, patch.MD5, err = s.fkDao.UpBFSV2(path.Join("config", env), patchPath, appKey); err != nil {
			log.Error("%v", err)
			continue
		}
		log.Info("AppFFPublish appkey(%v) env(%v) success up bfs diff url(%v) md5(%v)", appKey, env, patch.URL, patch.MD5)
		diffs[strconv.FormatInt(ffPublish.ID, 10)] = patch.URL
		//nolint:gomnd
		if len(diffs) == 3 {
			break
		}
	}
	var ds []byte
	if len(diffs) > 0 {
		if ds, err = json.Marshal(diffs); err != nil {
			log.Error("%v", err)
			err = nil
		}
	}
	log.Info("AppFFPublish appkey(%v) env(%v) success diff", appKey, env)
	log.Info("AppFFPublish appkey(%v) env(%v) start update url(%v) localPath(%v) md5(%v) diffs(%v) fvid(%v)",
		appKey, env, cdnURL, zipFilePath, md5Str, string(ds), fvid)
	if _, err = s.fkDao.TxUpFFConfigPublishURL(tx, cdnURL, zipFilePath, md5Str, string(ds), fvid); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("AppFFPublish appkey(%v) env(%v) success update ff publish", appKey, env)
	var tf = &model.TotalFile{
		MCV:      fvids,
		Platform: app.Platform,
		Version:  make(map[string]*model.FileVersion),
	}
	tf.Version["default"] = &model.FileVersion{
		Diffs:   diffs,
		Md5:     md5Str,
		URL:     cdnURL,
		Version: fvids,
	}
	var tfContentByte []byte
	if tfContentByte, err = json.Marshal(tf); err != nil {
		log.Error("%v", err)
		return
	}
	var (
		TotalZipFilePath = make(map[string]string)
		TotalcdnURL      = make(map[string]string)
		mutexFile        sync.Mutex
		mutexBFS         sync.Mutex
	)
	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() (err error) {
		var (
			totalFileName       = appKey
			totalPath, totalURL string
		)
		log.Info("AppFFPublish appkey(%v) env(%v) start write ff_file folder(%v) totalPath(%v) content(%v)",
			appKey, env, folder, totalFileName, string(tfContentByte))
		if totalPath, err = s.fkDao.WriteConfigFile(folder, totalFileName, tfContentByte); err != nil {
			log.Error("%v", err)
			return
		}
		mutexFile.Lock()
		TotalZipFilePath["lengqidong"] = totalPath
		mutexFile.Unlock()
		log.Info("AppFFPublish appkey(%v) env(%v) success write ff file %v", appKey, env, totalPath)
		log.Info("AppFFPublish appkey(%v) env(%v) start up bfs %v", appKey, env, totalPath)
		if totalURL, _, err = s.fkDao.UpBFSV2(path.Join("ff", env), totalPath, appKey); err != nil {
			log.Error("%v", err)
			return
		}
		mutexBFS.Lock()
		TotalcdnURL["lengqidong"] = totalURL
		mutexBFS.Unlock()
		log.Info("AppFFPublish appkey(%v) env(%v) success up bfs %v", appKey, env, totalURL)
		return
	})
	g.Go(func() (err error) {
		var (
			totalFileName       = fmt.Sprintf("%v_%v", appKey, fvids)
			totalPath, totalURL string
		)
		log.Info("AppFFPublish appkey(%v) env(%v) start write ff_file folder(%v) totalPath(%v) content(%v)",
			appKey, env, folder, totalFileName, string(tfContentByte))
		if totalPath, err = s.fkDao.WriteConfigFile(folder, totalFileName, tfContentByte); err != nil {
			log.Error("%v", err)
			return
		}
		mutexFile.Lock()
		TotalZipFilePath["gengxin"] = totalPath
		mutexFile.Unlock()
		log.Info("AppFFPublish appkey(%v) env(%v) success write ff file %v", appKey, env, totalPath)
		log.Info("AppFFPublish appkey(%v) env(%v) start up bfs %v", appKey, env, totalPath)
		if totalURL, _, err = s.fkDao.UpBFSV2(path.Join("ff", env), totalPath, appKey); err != nil {
			log.Error("%v", err)
			return
		}
		mutexBFS.Lock()
		TotalcdnURL["gengxin"] = totalURL
		mutexBFS.Unlock()
		log.Info("AppFFPublish appkey(%v) env(%v) success up bfs %v", appKey, env, totalURL)
		return
	})
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	var tfb, tub []byte
	if tfb, err = json.Marshal(TotalZipFilePath); err != nil {
		log.Error("%v", err)
		return
	}
	if tub, err = json.Marshal(TotalcdnURL); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("AppFFPublish appkey(%v) env(%v) start up total_path(%v) total_url(%v)", appKey, env, string(tfb), string(tub))
	if _, err = s.fkDao.TxUpFFPublishTotal(tx, string(tfb), string(tub), fvid); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("AppFFPublish appkey(%v) env(%v) success up total", appKey, env)
	log.Info("AppFFPublish appkey(%v) env(%v) start get last fvid(%v)", appKey, env, fvid)
	var lastfvid int64
	if lastfvid, err = s.fkDao.AppFFLastFvid(c, appKey, env, fvid); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("AppFFPublish appkey(%v) env(%v) success get last", appKey, env)
	log.Info("AppFFPublish appkey(%v) env(%v) start update state(%v) fvid(%v)", appKey, env, ffmdl.FFPublishHistoryState, lastfvid)
	if _, err = s.fkDao.TxUpFFConfigPublishState(tx, appKey, env, lastfvid, ffmdl.FFPublishHistoryState); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("AppFFPublish appkey(%v) env(%v) success update state", appKey, env)
	log.Info("AppFFPublish appkey(%v) env(%v) start update state(%v) fvid(%v)", appKey, env, ffmdl.FFPublishNowState, fvid)
	if _, err = s.fkDao.TxUpFFConfigPublishState(tx, appKey, env, fvid, ffmdl.FFPublishNowState); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("AppFFPublish appkey(%v) env(%v) success update state", appKey, env)
	log.Info("AppFFPublish appkey(%v) env(%v) success", appKey, env)
	return
}

func (s *Service) formTree(appKey, env string, appChannelMap map[int64]*appmdl.Channel, ffconfig *ffmdl.FF) (trees []*ffmdl.PublishTree) {
	if ffconfig.Version != "" {
		if tree, _ := s.formVersion(appKey, env, ffconfig.Version); tree != nil {
			trees = append(trees, tree)
		}
	}
	if ffconfig.UnVersion != "" {
		if tree := s.formUnVersion(ffconfig.UnVersion); tree != nil {
			trees = append(trees, tree)
		}
	}
	if ffconfig.RomVersion != "" {
		trees = append(trees, &ffmdl.PublishTree{
			OP:    "in",
			Prop:  "sv",
			Value: ffconfig.RomVersion,
		})
	}
	if ffconfig.Channel != "" {
		if tree, _ := s.formChannel(ffconfig.Channel, appChannelMap); tree != nil {
			trees = append(trees, tree)
		}
	}
	if ffconfig.Brand != "" {
		trees = append(trees, &ffmdl.PublishTree{
			OP:    "in",
			Prop:  "brand",
			Value: ffconfig.Brand,
		})
	}
	if ffconfig.UnBrand != "" {
		trees = append(trees, &ffmdl.PublishTree{
			OP:    "ex",
			Prop:  "brand",
			Value: ffconfig.UnBrand,
		})
	}
	if ffconfig.Network != "" {
		trees = append(trees, &ffmdl.PublishTree{
			OP:    "in",
			Prop:  "nt",
			Value: ffconfig.Network,
		})
	}
	if ffconfig.ISP != "" {
		trees = append(trees, &ffmdl.PublishTree{
			OP:    "in",
			Prop:  "ot",
			Value: ffconfig.ISP,
		})
	}
	return
}

func (s *Service) formVersion(appKey, env, ffconfig string) (tree *ffmdl.PublishTree, err error) {
	var ffversion *ffmdl.Version
	if err = json.Unmarshal([]byte(ffconfig), &ffversion); err != nil {
		log.Error("%v", err)
	} else {
		var (
			op    string
			value []string
		)
		min := ffversion.Min
		max := ffversion.Max
		if min == 0 {
			op = "lt"
			value = append(value, strconv.FormatInt(max, 10))
		} else if max == 0 {
			op = "gt"
			value = append(value, strconv.FormatInt(min, 10))
		} else {
			if min == max {
				op = "eq"
				value = append(value, strconv.FormatInt(min, 10))
			} else if min < max {
				op = "in"
				value = append(value, strconv.FormatInt(min, 10))
				value = append(value, strconv.FormatInt(max, 10))
			} else {
				err = fmt.Errorf("appKey(%v) env(%v) min(%v) > max(%v)", appKey, env, min, max)
				log.Error("%v", err)
			}
		}
		if op != "" && len(value) > 0 {
			var sp = ","
			if op == "in" {
				sp = "~"
			}
			tree = &ffmdl.PublishTree{
				OP:    op,
				Prop:  "av",
				Value: strings.Join(value, sp),
			}
		}
	}
	return
}

func (s *Service) formUnVersion(ffconfig string) (tree *ffmdl.PublishTree) {
	var vs []string
	if vs = strings.Split(ffconfig, ","); len(vs) > 0 {
		tree = &ffmdl.PublishTree{
			OP:    "ex",
			Prop:  "av",
			Value: ffconfig,
		}
	}
	return
}

func (s *Service) formChannel(ffconfig string, channels map[int64]*appmdl.Channel) (tree *ffmdl.PublishTree, err error) {
	var (
		chIDs []int64
		value []string
	)
	if chIDs, err = xstr.SplitInts(ffconfig); err != nil {
		log.Error("%v", err)
	} else {
		for _, chID := range chIDs {
			if appChannel, ok := channels[chID]; ok {
				value = append(value, appChannel.Code)
			}
		}
	}
	if len(value) > 0 {
		tree = &ffmdl.PublishTree{
			OP:    "in",
			Prop:  "ch",
			Value: strings.Join(value, ","),
		}
	}
	return
}

func (s *Service) formTree2(t *ffmdl.PublishTree, ts []*ffmdl.PublishTree, count int) {
	log.Info("form tree t(%+v) ts(%+v) count(%v)", t, ts, count)
	k := len(ts) - count
	*t = *ts[k]
	count--
	if count == 0 {
		return
	}
	t.Son = &ffmdl.PublishTree{}
	s.formTree2(t.Son, ts, count)
}

// AppFFHistory get app ff publish history.
func (s *Service) AppFFHistory(c context.Context, appKey, env string, pn, ps int) (res *ffmdl.HistoryResult, err error) {
	var (
		historys []*ffmdl.ConfigPublish
		total    int
	)
	if total, err = s.fkDao.AppFFHistoryCount(c, appKey, env); err != nil {
		log.Error("%v", err)
		return
	}
	if historys, err = s.fkDao.AppFFHistory(c, appKey, env, pn, ps); err != nil {
		log.Error("%v", err)
		return
	}
	res = &ffmdl.HistoryResult{
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    pn,
			Ps:    ps,
		},
		Items: historys,
	}
	return
}

// AppFFHistoryByID get app ff publish history by id
func (s *Service) AppFFHistoryByID(c context.Context, appKey string, ffid int64) (res *ffmdl.ConfigPublish, err error) {
	if res, err = s.fkDao.AppFFHistoryByID(c, appKey, ffid); err != nil {
		log.Error("%v", err)
	}
	return
}

// AppFFDiff get history diff.
// nolint:gocognit
func (s *Service) AppFFDiff(c context.Context, appKey, env string, fvid int64) (res []*ffmdl.Diff, err error) {
	var news []*ffmdl.File
	if news, err = s.fkDao.AppFFFile(c, appKey, env, fvid); err != nil {
		log.Error("%v", err)
		return
	}
	var lastfvid int64
	if lastfvid, err = s.fkDao.AppFFLastFvid(c, appKey, env, fvid); err != nil {
		log.Error("%v", err)
		return
	}
	var origins []*ffmdl.File
	if lastfvid > 0 {
		if origins, err = s.fkDao.AppFFFile(c, appKey, env, lastfvid); err != nil {
			log.Error("%v", err)
			return
		}
	}
	for _, new := range news {
		if new.State == ffmdl.FFStatePublish {
			continue
		}
		re := &ffmdl.Diff{
			Key:      new.Key,
			Desc:     new.Desc,
			State:    new.State,
			Operator: new.Operator,
			MTime:    new.MTime,
		}
		if new.State != ffmdl.FFStatDel {
			re.New = &ffmdl.DiffItem{
				Status:      new.Status,
				Salt:        new.Salt,
				Bucket:      new.Bucket,
				BucketCount: new.BucketCount,
				Whith:       new.Whith,
				BlackMid:    new.BlackMid,
				UnVersion:   new.UnVersion,
				RomVersion:  new.RomVersion,
				Brand:       new.Brand,
				UnBrand:     new.UnBrand,
				Network:     new.Network,
				ISP:         new.ISP,
				Channel:     new.Channel,
			}
			if new.Version != "" {
				var v *ffmdl.Version
				if err = json.Unmarshal([]byte(new.Version), &v); err != nil {
					log.Error("%v", err)
					err = nil
				}
				re.New.Version = v
			}
			if new.BlackList != "" {
				var bls []*ffmdl.FF
				if err = json.Unmarshal([]byte(new.BlackList), &bls); err != nil {
					log.Error("%v", err)
					err = nil
				}
				for _, bl := range bls {
					bi := &ffmdl.DiffItem{
						RomVersion: bl.RomVersion,
						Network:    bl.Network,
						Brand:      bl.Brand,
					}
					if bl.Version != "" {
						var bv *ffmdl.Version
						if err = json.Unmarshal([]byte(bl.Version), &bv); err != nil {
							log.Error("%v", err)
							err = nil
						}
						bi.Version = bv
					}
					re.New.BlackList = append(re.New.BlackList, bi)
				}
			}
		}
		for _, origin := range origins {
			if new.Key == origin.Key {
				if origin.State != ffmdl.FFStatDel {
					re.Origin = &ffmdl.DiffItem{
						Status:      origin.Status,
						Salt:        origin.Salt,
						Bucket:      origin.Bucket,
						BucketCount: origin.BucketCount,
						Whith:       origin.Whith,
						BlackMid:    origin.BlackMid,
						UnVersion:   origin.UnVersion,
						RomVersion:  origin.RomVersion,
						Brand:       origin.Brand,
						UnBrand:     origin.UnBrand,
						Network:     origin.Network,
						ISP:         origin.ISP,
						Channel:     origin.Channel,
					}
					if origin.Version != "" {
						var v *ffmdl.Version
						if err = json.Unmarshal([]byte(origin.Version), &v); err != nil {
							log.Error("%v", err)
							err = nil
						}
						re.Origin.Version = v
					}
					if origin.BlackList != "" {
						var bls []*ffmdl.FF
						if err = json.Unmarshal([]byte(origin.BlackList), &bls); err != nil {
							log.Error("%v", err)
							err = nil
						}
						for _, bl := range bls {
							bi := &ffmdl.DiffItem{
								RomVersion: bl.RomVersion,
								Network:    bl.Network,
								Brand:      bl.Brand,
							}
							if bl.Version != "" {
								var bv *ffmdl.Version
								if err = json.Unmarshal([]byte(bl.Version), &bv); err != nil {
									log.Error("%v", err)
									err = nil
								}
								bi.Version = bv
							}
							re.Origin.BlackList = append(re.Origin.BlackList, bi)
						}
					}
				}
			}
		}
		res = append(res, re)
	}
	return
}

// AppFFPublishDiff app ff publish diff.
// nolint:gocognit
func (s *Service) AppFFPublishDiff(c context.Context, appKey, env string) (res []*ffmdl.Diff, err error) {
	var news []*ffmdl.FF
	if news, err = s.fkDao.AppFFConfigs(c, appKey, env); err != nil {
		log.Error("%v", err)
		return
	}
	var lastfvid int64
	if lastfvid, err = s.fkDao.AppFFLastFvid(c, appKey, env, 0); err != nil {
		log.Error("%v", err)
		return
	}
	var origins []*ffmdl.File
	if lastfvid > 0 {
		if origins, err = s.fkDao.AppFFFile(c, appKey, env, lastfvid); err != nil {
			log.Error("%v", err)
			return
		}
	}
	for _, new := range news {
		if new.State == ffmdl.FFStatePublish {
			continue
		}
		re := &ffmdl.Diff{
			Key:      new.Key,
			Desc:     new.Desc,
			State:    new.State,
			Operator: new.Operator,
			MTime:    new.MTime,
		}
		if new.State != ffmdl.FFStatDel {
			re.New = &ffmdl.DiffItem{
				Status:      new.Status,
				Salt:        new.Salt,
				Bucket:      new.Bucket,
				BucketCount: new.BucketCount,
				Whith:       new.Whith,
				BlackMid:    new.BlackMid,
				UnVersion:   new.UnVersion,
				RomVersion:  new.RomVersion,
				Brand:       new.Brand,
				UnBrand:     new.UnBrand,
				Network:     new.Network,
				ISP:         new.ISP,
				Channel:     new.Channel,
			}
			if new.Version != "" {
				var v *ffmdl.Version
				if err = json.Unmarshal([]byte(new.Version), &v); err != nil {
					log.Error("%v", err)
					err = nil
				}
				re.New.Version = v
			}
			if new.BlackList != "" {
				var bls []*ffmdl.FF
				if err = json.Unmarshal([]byte(new.BlackList), &bls); err != nil {
					log.Error("%v", err)
					err = nil
				}
				for _, bl := range bls {
					bi := &ffmdl.DiffItem{
						RomVersion: bl.RomVersion,
						Network:    bl.Network,
						Brand:      bl.Brand,
					}
					if bl.Version != "" {
						var bv *ffmdl.Version
						if err = json.Unmarshal([]byte(bl.Version), &bv); err != nil {
							log.Error("%v", err)
							err = nil
						}
						bi.Version = bv
					}
					re.New.BlackList = append(re.New.BlackList, bi)
				}
			}
		}
		for _, origin := range origins {
			if new.Key == origin.Key {
				if origin.State != ffmdl.FFStatDel {
					re.Origin = &ffmdl.DiffItem{
						Status:      new.Status,
						Salt:        origin.Salt,
						Bucket:      origin.Bucket,
						BucketCount: origin.BucketCount,
						Whith:       origin.Whith,
						BlackMid:    origin.BlackMid,
						UnVersion:   origin.UnVersion,
						RomVersion:  origin.RomVersion,
						Brand:       origin.Brand,
						UnBrand:     origin.UnBrand,
						Network:     origin.Network,
						ISP:         origin.ISP,
						Channel:     origin.Channel,
					}
					if origin.Version != "" {
						var v *ffmdl.Version
						if err = json.Unmarshal([]byte(origin.Version), &v); err != nil {
							log.Error("%v", err)
							err = nil
						}
						re.Origin.Version = v
					}
					if origin.BlackList != "" {
						var bls []*ffmdl.FF
						if err = json.Unmarshal([]byte(origin.BlackList), &bls); err != nil {
							log.Error("%v", err)
							err = nil
						}
						for _, bl := range bls {
							bi := &ffmdl.DiffItem{
								RomVersion: bl.RomVersion,
								Network:    bl.Network,
								Brand:      bl.Brand,
							}
							if bl.Version != "" {
								var bv *ffmdl.Version
								if err = json.Unmarshal([]byte(bl.Version), &bv); err != nil {
									log.Error("%v", err)
									err = nil
								}
								bi.Version = bv
							}
							re.Origin.BlackList = append(re.Origin.BlackList, bi)
						}
					}
				}
			}
		}
		res = append(res, re)
	}
	return
}

// AppFFConfig get ff config.
func (s *Service) AppFFConfig(c context.Context, appKey, env, key string) (res *ffmdl.FF, err error) {
	if res, err = s.fkDao.AppFFConfig(c, appKey, env, key); err != nil {
		log.Error("%v", err)
		return
	}
	if res != nil && res.State < 0 {
		res = nil
	}
	return
}

// AppFFConfigDel del ff config.
func (s *Service) AppFFConfigDel(c context.Context, appKey, env, key string) (err error) {
	var tx *xsql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	var ff *ffmdl.FF
	if ff, err = s.fkDao.AppFFConfig(c, appKey, env, key); err != nil {
		log.Error("%v", err)
		return
	}
	if ff.State == ffmdl.FFStatAdd {
		if _, err = s.fkDao.TxDelFFConfig2(tx, appKey, env, key); err != nil {
			log.Error("%v", err)
		}
	} else {
		if _, err = s.fkDao.TxDelFFConfig(tx, appKey, env, key); err != nil {
			log.Error("%v", err)
		}
	}
	return
}

func (s *Service) AppFFModifyCount(c context.Context, appKey string) (res *ffmdl.ModifyCount, err error) {
	modifyItem := &ffmdl.ModifyCount{Test: 0, Prod: 0}
	if modifyItem.Test, err = s.fkDao.FFModiyCount(c, appKey, "test"); err != nil {
		log.Error("%v", err)
		return
	}
	if modifyItem.Prod, err = s.fkDao.FFModiyCount(c, appKey, "prod"); err != nil {
		log.Error("%v", err)
		return
	}
	res = modifyItem
	return
}
